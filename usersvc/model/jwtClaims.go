package model

import (
	"github.com/golang-jwt/jwt/v4"
)

type CustomerClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}
