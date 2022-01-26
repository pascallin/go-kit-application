package main

import (
	"github.com/pascallin/go-kit-application/gateway"
	"github.com/pascallin/go-kit-application/vendor/github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	go gateway.StartGateway()
}
