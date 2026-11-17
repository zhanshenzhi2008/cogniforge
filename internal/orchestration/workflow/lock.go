package workflow

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const execLockPrefix = "cogniforge:workflow:exec:lock:"
const execLockTTL = 90 * time.Second

// ExecLocker 防双 resume（Redis SetNX；无 Redis 时进程内互斥）
type ExecLocker struct {
	rdb *redis.Client
	mu  sync.Mutex
	local map[string]time.Time
}

func NewExecLocker(rdb *redis.Client) *ExecLocker {
	return &ExecLocker{rdb: rdb, local: make(map[string]time.Time)}
}

// TryLock 成功返回 unlock 函数；失败表示已有人在 resume
func (l *ExecLocker) TryLock(executionID string) (unlock func(), ok bool) {
	if executionID == "" {
		return func() {}, false
	}
	key := execLockPrefix + executionID
	ctx := context.Background()

	if l.rdb != nil {
		set, err := l.rdb.SetNX(ctx, key, "1", execLockTTL).Result()
		if err != nil || !set {
			return func() {}, false
		}
		return func() { _ = l.rdb.Del(ctx, key).Err() }, true
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	if exp, exists := l.local[executionID]; exists && exp.After(now) {
		return func() {}, false
	}
	l.local[executionID] = now.Add(execLockTTL)
	return func() {
		l.mu.Lock()
		delete(l.local, executionID)
		l.mu.Unlock()
	}, true
}
