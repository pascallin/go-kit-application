package addendpoint

import (
	"context"
	"time"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"

	"github.com/pascallin/go-micro-services/internal/addsvc/addservice"
)

type Set struct {
	SumEndpoint    endpoint.Endpoint
	ConcatEndpoint endpoint.Endpoint
}

func New(svc addservice.AddService, logger log.Logger, duration metrics.Histogram, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) Set {
	var sumEndpoint endpoint.Endpoint
	{
		sumEndpoint = MakeSumEndpoint(svc)
		sumEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(sumEndpoint)
		sumEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(sumEndpoint)
		sumEndpoint = opentracing.TraceServer(otTracer, "Sum")(sumEndpoint)
		if zipkinTracer != nil {
			sumEndpoint = zipkin.TraceEndpoint(zipkinTracer, "Sum")(sumEndpoint)
		}
		sumEndpoint = LoggingMiddleware(log.With(logger, "method", "Sum"))(sumEndpoint)
		sumEndpoint = InstrumentingMiddleware(duration.With("method", "Sum"))(sumEndpoint)
	}
	var concatEndpoint endpoint.Endpoint
	{
		concatEndpoint = MakeConcatEndpoint(svc)
		concatEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 100))(concatEndpoint)
		concatEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(concatEndpoint)
		concatEndpoint = opentracing.TraceServer(otTracer, "Concat")(concatEndpoint)
		if zipkinTracer != nil {
			concatEndpoint = zipkin.TraceEndpoint(zipkinTracer, "Concat")(concatEndpoint)
		}
		concatEndpoint = LoggingMiddleware(log.With(logger, "method", "Concat"))(concatEndpoint)
		concatEndpoint = InstrumentingMiddleware(duration.With("method", "Concat"))(concatEndpoint)
	}
	return Set{
		SumEndpoint:    sumEndpoint,
		ConcatEndpoint: concatEndpoint,
	}
}

func (s Set) Sum(ctx context.Context, a, b int) (int, error) {
	resp, err := s.SumEndpoint(ctx, SumRequest{A: a, B: b})
	if err != nil {
		return 0, err
	}
	response := resp.(SumResponse)
	return response.V, response.Err
}

func (s Set) Concat(ctx context.Context, a, b string) (string, error) {
	resp, err := s.ConcatEndpoint(ctx, ConcatRequest{A: a, B: b})
	if err != nil {
		return "", err
	}
	response := resp.(ConcatResponse)
	return response.V, response.Err
}

func MakeSumEndpoint(s addservice.AddService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(SumRequest)
		v, err := s.Sum(ctx, req.A, req.B)
		return SumResponse{V: v, Err: err}, nil
	}
}

func MakeConcatEndpoint(s addservice.AddService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(ConcatRequest)
		v, err := s.Concat(ctx, req.A, req.B)
		return ConcatResponse{V: v, Err: err}, nil
	}
}

var (
	_ endpoint.Failer = SumResponse{}
	_ endpoint.Failer = ConcatResponse{}
)

type SumRequest struct {
	A, B int
}

type SumResponse struct {
	V   int   `json:"v"`
	Err error `json:"-"` // should be intercepted by Failed/errorEncoder
}

func (r SumResponse) Failed() error { return r.Err }

type ConcatRequest struct {
	A, B string
}

type ConcatResponse struct {
	V   string `json:"v"`
	Err error  `json:"-"`
}

func (r ConcatResponse) Failed() error { return r.Err }
