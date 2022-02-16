package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/go-kit/kit/log"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/joho/godotenv"
	"github.com/oklog/oklog/pkg/group"
	stdopentracing "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	zipkin "github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/pascallin/go-kit-application/conn"
	"github.com/pascallin/go-kit-application/discovery"
	"github.com/pascallin/go-kit-application/pb"
	"github.com/pascallin/go-kit-application/usersvc"
	"github.com/pascallin/go-kit-application/usersvc/endpoints"
	"github.com/pascallin/go-kit-application/usersvc/transports"
)

func main() {
	godotenv.Load()

	// connect mongodb
	mongo, err := db.NewMongoDatabase()
	if err != nil {
		panic(err)
	}
	defer mongo.Close()

	var (
		grpcAddr = ":" + os.Getenv("USER_SVC_RPC_PORT")
		grpcPort = os.Getenv("USER_SVC_RPC_PORT")
		host     = os.Getenv("SERVICE_HOST")
		instance = os.Getenv("SERVICE_HOSTNAME")
		svcName  = os.Getenv("SERVICE_NAME")
	)

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var (
		zipkinURL    = os.Getenv("DEFAULT_ZIPKIN_URL")
		zipkinBridge = true
	)
	var zipkinTracer *zipkin.Tracer
	{
		if zipkinURL != "" {
			var (
				err         error
				hostPort    = "localhost:80"
				serviceName = svcName
				reporter    = zipkinhttp.NewReporter(zipkinURL)
			)
			defer reporter.Close()
			zEP, _ := zipkin.NewEndpoint(serviceName, hostPort)
			zipkinTracer, err = zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(zEP))
			if err != nil {
				logger.Log("err", err)
				os.Exit(1)
			}
			if !(zipkinBridge) {
				logger.Log("tracer", "Zipkin", "type", "Native", "URL", zipkinURL)
			}
		}
	}
	// Determine which OpenTracing tracer to use. We'll pass the tracer to all the
	// components that use it, as a dependency.
	var tracer stdopentracing.Tracer
	{
		if zipkinBridge && zipkinTracer != nil {
			logger.Log("tracer", "Zipkin", "type", "OpenTracing", "URL", zipkinURL)
			tracer = zipkinot.Wrap(zipkinTracer)
			zipkinTracer = nil // do not instrument with both native tracer and opentracing bridge
		} else {
			tracer = stdopentracing.GlobalTracer() // no-op
		}
	}

	var (
		endpoints  = endpoints.New(logger, tracer, zipkinTracer)
		grpcServer = transports.NewGRPCServer(endpoints, logger)
	)

	client, err := discovery.NewKitDiscoverClient()
	if err != nil {
		panic(err)
	}

	var g group.Group
	{
		grpcListener, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			logger.Log("transport", "gRPC", "during", "Listen", "err", err)
			os.Exit(1)
		}
		g.Add(func() error {
			logger.Log("transport", "gRPC", "addr", grpcAddr)
			baseServer := grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor))

			pb.RegisterUserServer(baseServer, grpcServer)
			// heath check register
			grpc_health_v1.RegisterHealthServer(baseServer, usersvc.NewHealthChecker())

			client, err := discovery.NewKitDiscoverClient()
			if err != nil {
				panic(err)
			}
			port, err := strconv.Atoi(grpcPort)
			if err != nil {
				panic(err)
			}
			status := client.Register(svcName, discovery.ServiceInstance{InstanceId: instance, InstanceHost: host, InstancePort: port}, make(map[string]string))
			logger.Log("consul discovery register ", status)

			return baseServer.Serve(grpcListener)
		}, func(error) {
			grpcListener.Close()
		})
	}

	{
		cancelInterrupt := make(chan struct{})
		g.Add(func() error {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			select {
			case sig := <-c:
				client.DeRegister(instance)
				return fmt.Errorf("received signal %s", sig)
			case <-cancelInterrupt:
				return nil
			}
		}, func(error) {
			close(cancelInterrupt)
		})
	}

	logger.Log("exit", g.Run())
}
