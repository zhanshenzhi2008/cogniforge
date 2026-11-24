package team

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	runKeyPrefix = "cogniforge:team:run:"
	runTTL       = 24 * time.Hour
)

// RunStore Redis 持久化小队 run_state（预览续聊；非 workflow checkpoint）
type RunStore struct {
	rdb *redis.Client
}

func NewRunStore(rdb *redis.Client) *RunStore {
	if rdb == nil {
		return nil
	}
	return &RunStore{rdb: rdb}
}

func runKey(userID, teamID, sessionID string) string {
	u := strings.TrimSpace(userID)
	t := strings.TrimSpace(teamID)
	s := strings.TrimSpace(sessionID)
	if s == "" {
		s = "default"
	}
	return runKeyPrefix + u + ":" + t + ":" + s
}

// Save 写入 run_state
func (s *RunStore) Save(ctx context.Context, userID, teamID, sessionID string, state *RunState) error {
	if s == nil || s.rdb == nil || state == nil {
		return nil
	}
	b, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal run_state: %w", err)
	}
	return s.rdb.Set(ctx, runKey(userID, teamID, sessionID), b, runTTL).Err()
}

// Load 读取；不存在返回 nil,nil
func (s *RunStore) Load(ctx context.Context, userID, teamID, sessionID string) (*RunState, error) {
	if s == nil || s.rdb == nil {
		return nil, nil
	}
	raw, err := s.rdb.Get(ctx, runKey(userID, teamID, sessionID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var st RunState
	if err := json.Unmarshal(raw, &st); err != nil {
		return nil, fmt.Errorf("unmarshal run_state: %w", err)
	}
	return &st, nil
}

// Delete 清除会话 run
func (s *RunStore) Delete(ctx context.Context, userID, teamID, sessionID string) error {
	if s == nil || s.rdb == nil {
		return nil
	}
	return s.rdb.Del(ctx, runKey(userID, teamID, sessionID)).Err()
}
