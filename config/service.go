package config

import (
	"os"

	"github.com/joho/godotenv"
)

func init() {
	// load .env
	godotenv.Load()
}

type ServiceConfig struct {
	Host            string
	HostName        string
	Name            string
	DebugPort       int
	HttpPort        int
	GrpcPort        int
	IsNeedDiscovery bool
}

func GetAddSvcConfig() ServiceConfig {
	return ServiceConfig{
		Host:            os.Getenv("SERVICE_HOST"),
		HostName:        "addsvc-test-1",
		Name:            "addsvc",
		DebugPort:       9081,
		HttpPort:        9082,
		GrpcPort:        9083,
		IsNeedDiscovery: false,
	}
}

func GetUserSvcConfig() ServiceConfig {
	return ServiceConfig{
		Host:            os.Getenv("SERVICE_HOST"),
		HostName:        "usersvc-test-1",
		Name:            "usersvc",
		DebugPort:       9091,
		HttpPort:        9092,
		GrpcPort:        9093,
		IsNeedDiscovery: false,
	}
}
