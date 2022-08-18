package config

import (
	"fmt"

	"github.com/caarlos0/env/v6"
)

type InfraConfig struct {
	ZIPKIN_URL string `env:"ZIPKIN_URL"`
	CONSUL_URL string `env:"CONSUL_URL"`
}

func GetInfraConfig() InfraConfig {
	cfg := InfraConfig{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}
	return cfg
}
