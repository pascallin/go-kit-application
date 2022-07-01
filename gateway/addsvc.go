package gateway

import (
	"context"
	"io"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
	"github.com/gorilla/mux"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"google.golang.org/grpc"

	svcendpoints "github.com/pascallin/go-kit-application/addsvc/endpoints"
	"github.com/pascallin/go-kit-application/addsvc/services"
	"github.com/pascallin/go-kit-application/addsvc/transports"
	"github.com/pascallin/go-kit-application/config"
	"github.com/pascallin/go-kit-application/pkg"
)

func registerAddsvc(ctx context.Context, r *mux.Router, tracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, client consulsd.Client) {
	config := config.GetGatewayConfig()
	logger := pkg.GetLogger()

	var (
		tags        = []string{}
		passingOnly = true
		endpoints   = svcendpoints.Set{}
		instancer   = consulsd.NewInstancer(client, logger, "addsvc", tags, passingOnly)
	)
	{
		factory := addsvcFactory(svcendpoints.MakeSumEndpoint, tracer, zipkinTracer, logger)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(config.RetryMax, config.RetryTimeout, balancer)
		endpoints.SumEndpoint = retry
	}
	{
		factory := addsvcFactory(svcendpoints.MakeConcatEndpoint, tracer, zipkinTracer, logger)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(config.RetryMax, config.RetryTimeout, balancer)
		endpoints.ConcatEndpoint = retry
	}

	// Here we leverage the fact that addsvc comes with a constructor for an
	// HTTP handler, and just install it under a particular path prefix in
	// our router.

	r.PathPrefix("/addsvc").Handler(http.StripPrefix("/addsvc", transports.NewHTTPHandler(endpoints, tracer, zipkinTracer, logger)))
}

func addsvcFactory(makeEndpoint func(services.Service) endpoint.Endpoint, tracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, logger log.Logger) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		// We could just as easily use the HTTP or Thrift client package to make
		// the connection to addsvc. We've chosen gRPC arbitrarily. Note that
		// the transport is an implementation detail: it doesn't leak out of
		// this function. Nice!

		conn, err := grpc.Dial(instance, grpc.WithInsecure())
		if err != nil {
			return nil, nil, err
		}
		service := transports.NewGRPCClient(conn, tracer, zipkinTracer, logger)
		endpoint := makeEndpoint(service)

		// Notice that the addsvc gRPC client converts the connection to a
		// complete addsvc, and we just throw away everything except the method
		// we're interested in. A smarter factory would mux multiple methods
		// over the same connection. But that would require more work to manage
		// the returned io.Closer, e.g. reference counting. Since this is for
		// the purposes of demonstration, we'll just keep it simple.

		return endpoint, conn, nil
	}
}
