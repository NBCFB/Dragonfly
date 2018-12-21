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

func AddPraise(caseId, caseTemplateId, userId string) (bool, error) {
	c := Pool.Get()
	defer c.Close()

	key := toCasePraiseKey(caseId, caseTemplateId, userId)

	exist, err := redis.Bool(c.Do("EXISTS", key))
	if err != nil {
		return false, &PraiseOperationError{Operation: "CHECK-EXISTENCE", CaseId: caseId, CaseTemplateId: caseTemplateId,
			UserId: userId, ErrMsg: err.Error()}
	}

	if exist {
		return false, nil
	}

	_, err = c.Do("SET", key, 1)
	if err != nil {
		return false, &PraiseOperationError{Operation: "SET-PRAISE", CaseId: caseId, CaseTemplateId: caseTemplateId,
			UserId: userId, ErrMsg: err.Error()}
	}

	return true, nil
}

func GetPraiseCount(caseId, caseTemplateId string) (int, error) {
	count := 0

	c := Pool.Get()
	defer c.Close()

	// Setup search pattern given parameters
	pattern := toCasePraisePattern(caseId, caseTemplateId)
	if pattern == "<invalid>" {
		return count, PATTERN_SETUP_ERR
	}

	iter := 0
	for {
		// Scan using MATCH and the pattern
		arr, err := redis.Values(c.Do("SCAN", iter, "MATCH", pattern))
		if err != nil {
			return count, &PraiseOperationError{Operation: "SEARCH-PRAISE-SCAN", CaseId: caseId,
				CaseTemplateId: caseTemplateId, UserId: "*", ErrMsg: err.Error()}
		}

		iter, _ = redis.Int(arr[0], nil)
		count += 1

		if iter == 0 {
			break
		}
	}

	return count, nil
}

// Convert to a redis key
func toCasePraiseKey(caseId, caseTemplateId, userId string) string {
	return fmt.Sprintf("case:%s:%s:%s", caseId, caseTemplateId, userId)
}

// Convert to a search pattern
func toCasePraisePattern(caseId, caseTemplateId string) string {
	return fmt.Sprintf("case:%s:%s:*", caseId, caseTemplateId)
}
