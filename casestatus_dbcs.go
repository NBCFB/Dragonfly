package Dragonfly

import (
	"fmt"
	"github.com/go-redis/redis"
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

/* --- Model definition --- */
type CaseStatus struct {
	UserId		string	`json:"userId"`
	CorpId		string	`json:"corpId"`
	CaseId		string	`json:"caseId"`
	Status		int		`json:"status"`
}

func (c *RedisCallers) SetStatus(userId, corpId, caseId string, val int) error {

	key := toCaseStatusKey(userId, corpId, caseId)
	if _, err := c.Client.Get(key).Result(); err != nil && err != redis.Nil {
		return &StatusOperationError{Operation: "Set Status :: Validate Key", UserId: userId, CorpId: corpId,
			CaseId: caseId, ErrMsg: err.Error()}
	}

	err := c.Client.Set(key, val, 0).Err()
	if err != nil {
		return &StatusOperationError{Operation: "Set Status :: Set Key", UserId: userId, CorpId: corpId,
			CaseId: caseId, ErrMsg: err.Error()}
	}

	return nil
}

func (c *RedisCallers) DeleteStatus(cases []CaseStatus) error {

	keys := []string{}
	for _, c := range(cases) {
		keys = append(keys, toCaseStatusKey(c.UserId, c.CorpId, c.CaseId))
	}

	if len(keys) > 0 {
		err := c.Client.Del(keys...).Err()
		if err != nil {
			return &StatusOperationError{Operation: "Set Status :: Del Keys", UserId: "*", CorpId: "*",
				CaseId: "*", ErrMsg: err.Error()}
		}
	}

	return nil
}

func (c *RedisCallers) HasUnreadStatus(userId, corpId string) (bool, error) {

	// Setup key and search pattern given parameters
	pattern := toCaseStatusPattern(userId, corpId, "")
	if pattern == "<invalid>" {
		return false, PATTERN_SETUP_ERR
	}

	keys, err := c.Client.Keys(pattern).Result()
	if err != nil {
		return false, &StatusOperationError{Operation: "Check unread status :: Set Key", UserId: userId, CorpId: corpId,
			CaseId: "-", ErrMsg: err.Error()}
	}

	count := len(keys)
	if count == 0 {
		return false, nil
	} else {
		for _, k :=  range(keys) {
			v, _ := c.Client.Get(k).Int()
			if v == 0 {
				return true, nil
			}
		}

		return false, nil
	}
}


func (c *RedisCallers) GetStatus(userId, corpId, caseId string) (int, error) {

	key := toCaseStatusKey(userId, corpId, caseId)
	v, err := c.Client.Get(key).Int()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		} else {
			return 0, &StatusOperationError{Operation: "Set Status :: Validate Key", UserId: userId, CorpId: corpId,
				CaseId: caseId, ErrMsg: err.Error()}
		}
	}

	return v, nil
}

// Make sure user:corpid has status=0

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