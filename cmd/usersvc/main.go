package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/log"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/joho/godotenv"
	"github.com/oklog/oklog/pkg/group"
	"google.golang.org/grpc"

	"github.com/pascallin/go-micro-services/internal/pkg/db"
	"github.com/pascallin/go-micro-services/internal/usersvc"
	"github.com/pascallin/go-micro-services/internal/usersvc/transports"
	"github.com/pascallin/go-micro-services/pb"
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
		// zipkinURL    = os.Getenv("DEFAULT_ZIPKIN_URL")
		// zipkinBridge = false
	)

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var (
		endpoints  = usersvc.New()
		grpcServer = transports.NewGRPCServer(endpoints, logger)
	)

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
