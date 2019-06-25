package Dragonfly

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"strings"
	"time"
)

// A RedisObj represents an object with string typed key and value.
type RedisObj struct {
	K string
	V string
}

// A RedisError represents a structured redis error.
type RedisError struct {
	Action string
	Key    string
	Val    string
	ErrMsg string
}

// Error returns a formatted redis error message.
func (e *RedisError) Error() string {
	return fmt.Sprintf("Action[%s] Key=%s, Val=%s, %s", e.Action, e.Key, e.Val, e.ErrMsg)
}

// RedisCallers represent a client that connects to our redis DB.
type RedisCallers struct {
	Client *redis.Client
}

// NewCaller initialises a new connection client to the redis DB.
func NewCallerSentinel(config *viper.Viper) *RedisCallers {

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
	}

	return nil
}

func NewCaller(config *viper.Viper) *RedisCallers {

	if config != nil {
		mode := config.GetString("Mode")
		host := config.GetString(fmt.Sprintf("%v.%v.%v", mode, "redisDB", "host"))
		pass := config.GetString(fmt.Sprintf("%v.%v.%v", mode, "redisDB", "pass"))

		c := redis.NewClient(&redis.Options{
			Addr:     host + ":6379",
			Password: pass,

			MaxRetries: 3,

			DialTimeout:  500 * time.Millisecond,
			ReadTimeout:  500 * time.Millisecond,
			WriteTimeout: 500 * time.Millisecond,

			PoolSize: 10000,
		})

		return &RedisCallers{
			Client: c,
		}
	}

	return nil
}

// Set creates a new entry or updates an existed entry with a specified lifetime, the default lifetime is zero.
func (c *RedisCallers) Set(k, v string, exp time.Duration) (newVal string, err error) {
	if validate(k, v) {
		_, err = c.Client.Set(k, v, exp).Result()
		if err != nil {
			return newVal, &RedisError{
				Action: "Set",
				Key:    k,
				Val:    v,
				ErrMsg: err.Error(),
			}
		}
	} else {
		return newVal, &RedisError{
			Action: "Set",
			Key:    k,
			Val:    v,
			ErrMsg: "key or value is empty",
		}
	}

	return v, nil
}

// SetInBatch creates multiple new entries or updates multiple existed entries given by a slice of RedisObj
func (c *RedisCallers) SetInBatch(objs []RedisObj) (err error) {
	if len(objs) > 0 {
		var KVList []string
		for _, obj := range objs {
			KVList = append(KVList, obj.K, obj.V)
		}

		statusCmd := c.Client.MSet(KVList)

		if statusCmd.Err() != nil {
			err = &RedisError{
				Action: "Set - In Batch",
				Key:    "*",
				Val:    "*",
				ErrMsg: fmt.Sprintf("error occurs when setting keys:[%#v]", objs),
			}
		}

		return
	}

	return &RedisError{
		Action: "Set",
		Key:    "*",
		Val:    "*",
		ErrMsg: "batch is empty",
	}
}

// Get gets an entry by its key.
func (c *RedisCallers) Get(k string) (v string, err error) {
	if validate(k) {
		v, err = c.Client.Get(k).Result()
		if err != nil {
			return v, &RedisError{
				Action: "Get Key",
				Key:    k,
				Val:    v,
				ErrMsg: err.Error(),
			}
		}
	} else {
		return v, &RedisError{
			Action: "Get",
			Key:    k,
			Val:    v,
			ErrMsg: "key is empty",
		}
	}

	return v, nil
}

// Search returns any entries that fulfills the given search patten and keywords. Pattern is used to check the key
// while keywords are used to check the values of the matched keys.
// For example, assume we have {'user:1', 'aaa'}, {'user:2', 'abc'}, {'user:3', 'bbb'} three entries,
// if the given pattern is 'user:*' and the keywords is ['a'],
// the returned entries are:  {'user:1', 'aaa'}, {'user:2', 'abc'}.
// * Using scan when keySpace is big
func (c *RedisCallers) SearchByScan(patten string, keywords []string, count int64) (objs []RedisObj, err error) {
	if validate(patten) {
		var keys []string
		scanCmd := c.Client.Scan(0, patten, count)
		err = scanCmd.Err()
		if err != nil {
			return objs, nil
		}
		iter := scanCmd.Iterator()

		for iter.Next() {
			keys = append(keys, iter.Val())
		}

		values := c.Client.MGet(keys...).Val()

		if len(values) == len(keys) {
			for idx, key := range keys {
				if len(keywords) > 0 {
					if match(values[idx].(string), keywords...) {
						obj := RedisObj{
							K: key,
							V: values[idx].(string),
						}

						objs = append(objs, obj)
					}
				} else {
					obj := RedisObj{
						K: key,
						V: values[idx].(string),
					}

					objs = append(objs, obj)
				}
			}
		} else {
			return objs, &RedisError{
				Action: "Search",
				Key:    "*",
				Val:    "*",
				ErrMsg: "Keys and values not match",
			}
		}

	} else {
		return objs, &RedisError{
			Action: "Search",
			Key:    "*",
			Val:    "*",
			ErrMsg: "pattern is empty",
		}
	}
	return objs, err
}

// * Using scan when keySpace is small
func (c *RedisCallers) SearchByKeys(patten string, keywords []string) (objs []RedisObj, err error) {
	if validate(patten) {
		keys := c.Client.Keys(patten).Val()
		values := c.Client.MGet(keys...).Val()

		if len(values) == len(keys) {
			for idx, key := range keys {
				if len(keywords) > 0 {
					if match(values[idx].(string), keywords...) {
						obj := RedisObj{
							K: key,
							V: values[idx].(string),
						}

						objs = append(objs, obj)
					}
				} else {
					obj := RedisObj{
						K: key,
						V: values[idx].(string),
					}

					objs = append(objs, obj)
				}
			}
		} else {
			return objs, &RedisError{
				Action: "Search",
				Key:    "*",
				Val:    "*",
				ErrMsg: "Keys and values not match",
			}
		}

	} else {
		return objs, &RedisError{
			Action: "Search",
			Key:    "*",
			Val:    "*",
			ErrMsg: "pattern is empty",
		}
	}
	return objs, err
}


// Del deletes entries by their keys
func (c *RedisCallers) Del(keys ...string) error {
	if validate(keys...) {
		err := c.Client.Del(keys...).Err()
		if err != nil {
			return &RedisError{
				Action: "Delete",
				Key:    joinKey(keys...),
				Val:    "",
				ErrMsg: err.Error(),
			}
		}
	} else {
		return &RedisError{
			Action: "Delete",
			Key:    "*",
			Val:    "*",
			ErrMsg: "key(s) are empty",
		}
	}

	return nil
}

// validate checks if inputs are empty
func validate(args ...string) (ok bool) {
	for _, arg := range args {
		a := strings.TrimSpace(arg)
		if len(a) == 0 {
			return false
		}
	}

	return true
}

// joinKey converts a collection of keys to a string
func joinKey(keys ...string) string {
	ks := ""
	for _, k := range keys {
		ks += k + " "
	}

	return ks
}

// match checks if a string contains any of the keyword in arg keywords
func match(str string, keywords ...string) bool {
	for _, k := range keywords {
		if str == k {
			return true
		}
	}

	return false
}
