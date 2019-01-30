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

func NewClient(config *viper.Viper) (*redis.Client, error) {
	mode := config.GetString("Mode")
	host := config.GetString(fmt.Sprintf("%v.%v.%v", mode, "redisDB", "host"))
	pass := config.GetString(fmt.Sprintf("%v.%v.%v", mode, "redisDB", "pass"))

	redisdb := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:		"mymaster",
		SentinelAddrs: 	[]string{ host + ":26379" },
		Password: 		pass,

		MaxRetries:     3,

		DialTimeout:	500 * time.Millisecond,
		ReadTimeout:	500 * time.Millisecond,
		WriteTimeout:	500 * time.Millisecond,

		PoolSize:       10000,
	})

	_, err := redisdb.Ping().Result()
	if err != nil {
		return nil, err
	}

	return redisdb, nil
}

func NewCaller(client *redis.Client) *RedisCallers {
	return &RedisCallers{
		Client: client,
	}
}

