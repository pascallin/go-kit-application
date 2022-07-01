package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/pascallin/go-kit-application/config"
	"github.com/pascallin/go-kit-application/pkg"
	"github.com/pascallin/go-kit-application/usersvc"
)

func main() {
	c := config.GetUserSvcConfig()
	logger := pkg.GetLogger()

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
		if err := usersvc.GrpcServe(logger); err != nil {
			logger.Log(err.Error())
		}
	}()

	// Listen for the interrupt signal.
	<-ctx.Done()

	logger.Log("service", "exiting")
}
