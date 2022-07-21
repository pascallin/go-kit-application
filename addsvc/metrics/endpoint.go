package metrics

import (
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"

	"github.com/pascallin/go-kit-application/config"
)

var (
	_requestDurationMetrics metrics.Histogram
)

func GetEndpointMetrics() (duration metrics.Histogram) {
	if _requestDurationMetrics != nil {
		return _requestDurationMetrics
	}
	c := config.GetAddSvcConfig()
	// Endpoint-level metrics.
	_requestDurationMetrics = prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "example",
		Subsystem: c.Name,
		Name:      "request_duration_seconds",
		Help:      "Request duration in seconds.",
	}, []string{"method", "success"})

	return _requestDurationMetrics
}
