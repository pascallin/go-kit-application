package pkg

import (
	stdopentracing "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	zipkin "github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"

	log "github.com/sirupsen/logrus"

	"github.com/pascallin/go-kit-application/config"
)

func InitTracer(name string) (*zipkin.Tracer, stdopentracing.Tracer, error) {
	var (
		zipkinURL    = config.GetInfraConfig().ZIPKIN_URL
		zipkinBridge = true
	)
	var zipkinTracer *zipkin.Tracer
	{
		if zipkinURL != "" {
			var (
				err         error
				hostPort    = "localhost:80"
				serviceName = name
				reporter    = zipkinhttp.NewReporter(zipkinURL)
			)
			defer reporter.Close()
			zEP, _ := zipkin.NewEndpoint(serviceName, hostPort)
			zipkinTracer, err = zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(zEP))
			if err != nil {
				log.Error("err", err)
				return nil, nil, err
			}
			if !(zipkinBridge) {
				log.WithFields(log.Fields{
					"tracer": "Zipkin",
					"type":   "Native",
					"URL":    zipkinURL,
				}).Infof("Zipkin tracer without bridge")
			}
		}
	}
	// Determine which OpenTracing tracer to use. We'll pass the tracer to all the
	// components that use it, as a dependency.
	var tracer stdopentracing.Tracer
	{
		if zipkinBridge && zipkinTracer != nil {
			tracer = zipkinot.Wrap(zipkinTracer)
			zipkinTracer = nil // do not instrument with both native tracer and opentracing bridge
		} else {
			tracer = stdopentracing.GlobalTracer() // no-op
		}
	}

	return zipkinTracer, tracer, nil
}
