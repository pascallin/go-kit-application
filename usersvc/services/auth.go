package services

import (
	"context"
	"errors"

	"github.com/go-kit/log"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pascallin/go-kit-application/config"
	"github.com/pascallin/go-kit-application/usersvc/model"
)

var (
	ErrSignMethod   = errors.New("unexpected signing method")
	ErrInvalidToken = errors.New("invalid token")
)

type IAuthService interface {
	Valid(ctx context.Context, token string) (bool, error)
}

type AuthService struct {
	logger log.Logger
}

func NewAuthService(logger log.Logger) IAuthService {
	return AuthService{
		logger: logger,
	}
}

func (s AuthService) Valid(ctx context.Context, tokenStr string) (bool, error) {
	s.logger.Log("tokenStr", tokenStr)

	tokenClaims, err := jwt.ParseWithClaims(tokenStr, &model.CustomerClaims{}, func(token *jwt.Token) (i interface{}, e error) {
		// if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		// 	return nil, ErrSignMethod
		// }
		return []byte(config.GetAppSecretConfig().JwtSecret), nil
	})
	if err != nil {
		return false, err
	}

	if claim, ok := tokenClaims.Claims.(*model.CustomerClaims); ok && tokenClaims.Valid {
		s.logger.Log("token", tokenStr, "claim", claim)
		return true, nil
	} else {
		return false, ErrInvalidToken
	}
}
