package strtransport

import (
	"context"
	"errors"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"time"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	"github.com/go-kit/kit/transport"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"

	"github.com/pascallin/go-micro-services/pb"
	"github.com/pascallin/go-micro-services/pkg/strsvc"
)

type grpcServer struct {
	uppercase grpctransport.Handler
}

func NewGRPCServer(endpoint strsvc.Set, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, logger log.Logger ) pb.StrServer {
	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}

	if zipkinTracer != nil {
		options = append(options, zipkin.GRPCServerTrace(zipkinTracer))
	}

	return &grpcServer{
		grpctransport.NewServer(
			strsvc.Set{}.UppercaseEndpoint,
			decodeGRPCUppercaseRequest,
			encodeGRPCUppercaseResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(otTracer, "Uppercase", logger)))...,
		),
	}
}

func (s *grpcServer) Uppercase(ctx context.Context, req *pb.UppercaseRequest) (*pb.UppercaseReply, error) {
	_, rep, err := s.uppercase.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.UppercaseReply), nil
}

func NewGRPCClient(conn *grpc.ClientConn, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, logger log.Logger) strsvc.StringService {
	limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 100))
	var options []grpctransport.ClientOption
	if zipkinTracer != nil {
		options = append(options, zipkin.GRPCClientTrace(zipkinTracer))
	}
	var uppercaseEndpoint endpoint.Endpoint
	{
		uppercaseEndpoint = grpctransport.NewClient(
			conn,
			"pb.Str",
			"Uppercase",
			encodeGRPCUppercaseRequest,
			decodeGRPCUppercaseResponse,
			pb.UppercaseReply{},
			append(options, grpctransport.ClientBefore(opentracing.ContextToGRPC(otTracer, logger)))...,
		).Endpoint()
		uppercaseEndpoint = opentracing.TraceClient(otTracer, "Uppercase")(uppercaseEndpoint)
		uppercaseEndpoint = limiter(uppercaseEndpoint)
		uppercaseEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "Uppercase",
			Timeout: 30 * time.Second,
		}))(uppercaseEndpoint)
	}
	return strsvc.Set{
		UppercaseEndpoint:    uppercaseEndpoint,
	}
}

func decodeGRPCUppercaseRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.UppercaseRequest)
	return strsvc.UppercaseRequest{S: string(req.S)}, nil
}

func decodeGRPCUppercaseResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.UppercaseReply)
	return strsvc.UppercaseResponse{V: string(reply.V), Err: str2err(reply.Err)}, nil
}

func encodeGRPCUppercaseRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(strsvc.UppercaseRequest)
	return &pb.UppercaseRequest{S: string(req.S)}, nil
}

func encodeGRPCUppercaseResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(strsvc.UppercaseResponse)
	return &pb.UppercaseReply{V: string(resp.V), Err: err2str(resp.Err)}, nil
}


func str2err(s string) error {
	if s == "" {
		return nil
	}
	return errors.New(s)
}
func err2str(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

