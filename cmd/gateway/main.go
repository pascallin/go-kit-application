package main

import (
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/gorilla/mux"
	"github.com/hashicorp/consul/api"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/lb"

	"github.com/pascallin/go-micro-services/internal/strsvc"
	"github.com/pascallin/go-micro-services/internal/strsvc/strtransport"
)

func main() {
	var (
		httpAddr     = flag.String("http.addr", ":8000", "Address for HTTP (JSON) server")
		consulAddr   = flag.String("consul.addr", "localhost:8400", "Consul agent address")
		retryMax     = flag.Int("retry.max", 3, "per-request retries to different instances")
		retryTimeout = flag.Duration("retry.timeout", 500*time.Millisecond, "per-request timeout, including retries")
	)
	flag.Parse()

	// Logging domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	// Service discovery domain. In this example we use Consul.
	var client consulsd.Client
	{
		consulConfig := api.DefaultConfig()
		if len(*consulAddr) > 0 {
			consulConfig.Address = *consulAddr
		}
		consulClient, err := api.NewClient(consulConfig)
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
		client = consulsd.NewClient(consulClient)
	}

	// Transport domain.
	tracer := stdopentracing.GlobalTracer() // no-op
	zipkinTracer, _ := stdzipkin.NewTracer(nil, stdzipkin.WithNoopTracer(true))
	//ctx := context.Background()
	r := mux.NewRouter()

	// Now we begin installing the routes. Each route corresponds to a single
	// method: sum, concat, uppercase, and count.

	// strsvc routes.
	{
		var (
			tags        = []string{}
			passingOnly = true
			endpoints   = stringsvc.Set{}
			instancer   = consulsd.NewInstancer(client, logger, "strsvc", tags, passingOnly)
		)
		{
			factory := stringsvcFactory(stringsvc.MakeUppercaseEndpoint, tracer, zipkinTracer, logger)
			endpointer := sd.NewEndpointer(instancer, factory, logger)
			balancer := lb.NewRoundRobin(endpointer)
			retry := lb.Retry(*retryMax, *retryTimeout, balancer)
			endpoints.UppercaseEndpoint = retry
		}
		r.PathPrefix("/string").Handler(http.StripPrefix("/string", strtransport.NewHTTPHandler(endpoints, tracer, zipkinTracer, logger)))
	}

	// Interrupt handler.
	errc := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// HTTP strtransport.
	go func() {
		logger.Log("strtransport", "HTTP", "addr", *httpAddr)
		errc <- http.ListenAndServe(*httpAddr, r)
	}()

	// Run!
	logger.Log("exit", <-errc)
}

func stringsvcFactory(makeEndpoint func(service stringsvc.StringService) endpoint.Endpoint, tracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, logger log.Logger) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		conn, err := grpc.Dial(instance, grpc.WithInsecure())
		if err != nil {
			return nil, nil, err
		}
		service := strtransport.NewGRPCClient(conn, tracer, zipkinTracer, logger)
		endpoint := makeEndpoint(service)

		return endpoint, conn, nil
	}
}