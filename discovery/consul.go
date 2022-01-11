package discovery

import (
	"github.com/hashicorp/consul/api"
)

type ConsulControl struct {
	Client *api.Client
	Config *api.AgentServiceRegistration
}

func ConnConsul(addr string) (*ConsulControl, error) {
	csc := new(ConsulControl)
	config := api.DefaultConfig()
	config.Address = addr
	var err error
	csc.Client, err = api.NewClient(config)
	return csc, err
}

func (c *ConsulControl) Register(config *api.AgentServiceRegistration) error {
	c.Config = config

	var err error
	if err = c.Client.Agent().ServiceRegister(c.Config); err != nil {
		return err
	}
	return nil
}

func (c *ConsulControl) UnRegister(serviceId string) error {
	var err error
	if err = c.Client.Agent().ServiceDeregister(serviceId); err != nil {
		return err
	}
	return nil
}
