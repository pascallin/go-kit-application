package main

import (
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/go-kit/log"
	"github.com/joho/godotenv"
	"github.com/oklog/oklog/pkg/group"
	"github.com/pascallin/go-kit-application/addsvc"
	addtransport "github.com/pascallin/go-kit-application/addsvc/transports"
	"github.com/pascallin/go-kit-application/discovery"

	"github.com/pascallin/go-kit-application/pb"
	"github.com/pascallin/go-kit-application/tracing"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	godotenv.Load()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}
	// Create the (sparse) metrics we'll use in the service. They, too, are
	// dependencies that we pass to components that use them.
	var ints, chars metrics.Counter
	{
		// Business-level metrics.
		ints = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "example",
			Subsystem: "addsvc",
			Name:      "integers_summed",
			Help:      "Total count of integers summed via the Sum method.",
		}, []string{})
		chars = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "example",
			Subsystem: "addsvc",
			Name:      "characters_concatenated",
			Help:      "Total count of characters concatenated via the Concat method.",
		}, []string{})
	}
	var duration metrics.Histogram
	{
		// Endpoint-level metrics.
		duration = prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "example",
			Subsystem: "addsvc",
			Name:      "request_duration_seconds",
			Help:      "Request duration in seconds.",
		}, []string{"method", "success"})
	}
	http.DefaultServeMux.Handle("/metrics", promhttp.Handler())

	zipkinTracer, octracer, err := tracing.NewOpentracingTracer("localhost:80", "addsvc")
	if err != nil {
		panic(err)
	}

	var (
		service    = addsvc.NewService(logger, ints, chars)
		endpoints  = addsvc.NewEndpoints(service, logger, duration, octracer, zipkinTracer)
		grpcServer = addtransport.NewGRPCServer(endpoints, octracer, zipkinTracer, logger)
	)

	var (
		grpcAddr = ":" + os.Getenv("ADD_SVC_GRPC_PORT")
	)
	var g group.Group
	{
		// The gRPC listener mounts the Go kit gRPC server we created.
		grpcListener, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			logger.Log("transport", "gRPC", "during", "Listen", "err", err)
			os.Exit(1)
		}
		g.Add(func() error {
			logger.Log("transport", "gRPC", "addr", grpcAddr)
			baseServer := grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor))
			// register service
			pb.RegisterAddServer(baseServer, grpcServer)
			// heath check register
			grpc_health_v1.RegisterHealthServer(baseServer, addsvc.NewHealthChecker())

			client, err := discovery.NewKitDiscoverClient()
			if err != nil {
				panic(err)
			}
			port, err := strconv.Atoi(os.Getenv("ADD_SVC_GRPC_PORT"))
			if err != nil {
				panic(err)
			}
			status := client.Register("addsvc", discovery.ServiceInstance{InstanceId: "addsvc", InstanceHost: os.Getenv("SERVICE_HOST"), InstancePort: port}, make(map[string]string))
			logger.Log("consul discovery register ", status)

			return baseServer.Serve(grpcListener)
		}, func(error) {
			grpcListener.Close()
		})
	}
	logger.Log("exit", g.Run())
}
