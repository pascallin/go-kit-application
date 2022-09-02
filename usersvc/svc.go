package usersvc

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/go-kit/kit/log"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/pascallin/go-kit-application/config"
	"github.com/pascallin/go-kit-application/conn"
	pb "github.com/pascallin/go-kit-application/pb/usersvc"
	"github.com/pascallin/go-kit-application/pkg"
	"github.com/pascallin/go-kit-application/usersvc/endpoints"
	"github.com/pascallin/go-kit-application/usersvc/services"
	"github.com/pascallin/go-kit-application/usersvc/transports"
)

func GrpcServe(logger log.Logger) error {
	c := config.GetUserSvcConfig()
	db, err := conn.GetMongo(context.Background())
	if err != nil {
		return err
	}

	zipkinTracer, tracer, err := pkg.InitTracer(c.Name)
	if err != nil {
		return err
	}

	service, err := services.InitializeService(db.DB, logger)
	if err != nil {
		return err
	}
	endpoints := endpoints.New(service, logger, tracer, zipkinTracer)
	grpcServer := transports.NewGRPCServer(endpoints, logger)

	grpcAddr := fmt.Sprintf(":%d", c.GrpcPort)
	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Log("transport", "gRPC", "during", "Listen", "err", err)
		return err
	}

	logger.Log("transport", "gRPC", "addr", c.GrpcPort)
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			kitgrpc.Interceptor,
			grpc_recovery.UnaryServerInterceptor(),
		)),
	)

	pb.RegisterUserServer(server, grpcServer)
	// heath check register
	grpc_health_v1.RegisterHealthServer(server, NewGrpHealthChecker())

	return server.Serve(grpcListener)
}

// @title user service
// @version 1.0
// @description  user service
// @securityDefinitions.apikey  ServiceApiKey
// @in                          header
// @name                        x-api-key
func HttpServe(logger log.Logger) error {
	c := config.GetUserSvcConfig()
	db, err := conn.GetMongo(context.Background())
	if err != nil {
		return err
	}

	service, err := services.InitializeService(db.DB, logger)
	if err != nil {
		return err
	}
	httpHandler := transports.MakeHandler(service, logger)

	// The HTTP listener mounts the Go kit HTTP handler we created.
	httpListener, err := net.Listen("tcp", fmt.Sprintf(":%d", c.HttpPort))
	if err != nil {
		logger.Log("transport", "HTTP", "during", "Listen", "err", err)
		return err
	}

	logger.Log("transport", "HTTP", "addr", fmt.Sprintf(":%d", c.HttpPort))
	return http.Serve(httpListener, httpHandler)
}
