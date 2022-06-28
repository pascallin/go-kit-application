package config

import (
	"os"

	"github.com/joho/godotenv"
)

func init() {
	// load .env
	godotenv.Load()
}

type Infra struct {
	ZIPKIN_URL string `json:"zipkinUrL"`
	CONSUL_URL string `json:"consulUrl"`
}

func GetInfraConfig() Infra {
	return Infra{
		ZIPKIN_URL: os.Getenv("ZIPKIN_URL"),
		CONSUL_URL: os.Getenv("CONSUL_URL"),
	}
}
