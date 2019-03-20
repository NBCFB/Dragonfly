package Dragonfly

import (
	"fmt"
	"github.com/go-redis/redis"
	"strings"
)

/* ---  Error definition --- */
type SummaryOperationError struct {
	Operation string
	CorpId    string
	CaseId    string
	UserId    string
	ErrMsg    string
}

func (e *SummaryOperationError) Error() string {
	return fmt.Sprintf("operation[%s] on [UserId:%s, CorpId:%s, CaseId:%s] %s",
		e.Operation, e.UserId, e.CorpId, e.CaseId, e.ErrMsg)
}

func (c *RedisCallers) SetSummary(userId, corpId, caseId string, val int) error {
	key := toCaseSummaryKey(userId, corpId, caseId)

	if err := c.Client.Set(key, val, 0).Err(); err != nil {
		return &SummaryOperationError{Operation: "Set Summary :: Validate Key", UserId: userId,
			CorpId: corpId, CaseId: caseId, ErrMsg: err.Error()}
	}

	return nil
}

func (c *RedisCallers) DeleteSummary(userId, corpId, caseId string) error {
	key := toCaseSummaryKey(userId, corpId, caseId)
	err := c.Client.Del(key).Err()
	if err != nil {
		return &SummaryOperationError{Operation: "Del Summary ", UserId: userId,
			CorpId: corpId, CaseId: caseId, ErrMsg: err.Error()}
	}

	return nil
}

func (c *RedisCallers) GetSummaryCount(userId, corpId string) (int, error) {
	pattern := toCaseSummaryPattern(userId, corpId)
	if pattern == "<invalid>" {
		return 0, PATTERN_SETUP_ERR
	}

	keys, err := c.Client.Keys(pattern).Result()
	if err != nil {
		return 0, &SummaryOperationError{Operation: "Get Summary Count:: Search Pattern", UserId: userId,
			CorpId: corpId, CaseId: "*", ErrMsg: err.Error()}
	}
	return len(keys), nil
}

func (c *RedisCallers) HasSummaryKey(userId, corpId, caseId string) (bool, error) {
	key := toCaseSummaryKey(userId, corpId, caseId)
	_, err := c.Client.Get(key).Int()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, &SummaryOperationError{Operation: "Get Summary", UserId: userId,
			CorpId: corpId, CaseId: caseId, ErrMsg: err.Error()}
	}

	return true, nil
}

func (c *RedisCallers) GetSummaryIds(userId, corpId string) ([]string, error) {
	var ids []string
	pattern := toCaseSummaryPattern(userId, corpId)
	if pattern == "<invalid>" {
		return ids, PATTERN_SETUP_ERR
	}

	keys, err := c.Client.Keys(pattern).Result()
	if err != nil {
		return ids, &SummaryOperationError{Operation: "Get Summary Id:: Search Pattern", UserId: userId,
			CorpId: corpId, CaseId: "*", ErrMsg: err.Error()}
	}

	for _, key := range keys {
		id := strings.Split(key, ":")[strings.LastIndex(key, ":")+1]
		ids = append(ids, id)
	}

	return ids, nil
}

func (c RedisCallers) DeleteSummaryKeys(userId, corpId string) error {
	pattern := toCaseSummaryPattern(userId, corpId)
	if pattern == "<invalid>" {
		return PATTERN_SETUP_ERR
	}

	keys, err := c.Client.Keys(pattern).Result()
	if err != nil {
		return &SummaryOperationError{Operation: "Del Summary :: Search Pattern", UserId: userId,
			CorpId: corpId, CaseId: "*", ErrMsg: err.Error()}
	}

	if len(keys) > 0 {
		err := c.Client.Del(keys...).Err()
		if err != nil {
			return &SummaryOperationError{Operation: "Del Summary :: Search Pattern", UserId: userId,
				CorpId: corpId, CaseId: "*", ErrMsg: err.Error()}
		}
	}

	return nil
}

// Convert to a redis key
func toCaseSummaryKey(userId, corpId, caseId string) string {
	return fmt.Sprintf("cs:%s:%s:%s", userId, corpId, caseId)
}

// Convert to a search pattern
func toCaseSummaryPattern(userId, corpId string) string {
	return fmt.Sprintf("cs:%s:%s:*", userId, corpId)
}
