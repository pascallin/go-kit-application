package stringsvc

import (
	"errors"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
)

// StringService provides operations on strings.
type StringService interface {
	Uppercase(string) (string, error)
	Count(string) int
}

// stringService is a concrete implementation of StringService
type stringService struct{}

func (stringService) Uppercase(s string) (string, error) {
	if s == "" {
		return "", errors.New("empty strsvc")
	}
	return strings.ToUpper(s), nil
}

func (stringService) Count(s string) int {
	return len(s)
}

// ErrEmpty is returned when input string is empty
var ErrEmpty = errors.New("Empty string")

// NewStringService returns a naïve, stateless implementation of Service.
func NewStringService() stringService {
	return stringService{}
}

// // New returns a basic Service with all of the expected middlewares wired in.
func NewService(logger log.Logger, ints, chars metrics.Counter) stringService {
	var svc stringService
	{
		svc = NewStringService()
		// svc = LoggingMiddleware(logger)(svc)
		//svc = InstrumentingMiddleware(ints, chars)(svc)
	}
	return svc
}

// ServiceMiddleware is a chainable behavior modifier for StringService.
type ServiceMiddleware func(StringService) StringService
