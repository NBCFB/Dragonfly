package case_status

import (
	"fmt"
	db "github.com/NBCFB/Dragonfly/helper"
	"github.com/gomodule/redigo/redis"
)

type StatusOperationError struct {
	Operation	string
	UserId		string
	CorpId		string
	CaseId		string
	ErrMsg		string
}

func (e *StatusOperationError) Error() string {
	return fmt.Sprintf("operation[%s] on [UserId:%s, CorpId:%s, CaseId:%s] %s", e.UserId, e.CorpId, e.CaseId,
		e.ErrMsg)
}

func SetStatus(cs CaseStatus) error {
	c := db.Pool.Get()
	defer c.Close()

	_, err := c.Do("SET", toKey(cs.UserId, cs.CorpId, cs.CaseId), cs.Status)
	if err != nil {
		return &StatusOperationError{Operation: "SET", UserId: cs.UserId, CorpId: cs.CorpId, CaseId: cs.CaseId,
			ErrMsg: err.Error()}
	}

	return nil
}

func BatchSetStatus(css []CaseStatus) error {
	c := db.Pool.Get()
	defer c.Close()

	c.Send("MULTI")
	for _, cs := range(css) {
		key := toKey(cs.UserId, cs.CorpId, cs.CaseId)
		c.Send("DEL", key)
	}
	_, err := c.Do("EXEC")
	if err != nil {
		return &StatusOperationError{Operation: "B-SET", UserId: "*", CorpId: "*", CaseId: "*", ErrMsg: err.Error()}
	}

	return nil
}

func DeleteStatus(cs CaseStatus) error {
	c := db.Pool.Get()
	defer c.Close()

	_, err := c.Do("DEL", toKey(cs.UserId, cs.CorpId, cs.CaseId))
	if err != nil {
		return &StatusOperationError{Operation: "DEL", UserId: cs.UserId, CorpId: cs.CorpId, CaseId: cs.CaseId,
			ErrMsg: err.Error()}
	}

	return nil
}

func GetStatusByMatch(cs CaseStatus) ([]int, error) {
	c := db.Pool.Get()
	defer c.Close()

	vals := []int{}
	iter := 0
	pattern := toPattern(cs.UserId, cs.CorpId, cs.CaseId)
	for {
		arr, err := redis.Values(c.Do("SCAN", iter, "MATCH", pattern))
		if err != nil {
			return nil, &StatusOperationError{Operation: "SEARCH", UserId: cs.UserId, CorpId: cs.CorpId,
				CaseId: cs.CaseId, ErrMsg: err.Error()}
		}

		iter, _ = redis.Int(arr[0], nil)
		k, _ := redis.Strings(arr[1], nil)
		v, err := redis.Int(c.Do("GET", k))
		if err != nil {
			return vals, &StatusOperationError{Operation: "SEARCH", UserId: cs.UserId, CorpId: cs.CorpId,
				CaseId: cs.CaseId, ErrMsg: err.Error()}
		}
		vals = append(vals, v)

		if iter == 0 {
			break
		}
	}

	return vals, nil
}

func GetStatusByKey(cs CaseStatus) (int, error) {
	c := db.Pool.Get()
	defer c.Close()

	v, err := redis.Int(c.Do("GET", toKey(cs.UserId, cs.CorpId, cs.CaseId)))
	if err != nil {
		return -1, &StatusOperationError{Operation: "GET", UserId: cs.UserId, CorpId: cs.CorpId,
			CaseId: cs.CaseId, ErrMsg: err.Error()}
	}

	return v, nil
}

func toKey(userId, corpId, caseId string) string {
	return fmt.Sprintf("user:%s:%s:%s", userId, corpId, caseId)
}

func toPattern(userId, corpId, caseId string) string {
	if userId == "" {
		return "<invalid>"
	} else {
		if corpId == "" {
			return fmt.Sprintf("user:%s:*", userId)
		} else {
			if caseId == "" {
				return fmt.Sprintf("user:%s:%s:*", userId, corpId)
			} else {
				return fmt.Sprintf("user:%s:%s:%s", userId, corpId, caseId)
			}
		}
	}
}
