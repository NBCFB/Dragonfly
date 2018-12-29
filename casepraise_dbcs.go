package Dragonfly

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
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

func SetPraise(caseTemplateId, userId string) (int, error) {
	newV := -1
	c := Pool.Get()
	defer c.Close()

	key := toCasePraiseKey(caseTemplateId, userId)
	fmt.Println(key)

	v, err := redis.Int(c.Do("GET", key))
	if err != nil {
		if err != redis.ErrNil {
			return newV, &PraiseOperationError{Operation: "GET-PRAISE", CaseTemplateId: caseTemplateId,
				UserId: userId, ErrMsg: err.Error()}
		}
	}

	if v == 1 {
		newV = 0
	} else if v == 0 {
		newV = 1
	}

	_, err = c.Do("SET", key, newV)
	if err != nil {
		return v, &PraiseOperationError{Operation: "SET-PRAISE", CaseTemplateId: caseTemplateId,
			UserId: userId, ErrMsg: err.Error()}
	}

	return newV, nil
}

func GetPraiseCount(caseTemplateId, userId string) (int, int, error) {
	count := 0
	currentV := -1

	c := Pool.Get()
	defer c.Close()

	// Setup key and search pattern given parameters
	key := toCasePraiseKey(caseTemplateId, userId)
	pattern := toCasePraisePattern(caseTemplateId)
	if pattern == "<invalid>" {
		return currentV, count, PATTERN_SETUP_ERR
	}

	iter := 0
	for {
		// Scan using MATCH and the pattern
		arr, err := redis.Values(c.Do("SCAN", iter, "MATCH", pattern))
		if err != nil {
			if err != redis.ErrNil {
				return currentV, count, &PraiseOperationError{Operation: "SEARCH-PRAISE-SCAN",
					CaseTemplateId: caseTemplateId, UserId: "*", ErrMsg: err.Error()}
			}
		}

		if arr != nil && len(arr) > 0 {
			// Get a matched key
			iter, _ = redis.Int(arr[0], nil)
			ks, _ := redis.Strings(arr[1], nil)

			for _, k := range(ks) {
				// Get value based on the obtained key
				v, err := redis.Int(c.Do("GET", k))
				if err != nil {
					if err != redis.ErrNil {
						return currentV, count, &PraiseOperationError{Operation: "SEARCH-STATUS-GET",
							CaseTemplateId: caseTemplateId, UserId: userId, ErrMsg: err.Error()}
					}
				}

				if v == 1 {
					count  += 1
				}
			}

			if iter == 0 {
				break
			}
		}
	}

	currentV, err := redis.Int(c.Do("GET", key))
	if err != nil {
		if err != redis.ErrNil {
			return currentV, count, &PraiseOperationError{Operation: "GET-PRAISE", CaseTemplateId: caseTemplateId,
				UserId: userId, ErrMsg: err.Error()}
		}
	}

	return currentV, count, nil
}

// Convert to a redis key
func toCasePraiseKey(caseTemplateId, userId string) string {
	return fmt.Sprintf("cp:%s:%s", caseTemplateId, userId)
}

// Convert to a search pattern
func toCasePraisePattern(caseTemplateId string) string {
	return fmt.Sprintf("cp:%s:*", caseTemplateId)
}
