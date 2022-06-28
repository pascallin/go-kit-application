package config

import (
	"os"

	"github.com/joho/godotenv"
)

func init() {
	// load .env
	godotenv.Load()
}

type AppSecret struct {
	JWT_SECRET string
}

func GetAppSecretConfig() AppSecret {
	return AppSecret{
		JWT_SECRET: os.Getenv("JWT_SECRET"),
	}
}
