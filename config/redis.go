package config

import (
	"os"

	"github.com/joho/godotenv"
)

func init() {
	// load .env
	godotenv.Load()
}

type Redis struct {
	Host     string `json:"host,omitempty"`
	Port     string `json:"port,omitempty"`
	Password string `json:"password,omitempty"`
	Database string `json:"database,omitempty"`
}

func GetRedisConfig() Redis {
	return Redis{
		Host:     os.Getenv("REDIS_HOST"),
		Port:     os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		Database: os.Getenv("REDIS_DB"),
	}
}
