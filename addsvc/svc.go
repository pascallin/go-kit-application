package addsvc

import (
	"fmt"
	"net"

	"github.com/go-kit/kit/log"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/pascallin/go-kit-application/addsvc/endpoints"
	"github.com/pascallin/go-kit-application/addsvc/services"
	"github.com/pascallin/go-kit-application/addsvc/transports"
	"github.com/pascallin/go-kit-application/config"
	pb "github.com/pascallin/go-kit-application/pb/addsvc"
	"github.com/pascallin/go-kit-application/pkg"
)

func GrpcServe(logger log.Logger) error {
	c := config.GetAddSvcConfig()

	zipkinTracer, tracer, err := pkg.InitTracer(c.Name)
	if err != nil {
		panic(err)
	}

	ints, chars := services.InitMetrics()
	duration := endpoints.InitMetrics()
	var (
		service    = services.NewService(logger, ints, chars)
		endpoints  = endpoints.NewEndpoints(service, logger, duration, tracer, zipkinTracer)
		grpcServer = transports.NewGRPCServer(endpoints, tracer, zipkinTracer, logger)
	)

	// The gRPC listener mounts the Go kit gRPC server we created.
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", c.GrpcPort))
	if err != nil {
		logger.Log("transport", "gRPC", "during", "Listen", "err", err)
		return err
	}

	logger.Log("transport", "gRPC", "addr", c.GrpcPort)
	baseServer := grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor))
	// register service
	pb.RegisterAddServer(baseServer, grpcServer)
	// heath check register
	grpc_health_v1.RegisterHealthServer(baseServer, NewHealthChecker())

	return baseServer.Serve(grpcListener)
}
