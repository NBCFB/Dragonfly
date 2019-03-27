package Dragonfly

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"time"
)

type RedisCallers struct {
	Client *redis.Client
}

func NewCaller(config *viper.Viper) *RedisCallers {

	if config != nil {
		mode := config.GetString("Mode")
		host := config.GetString(fmt.Sprintf("%v.%v.%v", mode, "redisDB", "host"))
		pass := config.GetString(fmt.Sprintf("%v.%v.%v", mode, "redisDB", "pass"))

		c := redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    "mymaster",
			SentinelAddrs: []string{host + ":26379"},
			Password:      pass,

			MaxRetries: 3,

			DialTimeout:  500 * time.Millisecond,
			ReadTimeout:  500 * time.Millisecond,
			WriteTimeout: 500 * time.Millisecond,

			PoolSize: 10000,
		})

		return &RedisCallers{
			Client: c,
		}
	} else {
		return nil
	}
}