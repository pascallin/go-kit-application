package conn

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"

	"github.com/pascallin/go-kit-application/config"
)

var (
	ronce               sync.Once
	redisSingleInstance *redis.Client
)

func GetRedis() *redis.Client {
	ronce.Do(func() {
		client, err := initRedis()
		if err != nil {
			log.Error(err)
			return
		}
		redisSingleInstance = client
	})
	return redisSingleInstance
}

func initRedis() (*redis.Client, error) {
	c := config.GetRedisConfig()

	db, err := strconv.ParseInt(c.Database, 10, 32)
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(&redis.Options{
		PoolSize: 1000,
		Addr:     fmt.Sprintf("%s:%s", c.Host, c.Port),
		Password: c.Password,
		DB:       int(db),
	})

	return rdb, nil
}
