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

func Init() {
	Pool = 	newPool()
	removeHooks()
}

func newPool() *redis.Pool {
	mode := viper.GetString("Mode")
	host := viper.GetString(fmt.Sprintf("%v.%v.%v", mode, "redisDB", "host"))
	pass := viper.GetString(fmt.Sprintf("%v.%v.%v", mode, "redisDB", "pass"))
	maxIdle := viper.GetInt(fmt.Sprintf("%v.%v.%v", mode, "redisDB", "maxIdle"))
	maxActive := viper.GetInt(fmt.Sprintf("%v.%v.%v", mode, "redisDB", "maxActive"))
	maxConnLifetime := viper.GetDuration(fmt.Sprintf("%v.%v.%v", mode, "redisDB", "maxConnLifetime"))
	idleTimeout := viper.GetDuration(fmt.Sprintf("%v.%v.%v", mode, "redisDB", "idleTimeout"))

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

