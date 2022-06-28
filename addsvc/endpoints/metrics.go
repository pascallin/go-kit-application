package endpoints

import (
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"

	"github.com/pascallin/go-kit-application/config"
)

func InitMetrics() (duration metrics.Histogram) {
	c := config.GetAddSvcConfig()
	// Endpoint-level metrics.
	duration = prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "example",
		Subsystem: c.Name,
		Name:      "request_duration_seconds",
		Help:      "Request duration in seconds.",
	}, []string{"method", "success"})

	return duration
}
