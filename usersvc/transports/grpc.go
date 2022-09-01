package transports

import (
	"context"
	"errors"

	"github.com/go-kit/kit/transport"
	"github.com/go-kit/kit/transport/grpc"
	"github.com/go-kit/log"

	pb "github.com/pascallin/go-kit-application/pb/usersvc"
	"github.com/pascallin/go-kit-application/usersvc/endpoints"
)

type grpcServer struct {
	register       grpc.Handler
	login          grpc.Handler
	updatePassword grpc.Handler
	validToken     grpc.Handler
	pb.UnimplementedUserServer
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
		login: grpc.NewServer(
			endpoints.LoginEndpoint,
			decodeGRPCLoginRequest,
			encodeGRPCLoginResponse,
			options...,
		),
		updatePassword: grpc.NewServer(
			endpoints.UpdatePasswordEndpoint,
			decodeGRPCUpdatePasswordRequest,
			encodeGRPCUpdatePasswordResponse,
			options...,
		),
		validToken: grpc.NewServer(
			endpoints.ValidTokenEndpoint,
			decodeGRPCValidTokenRequest,
			encodeGRPCValidTokenResponse,
			options...,
		),
	}
}

func (s *grpcServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	_, rep, err := s.register.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.RegisterResponse), nil
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
	return &pb.RegisterResponse{Id: res.Id, Err: res.Err}, nil
}

func (s *grpcServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	_, rep, err := s.login.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.LoginResponse), nil
}

func decodeGRPCLoginRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.LoginRequest)
	return endpoints.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}, nil
}

func encodeGRPCLoginResponse(_ context.Context, response interface{}) (interface{}, error) {
	res := response.(endpoints.LoginResponse)
	return &pb.LoginResponse{Token: res.Token, Err: err2str(res.Err)}, nil
}

func (s *grpcServer) UpdatePassword(ctx context.Context, req *pb.UpdatePasswordRequest) (*pb.UpdatePasswordResponse, error) {
	_, rep, err := s.updatePassword.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.UpdatePasswordResponse), nil
}

func decodeGRPCUpdatePasswordRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.UpdatePasswordRequest)
	return endpoints.UpdatePasswordRequest{
		Username:    req.Username,
		Password:    req.Password,
		NewPassword: req.NewPassword,
	}, nil
}

func (s *grpcServer) ValidToken(ctx context.Context, req *pb.ValidTokenReq) (*pb.ValidTokenRes, error) {
	_, rep, err := s.validToken.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.ValidTokenRes), nil
}

func encodeGRPCUpdatePasswordResponse(_ context.Context, response interface{}) (interface{}, error) {
	res := response.(endpoints.UpdatePasswordResponse)
	return &pb.UpdatePasswordResponse{Err: err2str(res.Err)}, nil
}

func decodeGRPCValidTokenRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.ValidTokenReq)
	return endpoints.ValidTokenEndpointRequest{
		Token: req.Token,
	}, nil
}

func encodeGRPCValidTokenResponse(_ context.Context, response interface{}) (interface{}, error) {
	res := response.(endpoints.ValidTokenEndpointResponse)
	return &pb.ValidTokenRes{IsValid: res.IsValid, Err: err2str(res.Err)}, nil
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
