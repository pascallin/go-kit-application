package discovery

import (
	"fmt"
	"log"
	"strconv"
	"sync"

	consulsd "github.com/go-kit/kit/sd/consul"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
)

type KitDiscoverClient struct {
	client      consulsd.Client
	config      *consulapi.Config
	mutex       sync.Mutex
	instanceMap sync.Map
}

type ServiceInstance struct {
	InstanceId   string
	InstanceHost string
	InstancePort int
}

func NewKitDiscoverClient(addr string) (client *KitDiscoverClient, err error) {
	c := new(KitDiscoverClient)
	config := consulapi.DefaultConfig()
	config.Address = addr
	apiClient, err := consulapi.NewClient(config)
	if err != nil {
		return nil, err
	}
	c.client = consulsd.NewClient(apiClient)
	c.config = config
	return c, err
}

func (c *KitDiscoverClient) Register(name string, instance ServiceInstance, healthCheckUrl string, meta map[string]string) bool {
	serviceRegistration := &consulapi.AgentServiceRegistration{
		ID:      instance.InstanceId,
		Name:    name,
		Address: instance.InstanceHost,
		Port:    instance.InstancePort,
		Meta:    meta,
		Check: &consulapi.AgentServiceCheck{
			DeregisterCriticalServiceAfter: "30s",
			HTTP:                           fmt.Sprintf("http://%s:%s/%s", instance.InstanceHost, strconv.Itoa(instance.InstancePort), healthCheckUrl),
			Interval:                       "15s",
		},
	}
	err := c.client.Register(serviceRegistration)
	if err != nil {
		log.Println("Register Service Error!")
		return false
	}
	log.Println("Register Service Success!")
	return true
}

func (c *KitDiscoverClient) DeRegister(instanceId string) bool {
	serviceRegistration := &consulapi.AgentServiceRegistration{
		ID: instanceId,
	}
	err := c.client.Deregister(serviceRegistration)
	if err != nil {
		log.Println("Deregister Service Error!")
		return false
	}

	log.Println("Deregister Service Success!")
	return true
}

func (c *KitDiscoverClient) DiscoveryServices(serviceName string) []interface{} {

	// try get from memory
	instanceList, ok := c.instanceMap.Load(serviceName)
	if ok {
		return instanceList.([]interface{})
	}

	// behavior lock
	c.mutex.Lock()
	defer c.mutex.Unlock()
	instanceList, ok = c.instanceMap.Load(serviceName)
	// try get from memory again
	if ok {
		return instanceList.([]interface{})
	}

	// watch service change event, update client map
	go func() {
		params := make(map[string]interface{})
		params["type"] = "service"
		params["service"] = serviceName
		plan, _ := watch.Parse(params)
		plan.Handler = func(u uint64, i interface{}) {
			if i == nil {
				return
			}

			v, ok := i.([]*consulapi.ServiceEntry)
			if !ok {
				return
			}

			if len(v) == 0 {
				c.instanceMap.Store(serviceName, []interface{}{})
			}

			var healthServices []interface{}
			for _, service := range v {
				if service.Checks.AggregatedStatus() == consulapi.HealthPassing {
					healthServices = append(healthServices, service)
				}
			}
			c.instanceMap.Store(serviceName, healthServices)
		}
		defer plan.Stop()
		plan.Run(c.config.Address)
	}()

	// get entries from etcd
	entries, _, err := c.client.Service(serviceName, "", false, nil)
	if err != nil {
		c.instanceMap.Store(serviceName, []interface{}{})
		log.Println("Discover Service Error")
		return nil
	}

	instances := make([]interface{}, 0, len(entries))
	for _, instance := range entries {
		instances = append(instances, instance)
	}

	c.instanceMap.Store(serviceName, instances)
	return instances
}
