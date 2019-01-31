package Dragonfly

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
)

/* ---  Error definition --- */
type PraiseOperationError struct {
	Operation		string
	CaseTemplateId	string
	UserId			string
	ErrMsg			string
}

func (e *PraiseOperationError) Error() string {
	return fmt.Sprintf("operation[%s] on [CaseTemplateId:%s, UserId:%s] %s",
		e.Operation, e.CaseTemplateId, e.UserId, e.ErrMsg)
}

var PATTERN_SETUP_ERR = errors.New("unable to setup pattern for key matching, userId is empty")

func (c *RedisCallers) SetPraise(caseTemplateId, userId string) (int, error) {
	status := 0

	key := toCasePraiseKey(caseTemplateId, userId)

	val, err := c.Client.Get(key).Int()
	if err != nil && err != redis.Nil {
		return 0, &PraiseOperationError{Operation: "Set Praise :: Validate Key", CaseTemplateId: caseTemplateId,
			UserId: userId, ErrMsg: err.Error()}
	}

	if val == status {
		status = 1
	}

	err = c.Client.Set(key, status, 0).Err()
	if err != nil {
		return 0, &PraiseOperationError{Operation: "Set Praise :: Set Key", CaseTemplateId: caseTemplateId,
			UserId: userId, ErrMsg: err.Error()}
	}

	return status, nil
}

func (c *RedisCallers) GetPraiseCount(caseTemplateId, userId string) (int, int, error) {

	// Setup key and search pattern given parameters
	pattern := toCasePraisePattern(caseTemplateId)
	if pattern == "<invalid>" {
		return 0, 0, PATTERN_SETUP_ERR
	}

	keys, err := c.Client.Keys(pattern).Result()
	if err != nil {
		return 0, 0, &PraiseOperationError{Operation: "Get Praise Count :: Search Pattern",
			CaseTemplateId: caseTemplateId, UserId: userId, ErrMsg: err.Error()}
	}

	count := len(keys)
	if count == 0 {
		return 0, 0, nil
	} else {
		hasPraiseNum := 0

		for _, value := range keys {
			v, err := c.Client.Get(value).Int()
			if err != nil && err != redis.Nil {
				return 0, hasPraiseNum, &PraiseOperationError{Operation: "Get Praise Count :: Get value by key",
					CaseTemplateId: caseTemplateId, UserId: userId, ErrMsg: err.Error()}
			}

			if v == 1 {
				hasPraiseNum ++
			}
		}

		key := toCasePraiseKey(caseTemplateId, userId)
		v, err := c.Client.Get(key).Int()
		if err != nil {
			if err == redis.Nil {
				return 0, hasPraiseNum, nil
			} else {
				return 0, hasPraiseNum, &PraiseOperationError{Operation: "Get Praise Count :: Get value by key",
				CaseTemplateId: caseTemplateId, UserId: userId, ErrMsg: err.Error()}
			}
		} else {
			return v, hasPraiseNum, nil
		}
	}
}

// Convert to a redis key
func toCasePraiseKey(caseTemplateId, userId string) string {
	return fmt.Sprintf("cp:%s:%s", caseTemplateId, userId)
}

// Convert to a search pattern
func toCasePraisePattern(caseTemplateId string) string {
	return fmt.Sprintf("cp:%s:*", caseTemplateId)
}