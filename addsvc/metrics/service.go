package metrics

import (
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"

	"github.com/pascallin/go-kit-application/config"
)

var _intsMetrics metrics.Counter
var _charsMetrics metrics.Counter

func GetServiceMetrics() (ints metrics.Counter, chars metrics.Counter) {
	if _intsMetrics != nil && _charsMetrics != nil {
		return _intsMetrics, _charsMetrics
	}
	c := config.GetAddSvcConfig()
	// Business-level metrics.
	_intsMetrics = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "example",
		Subsystem: c.Name,
		Name:      "integers_summed",
		Help:      "Total count of integers summed via the Sum method.",
	}, []string{})
	_charsMetrics = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "example",
		Subsystem: c.Name,
		Name:      "characters_concatenated",
		Help:      "Total count of characters concatenated via the Concat method.",
	}, []string{})

	return _intsMetrics, _charsMetrics
}
