package endpoints

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	"github.com/go-kit/log"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"

	userservice "github.com/pascallin/go-kit-application/usersvc/services"
)

type EndpointSet struct {
	RegisterEndpoint       endpoint.Endpoint
	LoginEndpoint          endpoint.Endpoint
	UpdatePasswordEndpoint endpoint.Endpoint
}

func New(svc userservice.Service, logger log.Logger, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) EndpointSet {
	var registerEndpoint, loginEndpoint, updatePasswordEndpoint endpoint.Endpoint
	{
		registerEndpoint = makeRegisterEndpoint(svc)
		registerEndpoint = LoggingMiddleware(log.With(logger, "method", "Register"))(registerEndpoint)
		registerEndpoint = opentracing.TraceServer(otTracer, "Register")(registerEndpoint)
		if zipkinTracer != nil {
			registerEndpoint = zipkin.TraceEndpoint(zipkinTracer, "Register")(registerEndpoint)
		}
	}
	{
		loginEndpoint = makeLoginEndpoint(svc)
		loginEndpoint = LoggingMiddleware(log.With(logger, "method", "Login"))(loginEndpoint)
		loginEndpoint = opentracing.TraceServer(otTracer, "Login")(loginEndpoint)
		if zipkinTracer != nil {
			loginEndpoint = zipkin.TraceEndpoint(zipkinTracer, "Login")(loginEndpoint)
		}
	}
	{
		updatePasswordEndpoint = makeUpdatePasswordEndpoint(svc)
		updatePasswordEndpoint = LoggingMiddleware(log.With(logger, "method", "UpdatePassword"))(updatePasswordEndpoint)
		updatePasswordEndpoint = opentracing.TraceServer(otTracer, "UpdatePassword")(updatePasswordEndpoint)
		if zipkinTracer != nil {
			updatePasswordEndpoint = zipkin.TraceEndpoint(zipkinTracer, "UpdatePassword")(updatePasswordEndpoint)
		}
	}
	return EndpointSet{
		RegisterEndpoint:       registerEndpoint,
		LoginEndpoint:          loginEndpoint,
		UpdatePasswordEndpoint: updatePasswordEndpoint,
	}
}

type RegisterRequest struct {
	Username, Password, Nickname string
}
type RegisterResponse struct {
	Id  string
	Err error
}

func makeRegisterEndpoint(s userservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(RegisterRequest)
		err, id := s.Register(ctx, req.Username, req.Password, req.Nickname)
		return RegisterResponse{Id: id.String(), Err: err}, nil
	}
}

type LoginRequest struct {
	Username, Password string
}

type LoginResponse struct {
	Token string
	Err   error
}

func makeLoginEndpoint(s userservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(LoginRequest)
		err, token := s.Login(ctx, req.Username, req.Password)
		return LoginResponse{Token: token, Err: err}, nil
	}
}

type UpdatePasswordRequest struct {
	Username, Password, NewPassword string
}

type UpdatePasswordResponse struct {
	Err error
}

func makeUpdatePasswordEndpoint(s userservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(UpdatePasswordRequest)
		err = s.UpdatePassword(ctx, req.Username, req.Password, req.NewPassword)
		return UpdatePasswordResponse{Err: err}, nil
	}
}
