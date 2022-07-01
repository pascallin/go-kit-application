package config

import "time"

type GatewayConfig struct {
	HttpPort     int
	RetryMax     int
	RetryTimeout time.Duration
}

func GetGatewayConfig() GatewayConfig {
	return GatewayConfig{
		HttpPort:     9090,
		RetryMax:     3,
		RetryTimeout: 500 * time.Microsecond,
	}
}
