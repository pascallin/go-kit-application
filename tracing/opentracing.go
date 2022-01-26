package tracing

import (
	"os"

	"github.com/lightstep/lightstep-tracer-go"
	stdopentracing "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"sourcegraph.com/sourcegraph/appdash"
	appdashot "sourcegraph.com/sourcegraph/appdash/opentracing"
)

var (
	zipkinURL      = os.Getenv("DEFAULT_ZIPKIN_URL")
	zipkinBridge   = false
	appdashAddr    = ""
	lightstepToken = ""
)

func NewOpentracingTracer(hostPort, serviceName string) (zipkinTracer *zipkin.Tracer, ocTracer stdopentracing.Tracer, err error) {
	{
		var (
			err      error
			reporter = zipkinhttp.NewReporter(zipkinURL)
		)
		defer reporter.Close()
		zEP, _ := zipkin.NewEndpoint(serviceName, hostPort)
		zipkinTracer, err = zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(zEP))
		if err != nil {
			return nil, nil, err
		}
	}
	// Determine which OpenTracing tracer to use. We'll pass the tracer to all the
	// components that use it, as a dependency.
	{
		if zipkinBridge && zipkinTracer != nil {
			ocTracer = zipkinot.Wrap(zipkinTracer)
			zipkinTracer = nil // do not instrument with both native tracer and opentracing bridge
		} else if lightstepToken != "" {
			ocTracer = lightstep.NewTracer(lightstep.Options{
				AccessToken: lightstepToken,
			})
			defer lightstep.FlushLightStepTracer(ocTracer)
		} else if appdashAddr != "" {
			ocTracer = appdashot.NewTracer(appdash.NewRemoteCollector(appdashAddr))
		} else {
			ocTracer = stdopentracing.GlobalTracer() // no-op
		}
	}

	return zipkinTracer, ocTracer, nil
}
