package Dragonfly

import (
	"fmt"
	"strings"
	"time"
)

type RedisError struct {
	Action		string
	Key			string
	Val			string
	ErrMsg		string
}

type RedisObj struct {
	K string
	V string
}

func (e *RedisError) Error() string {
	return fmt.Sprintf("Action[%s] Key=%s, Val=%s, %s", e.Action, e.Key, e.Val, e.ErrMsg)
}

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

func (c *RedisCallers) SetInBatch(objs []RedisObj) (err error) {
	if len(objs) > 0 {
		var errKeys []string
		for _, obj := range objs {
			_, e := c.Set(obj.K, obj.V, 0)
			if e != nil {
				errKeys = append(errKeys, obj.K)
			}
		}

		if len(errKeys) > 0 {
			err = &RedisError{
				Action: "Set - In Batch",
				Key:    "*",
				Val:    "*",
				ErrMsg: fmt.Sprintf("error occurs when setting keys:[%v]", errKeys),
			}
		}

		return err

	} else {
		return &RedisError{
			Action: "Set",
			Key:    "*",
			Val:    "*",
			ErrMsg: "batch is empty",
		}
	}
}

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

func (c *RedisCallers) Search(patten string, keywords []string) (objs []RedisObj, err error) {
	if validate(patten) {
		scanCmd := c.Client.Scan(0, patten, -1)
		err = scanCmd.Err()
		if err != nil {
			return objs, nil
		}

		iter := scanCmd.Iterator()

		var errKeys []string
		for iter.Next() {
			k := iter.Val()
			v, e := c.Get(k)

			if e == nil {
				if len(keywords) > 0 {
					if match(v, keywords...) {
						obj := RedisObj{
							K: k,
							V: v,
						}

						objs = append(objs, obj)
					}
				} else {
					obj := RedisObj{
						K: k,
						V: v,
					}

					objs = append(objs, obj)
				}
			} else {
				errKeys = append(errKeys, k)
			}
		}

		if len(errKeys) > 0 {
			err = &RedisError{
				Action: "Search - Iterate",
				Key:    "*",
				Val:    "*",
				ErrMsg: fmt.Sprintf("error occurs when getting keys:[%v]", errKeys),
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

func validate(args ...string) (ok bool) {
	for _, arg := range args {
		a := strings.TrimSpace(arg)
		if len(a) == 0 {
			return false
		}
	}

	return true
}

func joinKey(keys ...string) string {
	ks := ""
	for _, k := range keys {
		ks += k + " "
	}

	return ks
}

func match(str string, keywords ...string) bool {
	for _, k := range keywords {
		if str == k {
			return true
		}
	}

	return false
}