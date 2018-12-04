package helper

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
	return &redis.Pool{
		MaxIdle:   5,
		MaxActive: 12000, // max number of connections
		MaxConnLifetime: 3 * time.Minute,
		IdleTimeout: 200 * time.Second,
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

