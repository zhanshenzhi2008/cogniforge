package modelcache

import (
	"sync"
	"time"
)

const defaultLocalTTL = 30 * time.Second

// localCache 进程内缓存，对应 Java Caffeine 的「单 key + TTL」。
// 一致性不靠 TTL，靠每次读取对比 Redis 的 rev；TTL 只在 Redis 不可用时兜底。
type localCache struct {
	mu      sync.RWMutex
	snap    *Snapshot
	plain   string
	headers map[string]string
	expire  time.Time
}

func (l *localCache) getIfRev(rev int64) (*Snapshot, string, map[string]string, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.snap == nil || l.snap.Rev != rev {
		return nil, "", nil, false
	}
	return l.snap, l.plain, l.headers, true
}

func (l *localCache) getIfFresh(now time.Time) (*Snapshot, string, map[string]string, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.snap == nil || now.After(l.expire) {
		return nil, "", nil, false
	}
	return l.snap, l.plain, l.headers, true
}

func (l *localCache) set(snap *Snapshot, plain string, headers map[string]string, ttl time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.snap = snap
	l.plain = plain
	l.headers = headers
	l.expire = time.Now().Add(ttl)
}

func (l *localCache) clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.snap = nil
	l.plain = ""
	l.headers = nil
	l.expire = time.Time{}
}
