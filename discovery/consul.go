package discovery

import (
	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"os"

	"github.com/pascallin/go-micro-services/config"
)

func getClient () consulsd.Client {

	config := Config.ReadConfig()
	logger := getLogger()

	var client consulsd.Client
	{
		consulConfig := api.DefaultConfig()
		consulConfig.Address = config.Consul
		consulClient, err := api.NewClient(consulConfig)
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
		client = consulsd.NewClient(consulClient)
	}
	return client
}