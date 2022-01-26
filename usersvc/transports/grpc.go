package transports

import (
	"context"
	"errors"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	"github.com/go-kit/kit/transport/grpc"

	"github.com/pascallin/go-kit-application/pb"
	"github.com/pascallin/go-kit-application/usersvc/endpoints"
)

type grpcServer struct {
	register grpc.Handler
	pb.UnimplementedUserServer
}

func (s *grpcServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	_, rep, err := s.register.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.RegisterResponse), nil
}

func NewGRPCServer(endpoints endpoints.EndpointSet, logger log.Logger) pb.UserServer {
	options := []grpc.ServerOption{
		grpc.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}
	return &grpcServer{
		register: grpc.NewServer(
			endpoints.RegisterEndpoint,
			decodeGRPCRegisterRequest,
			encodeGRPCRegisterResponse,
			options...,
		),
	}
}

func decodeGRPCRegisterRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.RegisterRequest)
	return endpoints.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
	}, nil
}

func encodeGRPCRegisterResponse(_ context.Context, response interface{}) (interface{}, error) {
	res := response.(endpoints.RegisterResponse)
	return &pb.RegisterResponse{Id: res.Id, Err: err2str(res.Err)}, nil
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
