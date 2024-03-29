// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package services

import (
	"github.com/go-kit/log"
	"go.mongodb.org/mongo-driver/mongo"
)

// Injectors from wire.go:

func InitializeService(db *mongo.Database, logger log.Logger) (Service, error) {
	iUserService := NewUserService(db, logger)
	iAuthService := NewAuthService(logger)
	service := NewService(iUserService, iAuthService)
	return service, nil
}
