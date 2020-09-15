package register

import (
	"github.com/hashicorp/consul/api"
)

type ConsulContrl struct {
	Client *api.Client
	Config * api.AgentServiceRegistration
}

func ConnConsul(addr string) (*ConsulContrl, error) {
	csc := new(ConsulContrl)
	config := api.DefaultConfig()
	config.Address = addr
	var err error
	csc.Client,err = api.NewClient(config)
	return csc,err
}

func (c *ConsulContrl) Register (config *api.AgentServiceRegistration) error {
	c.Config = config

	var err error
	if err = c.Client.Agent().ServiceRegister(c.Config); err != nil {
		return err
	}
	return nil
}

func (c *ConsulContrl) UnRegister (serviceId string) error {
	var err error
	if err = c.Client.Agent().ServiceDeregister(serviceId); err != nil {
		return err
	}
	return nil
}