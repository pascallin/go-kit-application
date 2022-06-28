package usersvc

import (
	"fmt"
	"net"

	"github.com/go-kit/kit/log"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/pascallin/go-kit-application/config"
	pb "github.com/pascallin/go-kit-application/pb/usersvc"
	"github.com/pascallin/go-kit-application/pkg"
	"github.com/pascallin/go-kit-application/usersvc/endpoints"
	"github.com/pascallin/go-kit-application/usersvc/services"
	"github.com/pascallin/go-kit-application/usersvc/transports"
)

func GrpcServe(logger log.Logger) error {
	c := config.GetUserSvcConfig()

	zipkinTracer, tracer, err := pkg.InitTracer(c.Name)
	if err != nil {
		panic(err)
	}

	var (
		service    = services.NewService(logger)
		endpoints  = endpoints.New(service, logger, tracer, zipkinTracer)
		grpcServer = transports.NewGRPCServer(endpoints, logger)
	)

	grpcAddr := fmt.Sprintf(":%d", c.GrpcPort)
	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Log("transport", "gRPC", "during", "Listen", "err", err)
		panic(err)
	}

	logger.Log("transport", "gRPC", "addr", c.GrpcPort)
	server := grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor))

	pb.RegisterUserServer(server, grpcServer)
	// heath check register
	grpc_health_v1.RegisterHealthServer(server, NewHealthChecker())

	return server.Serve(grpcListener)
}
