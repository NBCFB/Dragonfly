package Dragonfly

import (
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"strings"
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
	return fmt.Sprintf("operation[%s] on [UserId:%s, CorpId:%s, CaseId:%s] %s", e.Operation, e.UserId, e.CorpId,
		e.CaseId, e.ErrMsg)
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

	_, err := c.Do("SET", toCaseStatusKey(userId, corpId, caseId), status)
	if err != nil {
		return &StatusOperationError{Operation: "SET-STATUS", UserId: userId, CorpId: corpId, CaseId: caseId,
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
		key := toCaseStatusKey(cs.UserId, cs.CorpId, cs.CaseId)
		c.Send("SET", key, cs.Status)
	}
	_, err := c.Do("EXEC")
	if err != nil {
		return &StatusOperationError{Operation: "B-SET-STATUS", UserId: "*", CorpId: "*", CaseId: "*", ErrMsg: err.Error()}
	}

	return nil
}

// Delete case status record
func DeleteStatus(userId, corpId, caseId string) error {
	c := Pool.Get()
	defer c.Close()

	_, err := c.Do("DEL", toCaseStatusKey(userId, corpId, caseId))
	if err != nil {
		return &StatusOperationError{Operation: "DEL-STATUS", UserId: userId, CorpId: corpId, CaseId: caseId,
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
	pattern := toCaseStatusPattern(userId, corpId, caseId)
	if pattern == "<invalid>" {
		return nil, PATTERN_SETUP_ERR
	}

	iter := 0
	for {
		// Scan using MATCH and the pattern
		arr, err := redis.Values(c.Do("SCAN", iter, "MATCH", pattern))
		if err != nil {
			return nil, &StatusOperationError{Operation: "SEARCH-STATUS-SCAN", UserId: userId, CorpId: corpId,
				CaseId: caseId, ErrMsg: err.Error()}
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
						return css, &StatusOperationError{Operation: "SEARCH-STATUS-GET", UserId: userId,
							CorpId: corpId, CaseId: caseId, ErrMsg: err.Error()}
					}
				}

				keyItems := strings.Split(k, ":")

				// Store it
				cs := CaseStatus{
					UserId: keyItems[1],
					CorpId: keyItems[2],
					CaseId: keyItems[3],
					Status: v,
				}
				css = append(css, cs)
			}
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

	v, err := redis.Int(c.Do("GET", toCaseStatusKey(userId, corpId, caseId)))
	if err != nil {
		return -1, &StatusOperationError{Operation: "GET-STATUS", UserId: userId, CorpId: corpId, CaseId: caseId,
				ErrMsg: err.Error()}
	}

	return v, nil
}

// Convert to a redis key
func toCaseStatusKey(userId, corpId, caseId string) string {
	return fmt.Sprintf("user:%s:%s:%s", userId, corpId, caseId)
}

// Convert to a search pattern
func toCaseStatusPattern(userId, corpId, caseId string) string {
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
