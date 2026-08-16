package modelcache

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"cogniforge/internal/config"
)

func redisAddr(cfg *config.Config) string {
	host := cfg.Redis.Host
	if host == "" {
		host = "localhost"
	}
	port := cfg.Redis.Port
	if port == 0 {
		port = 6379
	}
	return fmt.Sprintf("%s:%d", host, port)
}

func redisPassword(cfg *config.Config) string {
	if cfg.Redis.Password != "" {
		return cfg.Redis.Password
	}
	return ""
}

// DialRedis 连 Redis；失败返回 nil，调用方降级为仅本地缓存。
func DialRedis(cfg *config.Config) *redis.Client {
	if cfg == nil {
		return nil
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:         redisAddr(cfg),
		Password:     redisPassword(cfg),
		DB:           cfg.Redis.DB,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  500 * time.Millisecond,
		WriteTimeout: 500 * time.Millisecond,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Warn("redis unavailable, model cache will be local-only", "addr", redisAddr(cfg), "error", err)
		_ = rdb.Close()
		return nil
	}
	slog.Info("redis connected for model cache", "addr", redisAddr(cfg))
	return rdb
}

func NewFromRedis(rdb *redis.Client) *Cache {
	if rdb == nil {
		return New(nil)
	}
	return New(newRedisRemote(rdb))
}
