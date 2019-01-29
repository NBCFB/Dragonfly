package Dragonfly

import (
	"github.com/go-redis/redis"
	"time"
)

func NewConnection() (*redis.Client, error) {
	//mode := config.GetString("Mode")
	//host := config.GetString(fmt.Sprintf("%v.%v.%v", mode, "redisDB", "host"))
	host := "127.0.0.1"
	//pass := config.GetString(fmt.Sprintf("%v.%v.%v", mode, "redisDB", "pass"))

	redisdb := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:		"mymaster",
		SentinelAddrs: 	[]string{ host + ":26379" },
		Password: 		"401BoogiesFightes307Woogies",

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

