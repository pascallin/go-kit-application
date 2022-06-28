package conn

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/go-redis/redis/v8"

	"github.com/pascallin/go-kit-application/config"
)

var rlock = &sync.Mutex{}
var redisSingleInstance *redis.Client

func GetRedis() *redis.Client {
	if redisSingleInstance == nil {
		rlock.Lock()
		defer rlock.Unlock()
		if redisSingleInstance == nil {
			client := initRedis()
			redisSingleInstance = client
		}
	}
	return redisSingleInstance
}

func initRedis() *redis.Client {
	c := config.GetRedisConfig()

	db, err := strconv.ParseInt(c.Database, 10, 32)
	if err != nil {
		panic(err)
	}

	rdb := redis.NewClient(&redis.Options{
		PoolSize: 1000,
		Addr:     fmt.Sprintf("%s:%s", c.Host, c.Port),
		Password: c.Password,
		DB:       int(db),
	})

	return rdb
}
