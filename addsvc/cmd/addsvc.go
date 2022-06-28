package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/pascallin/go-kit-application/addsvc"
	"github.com/pascallin/go-kit-application/config"
	"github.com/pascallin/go-kit-application/pkg"
	"github.com/pascallin/go-kit-application/pkg/discovery"
)

func main() {
	logger := pkg.GetLogger()
	c := config.GetAddSvcConfig()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if c.IsNeedDiscovery {
		client, err := discovery.NewKitDiscoverClient()
		if err != nil {
			panic(err)
		}
		status := client.Register(c.Name, discovery.ServiceInstance{
			InstanceId:   c.HostName,
			InstanceHost: c.Host,
			InstancePort: c.GrpcPort,
		}, make(map[string]string))
		logger.Log("consul discovery register ", status)
		defer client.DeRegister(c.Name)
	}

	http.DefaultServeMux.Handle("/metrics", promhttp.Handler())
	go func() {
		// The debug listener mounts the http.DefaultServeMux, and serves up
		// stuff like the Prometheus metrics route, the Go debug and profiling
		// routes, and so on.
		debugListener, err := net.Listen("tcp", fmt.Sprintf(":%d", c.DebugPort))
		if err != nil {
			logger.Log("transport", "debug/HTTP", "during", "Listen", "err", err)
			return
		}
		if err := http.Serve(debugListener, http.DefaultServeMux); err != nil {
			logger.Log(err.Error())
		}
	}()

	go func() {
		if err := addsvc.GrpcServe(logger); err != nil {
			logger.Log(err.Error())
		}
	}()

	// Listen for the interrupt signal.
	<-ctx.Done()

	logger.Log("service", "exiting")
}
