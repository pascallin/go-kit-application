package config

import (
	"fmt"

	"github.com/caarlos0/env/v6"
)

type MongoConfig struct {
	URI      string `env:"MONGODB_URI"`
	DATABASE string `env:"MONGODB_DATABASE"`
}

func GetMongoConfig() MongoConfig {
	cfg := MongoConfig{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}
	return cfg
}
