package quota

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Store 配额计数。生产用 Redis；测试用内存。
type Store interface {
	Incr(ctx context.Context, key string, ttl time.Duration) (int64, error)
	IncrBy(ctx context.Context, key string, delta int64, ttl time.Duration) (int64, error)
	Get(ctx context.Context, key string) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)
}

type RedisStore struct {
	rdb *redis.Client
}

func NewRedisStore(rdb *redis.Client) *RedisStore {
	return &RedisStore{rdb: rdb}
}

func (s *RedisStore) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	return s.IncrBy(ctx, key, 1, ttl)
}

func (s *RedisStore) IncrBy(ctx context.Context, key string, delta int64, ttl time.Duration) (int64, error) {
	n, err := s.rdb.IncrBy(ctx, key, delta).Result()
	if err != nil {
		return 0, err
	}
	if n == delta && ttl > 0 {
		_ = s.rdb.Expire(ctx, key, ttl).Err()
	}
	return n, nil
}

func (s *RedisStore) Get(ctx context.Context, key string) (int64, error) {
	n, err := s.rdb.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return n, err
}

func (s *RedisStore) Decr(ctx context.Context, key string) (int64, error) {
	return s.rdb.Decr(ctx, key).Result()
}

type memEntry struct {
	n       int64
	expires time.Time
}

// MemoryStore 单测 / 无 Redis 时不要用于生产（生产 fail-closed）。
type MemoryStore struct {
	mu   sync.Mutex
	data map[string]memEntry
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{data: make(map[string]memEntry)}
}

func (s *MemoryStore) purgeLocked(now time.Time) {
	for k, v := range s.data {
		if !v.expires.IsZero() && now.After(v.expires) {
			delete(s.data, k)
		}
	}
}

func (s *MemoryStore) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	return s.IncrBy(ctx, key, 1, ttl)
}

func (s *MemoryStore) IncrBy(_ context.Context, key string, delta int64, ttl time.Duration) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	s.purgeLocked(now)
	e := s.data[key]
	e.n += delta
	if e.expires.IsZero() && ttl > 0 {
		e.expires = now.Add(ttl)
	}
	s.data[key] = e
	return e.n, nil
}

func (s *MemoryStore) Get(_ context.Context, key string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	s.purgeLocked(now)
	return s.data[key].n, nil
}

func (s *MemoryStore) Decr(_ context.Context, key string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e := s.data[key]
	e.n--
	s.data[key] = e
	return e.n, nil
}
