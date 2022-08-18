package config

import (
	"fmt"

	"github.com/caarlos0/env/v6"
)

type MysqlConfig struct {
	Host     string `env:"MYSQL_HOST"`
	Port     string `env:"MYSQL_PORT"`
	User     string `env:"MYSQL_USER"`
	Password string `env:"MYSQL_PASSWORD"`
	Database string `env:"MYSQL_DATABASE"`
}

func GetMysqlConfig() MysqlConfig {
	cfg := MysqlConfig{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}
	return cfg
}
