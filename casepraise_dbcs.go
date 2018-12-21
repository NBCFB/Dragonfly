package Dragonfly

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
)

/* ---  Error definition --- */
type PraiseOperationError struct {
	Operation		string
	CaseId			string
	CaseTemplateId	string
	UserId			string
	ErrMsg			string
}

func (e *PraiseOperationError) Error() string {
	return fmt.Sprintf("operation[%s] on [CaseId:%s, CaseTemplateId:%s, UserId:%s] %s",
		e.Operation, e.CaseId, e.CaseTemplateId, e.UserId, e.ErrMsg)
}

func SetPraise(caseId, caseTemplateId, userId string) (int, error) {
	newV := -1
	c := Pool.Get()
	defer c.Close()

	key := toCasePraiseKey(caseId, caseTemplateId, userId)

	v, err := redis.Int(c.Do("GET", key))
	if err != nil {
		return newV, &PraiseOperationError{Operation: "GET-PRAISE", CaseId: caseId, CaseTemplateId: caseTemplateId,
			UserId: userId, ErrMsg: err.Error()}
	}

	if v == 1 {
		newV = 0
	} else if v == 0 {
		newV = 1
	}

	_, err = c.Do("SET", key, newV)
	if err != nil {
		if err != redis.ErrNil {
			return v, &PraiseOperationError{Operation: "SET-PRAISE", CaseId: caseId, CaseTemplateId: caseTemplateId,
				UserId: userId, ErrMsg: err.Error()}
		}
	}

	return newV, nil
}

func GetPraiseCount(caseId, caseTemplateId, userId string) (int, int, error) {
	count := 0
	currentV := -1

	c := Pool.Get()
	defer c.Close()

	// Setup key and search pattern given parameters
	key := toCasePraiseKey(caseId, caseTemplateId, userId)
	pattern := toCasePraisePattern(caseId, caseTemplateId)
	if pattern == "<invalid>" {
		return currentV, count, PATTERN_SETUP_ERR
	}

	iter := 0
	for {
		// Scan using MATCH and the pattern
		arr, err := redis.Values(c.Do("SCAN", iter, "MATCH", pattern))
		if err != nil {
			if err != redis.ErrNil {
				return currentV, count, &PraiseOperationError{Operation: "SEARCH-PRAISE-SCAN", CaseId: caseId,
					CaseTemplateId: caseTemplateId, UserId: "*", ErrMsg: err.Error()}
			}
		}

		iter, _ = redis.Int(arr[0], nil)
		count += 1

		if iter == 0 {
			break
		}
	}

	currentV, err := redis.Int(c.Do("GET", key))
	if err != nil {
		fmt.Println(err == redis.ErrNil)
		if err != redis.ErrNil {
			return currentV, count, &PraiseOperationError{Operation: "GET-PRAISE", CaseId: caseId,
				CaseTemplateId: caseTemplateId, UserId: userId, ErrMsg: err.Error()}
		}
	}

	return currentV, count, nil
}

// Convert to a redis key
func toCasePraiseKey(caseId, caseTemplateId, userId string) string {
	return fmt.Sprintf("case:%s:%s:%s", caseId, caseTemplateId, userId)
}

// Convert to a search pattern
func toCasePraisePattern(caseId, caseTemplateId string) string {
	return fmt.Sprintf("case:%s:%s:*", caseId, caseTemplateId)
}
