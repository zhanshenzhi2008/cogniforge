package modelcache

import (
	"context"
	"log/slog"
	"time"
)

// Cache 两级缓存：本地（Caffeine 同类）+ Redis。
// 一致性：写配置时 BumpRev；读时用 Redis rev 校验本地，对不上就丢本地。
type Cache struct {
	local  localCache
	remote Remote
}

func New(remote Remote) *Cache {
	return &Cache{remote: remote}
}

func (c *Cache) enabled() bool {
	return c != nil
}

func redisCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 400*time.Millisecond)
}

// Get 返回快照。ok=false 表示两级都没有，调用方应查库再 Put。
func (c *Cache) Get() (snap *Snapshot, ok bool) {
	if !c.enabled() {
		return nil, false
	}

	ctx, cancel := redisCtx()
	defer cancel()

	if c.remote != nil {
		rev, err := c.remote.GetRev(ctx)
		if err != nil {
			slog.Warn("model cache redis rev failed, fallback local TTL", "error", err)
			snap, _, _, ok = c.local.getIfFresh(time.Now())
			return snap, ok
		}
		if snap, _, _, ok = c.local.getIfRev(rev); ok {
			return snap, true
		}
		loaded, err := c.remote.Load(ctx)
		if err != nil {
			slog.Warn("model cache redis load failed", "error", err)
			snap, _, _, ok = c.local.getIfFresh(time.Now())
			return snap, ok
		}
		if loaded != nil && loaded.ID != "" && (rev == 0 || loaded.Rev == rev) {
			c.local.set(loaded, "", nil, defaultLocalTTL)
			return loaded, true
		}
		return nil, false
	}

	snap, _, _, ok = c.local.getIfFresh(time.Now())
	return snap, ok
}

// GetHot 本地命中且 rev 一致时带上已解密的 Key，避免每次 AES。
func (c *Cache) GetHot() (snap *Snapshot, plain string, headers map[string]string, ok bool) {
	if !c.enabled() {
		return nil, "", nil, false
	}
	ctx, cancel := redisCtx()
	defer cancel()

	if c.remote != nil {
		rev, err := c.remote.GetRev(ctx)
		if err != nil {
			return c.local.getIfFresh(time.Now())
		}
		if snap, plain, headers, ok = c.local.getIfRev(rev); ok && plain != "" {
			return snap, plain, headers, true
		}
		loaded, err := c.remote.Load(ctx)
		if err != nil || loaded == nil || loaded.ID == "" {
			return nil, "", nil, false
		}
		if rev != 0 && loaded.Rev != rev {
			return nil, "", nil, false
		}
		c.local.set(loaded, "", nil, defaultLocalTTL)
		return loaded, "", nil, true
	}
	return c.local.getIfFresh(time.Now())
}

func (c *Cache) Put(snap *Snapshot) {
	c.PutHot(snap, "", nil)
}

func (c *Cache) PutHot(snap *Snapshot, plain string, headers map[string]string) {
	if !c.enabled() || snap == nil {
		return
	}
	c.local.set(snap, plain, headers, defaultLocalTTL)

	if c.remote == nil {
		return
	}
	ctx, cancel := redisCtx()
	defer cancel()
	if err := c.remote.Save(ctx, snap); err != nil {
		slog.Warn("model cache redis save failed", "error", err)
	}
}

// Invalidate 配置变更后调用：rev+1、删 Redis snapshot、清空本地。
func (c *Cache) Invalidate() {
	if !c.enabled() {
		return
	}
	c.local.clear()
	if c.remote == nil {
		return
	}
	ctx, cancel := redisCtx()
	defer cancel()
	if _, err := c.remote.BumpRev(ctx); err != nil {
		slog.Warn("model cache redis bump failed", "error", err)
	}
}

func (c *Cache) CurrentRev() int64 {
	if !c.enabled() || c.remote == nil {
		return 1
	}
	ctx, cancel := redisCtx()
	defer cancel()
	n, err := c.remote.GetRev(ctx)
	if err != nil || n == 0 {
		return 1
	}
	return n
}
