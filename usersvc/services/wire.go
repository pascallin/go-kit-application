//go:build wireinject
// +build wireinject

package services

import (
	"github.com/go-kit/log"
	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitializeService(db *mongo.Database, logger log.Logger) (Service, error) {
	wire.Build(NewService, NewUserService, NewAuthService)
	return Service{}, nil
}
