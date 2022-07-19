package pkg

import (
	"fmt"
	"net"
	"net/http"

	"github.com/go-kit/kit/log"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func StartDebugServer(port int, logger log.Logger) error {
	http.DefaultServeMux.Handle("/metrics", promhttp.Handler())

	// The debug listener mounts the http.DefaultServeMux, and serves up
	// stuff like the Prometheus metrics route, the Go debug and profiling
	// routes, and so on.
	debugListener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	logger.Log("transport", "http", "addr", port)

	if err := http.Serve(debugListener, http.DefaultServeMux); err != nil {
		return err
	}
	return nil

}
