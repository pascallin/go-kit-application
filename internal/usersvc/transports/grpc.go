package transports

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	"github.com/go-kit/kit/transport/grpc"

	"github.com/pascallin/go-kit-application/internal/usersvc"
	"github.com/pascallin/go-kit-application/pb"
)

type grpcServer struct {
	register grpc.Handler
}

func (s *grpcServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	_, rep, err := s.register.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.RegisterResponse), nil
}

func NewGRPCServer(endpoints usersvc.EndpointSet, logger log.Logger) pb.UserServer {
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
	return usersvc.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
	}, nil
}

func encodeGRPCRegisterResponse(_ context.Context, response interface{}) (interface{}, error) {
	res := response.(usersvc.RegisterResponse)
	return &pb.RegisterResponse{Err: res.Err.Error(), Id: res.Id}, nil
}
