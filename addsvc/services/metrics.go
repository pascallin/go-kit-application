package services

import (
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"

	"github.com/pascallin/go-kit-application/config"
)

func InitMetrics() (ints metrics.Counter, chars metrics.Counter) {
	c := config.GetAddSvcConfig()
	// Business-level metrics.
	ints = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "example",
		Subsystem: c.Name,
		Name:      "integers_summed",
		Help:      "Total count of integers summed via the Sum method.",
	}, []string{})
	chars = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "example",
		Subsystem: c.Name,
		Name:      "characters_concatenated",
		Help:      "Total count of characters concatenated via the Concat method.",
	}, []string{})

	return ints, chars
}
