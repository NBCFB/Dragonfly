package helper

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/spf13/viper"
)

func NewPool() *redis.Pool {
	mode := viper.GetString("Mode")
	host := viper.GetString(fmt.Sprintf("%v.%v.%v", mode, "redisDB", "host"))
	pass := viper.GetString(fmt.Sprintf("%v.%v.%v", mode, "redisDB", "pass"))
	return &redis.Pool{
		MaxIdle:   80,
		MaxActive: 12000, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host + ":6379", redis.DialPassword(pass))
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}

