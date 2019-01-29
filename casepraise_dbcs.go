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

func SetPraise(caseTemplateId, userId string, val int) (int, error) {
	client, err := NewConnection()
	if err != nil {
		return 0, err
	}

	key := toCasePraiseKey(caseTemplateId, userId)
	if _, err := client.Get(key).Result(); err != nil && err != redis.Nil {
		return 0, &PraiseOperationError{Operation: "Set Praise :: Validate Key", CaseTemplateId: caseTemplateId,
			UserId: userId, ErrMsg: err.Error()}
	}

	err = client.Set(key, val, 0).Err()
	if err != nil {
		return 0, &PraiseOperationError{Operation: "Set Praise :: Set Key", CaseTemplateId: caseTemplateId,
			UserId: userId, ErrMsg: err.Error()}
	}

	return val, nil
}

func GetPraiseCount(caseTemplateId, userId string) (int, int, error) {
	client, err := NewConnection()
	if err != nil {
		return 0, 0, err
	}

	// Setup key and search pattern given parameters
	pattern := toCasePraisePattern(caseTemplateId)
	if pattern == "<invalid>" {
		return 0, 0, PATTERN_SETUP_ERR
	}

	keys, err := client.Keys(pattern).Result()
	if err != nil {
		return 0, 0, &PraiseOperationError{Operation: "Get Praise Count :: Search Pattern",
			CaseTemplateId: caseTemplateId, UserId: userId, ErrMsg: err.Error()}
	}

	count := len(keys)
	if count == 0 {
		return 0, 0, nil
	} else {
		key := toCasePraiseKey(caseTemplateId, userId)
		v, err := client.Get(key).Int()
		if err != nil {
			if err != redis.Nil {
				return 0, count, nil
			} else {
				return 0, count, &PraiseOperationError{Operation: "Get Praise Count :: Get value by key",
				CaseTemplateId: caseTemplateId, UserId: userId, ErrMsg: err.Error()}
			}
		} else {
			return v, count, nil
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