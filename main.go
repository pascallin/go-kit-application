package main

import (
	"github.com/pascallin/go-kit-application/addsvc"
	"github.com/pascallin/go-kit-application/gateway"
	"github.com/pascallin/go-kit-application/usersvc"
)

func main() {
	addsvc.StartAddSVCService()
	usersvc.StartUserSVCService()
	gateway.StartGateway()
}
