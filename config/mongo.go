package config

import (
	"os"

	"github.com/joho/godotenv"
)

func init() {
	// load .env
	godotenv.Load()
}

type MongoConfig struct {
	URI      string `json:"uri,omitempty"`
	DATABASE string `json:"database"`
}

func GetMongoConfig() MongoConfig {
	return MongoConfig{
		URI:      os.Getenv("MONGODB_URI"),
		DATABASE: os.Getenv("MONGODB_DATABASE"),
	}
}
