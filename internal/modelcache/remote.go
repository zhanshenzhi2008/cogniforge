package modelcache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Remote 远端存储（Redis 或测试用内存）。
type Remote interface {
	GetRev(ctx context.Context) (int64, error)
	Load(ctx context.Context) (*Snapshot, error)
	Save(ctx context.Context, snap *Snapshot) error
	BumpRev(ctx context.Context) (int64, error)
}

type redisRemote struct {
	rdb *redis.Client
	ttl time.Duration
}

func newRedisRemote(rdb *redis.Client) *redisRemote {
	return &redisRemote{rdb: rdb, ttl: time.Duration(RedisTTLSec) * time.Second}
}

func (r *redisRemote) GetRev(ctx context.Context) (int64, error) {
	s, err := r.rdb.Get(ctx, KeyRev).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(s, 10, 64)
}

func (r *redisRemote) Load(ctx context.Context) (*Snapshot, error) {
	s, err := r.rdb.Get(ctx, KeySnapshot).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var snap Snapshot
	if err := json.Unmarshal([]byte(s), &snap); err != nil {
		return nil, err
	}
	return &snap, nil
}

func (r *redisRemote) Save(ctx context.Context, snap *Snapshot) error {
	if snap == nil {
		return fmt.Errorf("nil snapshot")
	}
	body, err := json.Marshal(snap)
	if err != nil {
		return err
	}
	pipe := r.rdb.TxPipeline()
	pipe.Set(ctx, KeyRev, strconv.FormatInt(snap.Rev, 10), r.ttl)
	pipe.Set(ctx, KeySnapshot, body, r.ttl)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *redisRemote) BumpRev(ctx context.Context) (int64, error) {
	n, err := r.rdb.Incr(ctx, KeyRev).Result()
	if err != nil {
		return 0, err
	}
	_ = r.rdb.Expire(ctx, KeyRev, r.ttl).Err()
	_ = r.rdb.Del(ctx, KeySnapshot).Err()
	return n, nil
}

// MemoryRemote 单测用，不连 Redis。
type MemoryRemote struct {
	rev  int64
	snap *Snapshot
}

func NewMemoryRemote() *MemoryRemote {
	return &MemoryRemote{}
}

func (m *MemoryRemote) GetRev(ctx context.Context) (int64, error) {
	return m.rev, nil
}

func (m *MemoryRemote) Load(ctx context.Context) (*Snapshot, error) {
	return m.snap, nil
}

func (m *MemoryRemote) Save(ctx context.Context, snap *Snapshot) error {
	cp := *snap
	m.snap = &cp
	m.rev = snap.Rev
	return nil
}

func (m *MemoryRemote) BumpRev(ctx context.Context) (int64, error) {
	m.rev++
	m.snap = nil
	return m.rev, nil
}
