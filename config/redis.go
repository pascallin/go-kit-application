package config

import (
	"fmt"

	"github.com/caarlos0/env/v6"
)

type RedisConfig struct {
	Host     string `env:"REDIS_HOST"`
	Port     string `env:"REDIS_PORT"`
	Password string `env:"REDIS_PASSWORD"`
	Database string `env:"REDIS_DB"`
}

func GetRedisConfig() RedisConfig {
	cfg := RedisConfig{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}
	return cfg
}
