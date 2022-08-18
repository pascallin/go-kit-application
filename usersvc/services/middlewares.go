package services

import (
	"context"

	"github.com/go-kit/kit/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Middleware describes a service (as opposed to endpoint) middleware.
type Middleware func(Service) Service

// LoggingMiddleware takes a logger as a dependency
// and returns a service Middleware.
func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next Service) Service {
		return loggingMiddleware{logger, next}
	}
}

type loggingMiddleware struct {
	logger log.Logger
	next   Service
}

func (mw loggingMiddleware) Register(ctx context.Context, username, password, nickname string) (id primitive.ObjectID, err error) {
	defer func() {
		mw.logger.Log("method", "Register", "username", username, "err", err)
	}()
	return mw.next.Register(ctx, username, password, nickname)
}

func (mw loggingMiddleware) Login(ctx context.Context, username, password string) (token string, err error) {
	defer func() {
		mw.logger.Log("method", "Login", "username", username, "err", err)
	}()
	return mw.next.Login(ctx, username, password)
}

func (mw loggingMiddleware) UpdatePassword(ctx context.Context, username string, password string, newPassword string) (err error) {
	defer func() {
		mw.logger.Log("method", "UpdatePassword", "username", username, "err", err)
	}()
	return mw.next.UpdatePassword(ctx, username, password, newPassword)
}
