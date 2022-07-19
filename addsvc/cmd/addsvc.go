package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/pascallin/go-kit-application/addsvc"
	"github.com/pascallin/go-kit-application/config"
	"github.com/pascallin/go-kit-application/pkg"
)

func main() {
	logger := pkg.GetLogger()
	c := config.GetAddSvcConfig()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if c.IsNeedDiscovery {
		client, err := pkg.NewKitDiscoverClient()
		if err != nil {
			panic(err)
		}
		status := client.Register(c.Name, pkg.ServiceInstance{
			InstanceId:   c.HostName,
			InstanceHost: c.Host,
			InstancePort: c.GrpcPort,
		}, make(map[string]string))
		logger.Log("consul discovery register ", status)
		defer client.Deregister(c.Name)
	}

	go func() {
		if err := pkg.StartDebugServer(c.DebugPort, logger); err != nil {
			logger.Log(err.Error())
		}
	}()

	go func() {
		if err := addsvc.GrpcServe(logger); err != nil {
			logger.Log(err.Error())
		}
	}()

	go func() {
		if err := addsvc.HttpServe(logger); err != nil {
			logger.Log(err.Error())
		}
	}()

	// Listen for the interrupt signal.
	<-ctx.Done()

	logger.Log("service", "exiting")
}
