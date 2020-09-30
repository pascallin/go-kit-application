package addsvc

import (
	"flag"
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/pascallin/go-micro-services/common"
	addendpoint2 "github.com/pascallin/go-micro-services/internal/addsvc/addendpoint"
	addservice2 "github.com/pascallin/go-micro-services/internal/addsvc/addservice"
	addtransport2 "github.com/pascallin/go-micro-services/internal/addsvc/addtransport"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/lightstep/lightstep-tracer-go"
	"github.com/oklog/oklog/pkg/group"
	stdopentracing "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"sourcegraph.com/sourcegraph/appdash"
	appdashot "sourcegraph.com/sourcegraph/appdash/opentracing"

	addpb "github.com/pascallin/go-micro-services/pb"
)

func StartAddSVCService() {
	var (
		debugAddr      = flag.String("debug.addr", ":8081", "Debug and metrics listen address")
		httpAddr	= flag.String("http-addr", ":8082", "HTTP listen address")
		grpcAddr       = flag.String("grpc-addr", ":8083", "gRPC listen address")
		zipkinURL      = flag.String("zipkin-url", "http://localhost:9411/api/v2/spans", "Enable Zipkin tracing via HTTP reporter URL e.g. http://localhost:9411/api/v2/spans")
		zipkinBridge   = flag.Bool("zipkin-ot-bridge", false, "Use Zipkin OpenTracing bridge instead of native implementation")
		lightstepToken = flag.String("lightstep-token", "", "Enable LightStep tracing via a LightStep access token")
		appdashAddr    = flag.String("appdash-addr", "", "Enable Appdash tracing via an Appdash server host:port")
	)

	flag.Parse()

	// Create a single logger, which we'll use and give to other components.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var zipkinTracer *zipkin.Tracer
	{
		if *zipkinURL != "" {
			var (
				err         error
				hostPort    = "localhost:80"
				serviceName = "addsvc"
				reporter    = zipkinhttp.NewReporter(*zipkinURL)
			)
			defer reporter.Close()
			zEP, _ := zipkin.NewEndpoint(serviceName, hostPort)
			zipkinTracer, err = zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(zEP))
			if err != nil {
				logger.Log("err", err)
				os.Exit(1)
			}
			if !(*zipkinBridge) {
				logger.Log("tracer", "Zipkin", "type", "Native", "URL", *zipkinURL)
			}
		}
	}

	// Determine which OpenTracing tracer to use. We'll pass the tracer to all the
	// components that use it, as a dependency.
	var tracer stdopentracing.Tracer
	{
		if *zipkinBridge && zipkinTracer != nil {
			logger.Log("tracer", "Zipkin", "type", "OpenTracing", "URL", *zipkinURL)
			tracer = zipkinot.Wrap(zipkinTracer)
			zipkinTracer = nil // do not instrument with both native tracer and opentracing bridge
		} else if *lightstepToken != "" {
			logger.Log("tracer", "LightStep") // probably don't want to print out the token :)
			tracer = lightstep.NewTracer(lightstep.Options{
				AccessToken: *lightstepToken,
			})
			defer lightstep.FlushLightStepTracer(tracer)
		} else if *appdashAddr != "" {
			logger.Log("tracer", "Appdash", "addr", *appdashAddr)
			tracer = appdashot.NewTracer(appdash.NewRemoteCollector(*appdashAddr))
		} else {
			tracer = stdopentracing.GlobalTracer() // no-op
		}
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

	var (
		service		= addservice2.New(logger, ints, chars)
		endpoints	= addendpoint2.New(service, logger, duration, tracer, zipkinTracer)
		httpHandler	= addtransport2.NewHTTPHandler(endpoints, logger, tracer, zipkinTracer)
		grpcServer  = addtransport2.NewGRPCServer(endpoints, tracer, zipkinTracer, logger)
	)

	var g group.Group

	{
		// The debug listener mounts the http.DefaultServeMux, and serves up
		// stuff like the Prometheus metrics route, the Go debug and profiling
		// routes, and so on.
		debugListener, err := net.Listen("tcp", *debugAddr)
		if err != nil {
			logger.Log("transport", "debug/HTTP", "during", "Listen", "err", err)
			os.Exit(1)
		}
		g.Add(func() error {
			logger.Log("transport", "debug/HTTP", "addr", *debugAddr)
			return http.Serve(debugListener, http.DefaultServeMux)
		}, func(error) {
			debugListener.Close()
		})
	}

	{
		// The HTTP listener mounts the Go kit HTTP handler we created.
		httpListener, err := net.Listen("tcp", *httpAddr)
		if err != nil {
			logger.Log("transport", "HTTP", "during", "Listen", "err", err)
			os.Exit(1)
		}
		g.Add(func() error {
			logger.Log("transport", "HTTP", "addr", *httpAddr)
			return http.Serve(httpListener, httpHandler)
		}, func(error) {
			httpListener.Close()
		})
	}
	{
		// The gRPC listener mounts the Go kit gRPC server we created.
		grpcListener, err := net.Listen("tcp", *grpcAddr)
		if err != nil {
			logger.Log("transport", "gRPC", "during", "Listen", "err", err)
			os.Exit(1)
		}
		g.Add(func() error {
			logger.Log("transport", "gRPC", "addr", *grpcAddr)
			// we add the Go Kit gRPC Interceptor to our gRPC service as it is used by
			// the here demonstrated zipkin tracing middleware.
			baseServer := grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor))
			addpb.RegisterAddServer(baseServer, grpcServer)
			return baseServer.Serve(grpcListener)
		}, func(error) {
			grpcListener.Close()
		})
	}

	ctrl, err := common.ConnConsul("http://localhost:8500")
	if err != nil {
		fmt.Errorf("register error")
		return
	}

	{
		// This function just sits and waits for ctrl-C.
		cancelInterrupt := make(chan struct{})
		g.Add(func() error {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			select {
			case sig := <-c:
				ctrl.UnRegister("addsvc")
				ctrl.UnRegister("addsvc_grpc")
				return fmt.Errorf("received signal %s", sig)
			case <-cancelInterrupt:
				return nil
			}
		}, func(error) {
			close(cancelInterrupt)
		})
	}

	ctrl.Register(&api.AgentServiceRegistration{
		Kind:              "HTTP",
		ID:                "addsvc",
		Name:              "addsvc",
		Tags:              []string{},
		Port:              8082,
		Address:           "127.0.0.1",
		EnableTagOverride: false,
		Meta:              map[string]string{},
		Weights: &api.AgentWeights{
			Passing: 10,
			Warning: 1,
		},
		//Check:             &api.AgentServiceCheck{
		//	Interval:                       "10s",
		//	Timeout:                        "5s",
		//	HTTP:                           "http://192.168.10.106:666/health",
		//	Method:                         "GET",
		//},})
	})
	ctrl.Register(&api.AgentServiceRegistration{
		Kind:              "GRPC",
		ID:                "addsvc_grpc",
		Name:              "addsvc_grpc",
		Tags:              []string{},
		Port:              8083,
		Address:           "127.0.0.1",
		EnableTagOverride: false,
		Meta:              map[string]string{},
		Weights: &api.AgentWeights{
			Passing: 10,
			Warning: 1,
		},
		//Check:             &api.AgentServiceCheck{
		//	Interval:                       "10s",
		//	Timeout:                        "5s",
		//	HTTP:                           "http://192.168.10.106:666/health",
		//	Method:                         "GET",
		//},})
	})

	logger.Log("exit", g.Run())
}
