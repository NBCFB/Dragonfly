package Dragonfly

import (
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
)

/* ---  Error definition --- */
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

var PATTERN_SETUP_ERR = errors.New("unable to setup pattern for key matching, userId is empty")

/* --- Model definition --- */
type CaseStatus struct {
	UserId		string	`json:"userId"`
	CorpId		string	`json:"corpId"`
	CaseId		string	`json:"caseId"`
	Status		int		`json:"status"`
}

/* --- Function definitions --- */
// Update status for a case
func SetStatus(userId, corpId, caseId string, status int) error {
	c := Pool.Get()
	defer c.Close()

	_, err := c.Do("SET", toKey(userId, corpId, caseId), status)
	if err != nil {
		return &StatusOperationError{Operation: "SET", UserId: userId, CorpId: corpId, CaseId: caseId,
			ErrMsg: err.Error()}
	}

	return nil
}

// Update status for multiple cases
func BatchSetStatus(css []CaseStatus) error {
	c := Pool.Get()
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

// Delete case status record
func DeleteStatus(userId, corpId, caseId string) error {
	c := Pool.Get()
	defer c.Close()

	_, err := c.Do("DEL", toKey(userId, corpId, caseId))
	if err != nil {
		return &StatusOperationError{Operation: "DEL", UserId: userId, CorpId: corpId, CaseId: caseId,
			ErrMsg: err.Error()}
	}

	return nil
}

// Get status of case(s) based on search pattern
func GetStatusByMatch(userId, corpId, caseId string) ([]CaseStatus, error) {
	var css []CaseStatus
	c := Pool.Get()
	defer c.Close()

	// Setup search pattern given parameters
	pattern := toPattern(userId, corpId, caseId)
	if pattern == "<invalid>" {
		return nil, PATTERN_SETUP_ERR
	}

	iter := 0
	for {
		// Scan using MATCH and the pattern
		arr, err := redis.Values(c.Do("SCAN", iter, "MATCH", pattern))
		if err != nil {
			return nil, &StatusOperationError{Operation: "SEARCH", UserId: userId, CorpId: corpId,
				CaseId: caseId, ErrMsg: err.Error()}
		}

		if arr != nil && len(arr) > 0 {
			// Get a matched key
			iter, _ = redis.Int(arr[0], nil)
			k, _ := redis.Strings(arr[1], nil)

			// Get value based on the obtained key
			v, err := redis.Int(c.Do("GET", k))
			if err != nil {
				return css, &StatusOperationError{Operation: "SEARCH", UserId: userId, CorpId: corpId, CaseId: caseId,
					ErrMsg: err.Error()}
			}

			// Store it
			cs := CaseStatus{
				UserId: userId,
				CorpId: corpId,
				CaseId: caseId,
				Status: v,
			}
			css = append(css, cs)
		}

		if iter == 0 {
			break
		}
	}

	return css, nil
}

// Get status of a single case
func GetStatusByKey(userId, corpId, caseId string) (int, error) {
	c := Pool.Get()
	defer c.Close()

	v, err := redis.Int(c.Do("GET", toKey(userId, corpId, caseId)))
	if err != nil {
		return -1, &StatusOperationError{Operation: "GET", UserId: userId, CorpId: corpId, CaseId: caseId,
			ErrMsg: err.Error()}
	}

	return v, nil
}

// Convert to a redis key
func toKey(userId, corpId, caseId string) string {
	return fmt.Sprintf("user:%s:%s:%s", userId, corpId, caseId)
}

// Convert to a search pattern
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
