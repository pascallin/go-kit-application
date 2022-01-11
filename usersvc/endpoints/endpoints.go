package endpoints

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/pascallin/go-kit-application/usersvc/services"
)

type EndpointSet struct {
	RegisterEndpoint endpoint.Endpoint
}

func New() EndpointSet {
	var registerEndpoint endpoint.Endpoint
	{
		registerEndpoint = makeRegisterEndpoint()
	}
	return EndpointSet{
		RegisterEndpoint: registerEndpoint,
	}
}

type RegisterRequest struct {
	Username, Password, Nickname string
}
type RegisterResponse struct {
	Id  string
	Err error
}

func makeRegisterEndpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(RegisterRequest)
		err, id := services.Register(req.Username, req.Password, req.Nickname)
		return RegisterResponse{Id: id.String(), Err: err}, nil
	}
}
