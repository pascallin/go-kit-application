package strsvc

import (
	"context"
	"errors"
	"strings"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/log"
)

// StringService provides operations on strings.
type StringService interface {
	Uppercase(context.Context, string) (string, error)
	// Count(context.Context, strsvc) int
}

// stringService is a concrete implementation of StringService
type stringService struct{}

func (stringService) Uppercase(_ context.Context, s string) (string, error) {
	if s == "" {
		return "", errors.New("empty strsvc")
	}
	return strings.ToUpper(s), nil
}

func (stringService) Count(_ context.Context, s string) int {
	return len(s)
}

// NewStringService returns a naïve, stateless implementation of Service.
func NewStringService() StringService {
	return stringService{}
}

// // New returns a basic Service with all of the expected middlewares wired in.
func NewService(logger log.Logger, ints, chars metrics.Counter) StringService {
	var svc StringService
	{
		svc = NewStringService()
		// svc = LoggingMiddleware(logger)(svc)
		//svc = InstrumentingMiddleware(ints, chars)(svc)
	}
	return svc
}
