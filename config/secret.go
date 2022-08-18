package config

import (
	"fmt"

	"github.com/caarlos0/env/v6"
)

type AppSecret struct {
	JwtSecret string `env:"JWT_SECRET"`
}

func GetAppSecretConfig() AppSecret {
	cfg := AppSecret{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}
	return cfg
}
