package Dragonfly

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	Pool *redis.Pool
)

func Init(config *viper.Viper) {
	Pool = 	newPool(config)
	removeHooks()
}

func newPool(config *viper.Viper) *redis.Pool {
	mode := config.GetString("Mode")
	host := config.GetString(fmt.Sprintf("%v.%v.%v", mode, "redisDB", "host"))
	pass := config.GetString(fmt.Sprintf("%v.%v.%v", mode, "redisDB", "pass"))
	maxIdle := config.GetInt(fmt.Sprintf("%v.%v.%v", mode, "redisDB", "maxIdle"))
	maxActive := config.GetInt(fmt.Sprintf("%v.%v.%v", mode, "redisDB", "maxActive"))
	maxConnLifetime := config.GetDuration(fmt.Sprintf("%v.%v.%v", mode, "redisDB", "maxConnLifetime"))
	idleTimeout := config.GetDuration(fmt.Sprintf("%v.%v.%v", mode, "redisDB", "idleTimeout"))

	return &redis.Pool{
		MaxIdle:   maxIdle,
		MaxActive: maxActive, // max number of connections
		MaxConnLifetime: maxConnLifetime * time.Second,
		IdleTimeout: idleTimeout * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host + ":6379", redis.DialPassword(pass))
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}

func removeHooks() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGKILL)
	go func() {
		<-c
		Pool.Close()
		os.Exit(0)
	}()
}

