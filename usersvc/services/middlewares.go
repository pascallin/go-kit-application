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

func (mw loggingMiddleware) Register(ctx context.Context, username, password, nickname string) (err error, id primitive.ObjectID) {
	defer func() {
		mw.logger.Log("method", "Register", "username", username, "username", username, "err", err)
	}()
	return mw.next.Register(ctx, username, password, nickname)
}
