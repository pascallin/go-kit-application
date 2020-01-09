package strsvc

import (
	"context"
	"time"

	"golang.org/x/time/rate"

	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/sony/gobreaker"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
)

// Set collects all of the endpoints that compose an im service. It's meant to
// be used as a helper struct, to collect all of the endpoints into a single
// parameter.
type Set struct {
	UppercaseEndpoint endpoint.Endpoint
}

func (s Set) Uppercase(ctx context.Context, str string) (string, error) {
	resp, err := s.UppercaseEndpoint(ctx, UppercaseRequest{S: str})
	if err != nil {
		return "", err
	}
	response := resp.(UppercaseResponse)
	return response.V, response.Err
}

// New returns a Set that wraps the provided server, and wires in all of the
// expected endpoint middlewares via the various parameters.
func NewEndpoint(svc StringService, logger log.Logger, duration metrics.Histogram, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) Set {
	var uppercaseEndpoint endpoint.Endpoint
	{
		uppercaseEndpoint = MakeUppercaseEndpoint(svc)
		uppercaseEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(uppercaseEndpoint)
		uppercaseEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(uppercaseEndpoint)
		uppercaseEndpoint = opentracing.TraceServer(otTracer, "Uppercase")(uppercaseEndpoint)
		if zipkinTracer != nil {
			uppercaseEndpoint = zipkin.TraceEndpoint(zipkinTracer, "Uppercase")(uppercaseEndpoint)
		}
		// uppercaseEndpoint = LoggingMiddleware(log.With(logger, "method", "Sum"))(uppercaseEndpoint)
		// uppercaseEndpoint = InstrumentingMiddleware(duration.With("method", "Sum"))(uppercaseEndpoint)
	}
	return Set{
		UppercaseEndpoint: uppercaseEndpoint,
	}
}

func MakeUppercaseEndpoint(s StringService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(UppercaseRequest)
		v, err := s.Uppercase(ctx, req.S)
		return UppercaseResponse{V: v, Err: err}, nil
	}
}

type UppercaseRequest struct {
	S string
}

type UppercaseResponse struct {
	V   string `json:"v"`
	Err error  `json:"-"` // should be intercepted by Failed/errorEncoder
}

func (r UppercaseResponse) Failed() error { return r.Err }
