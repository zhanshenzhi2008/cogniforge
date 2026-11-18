package token

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cogniforge/internal/gateway/middleware"
)

// TestService_Issue 无 Redis 时签发
func TestService_Issue(t *testing.T) {
	svc := NewService(nil)
	token, err := svc.Issue(context.Background(), "user-123")
	require.NoError(t, err)
	require.NotEmpty(t, token)
}

// TestService_Validate 验证通过
func TestService_Validate(t *testing.T) {
	svc := NewService(nil)
	token, err := svc.Issue(context.Background(), "user-123")
	require.NoError(t, err)

	claims, err := svc.Validate(token)
	require.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "llm_call", claims.Scope)
}

// TestService_Validate_Invalid 无效 token
func TestService_Validate_Invalid(t *testing.T) {
	svc := NewService(nil)
	_, err := svc.Validate("invalid-token")
	assert.Error(t, err)
}

// TestService_Validate_WrongScope scope 错误
func TestService_Validate_WrongScope(t *testing.T) {
	svc := NewService(nil)
	badClaims := &LLMTokenClaims{
		UserID: "user-123",
		Scope:  "wrong_scope",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	badToken := jwt.NewWithClaims(jwt.SigningMethodHS256, badClaims)
	badTokenStr, err := badToken.SignedString(middleware.JWTSecret)
	require.NoError(t, err)

	_, err = svc.Validate(badTokenStr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid scope")
}

// TestService_Issue_UserIDEmpty 空 userID
func TestService_Issue_UserIDEmpty(t *testing.T) {
	svc := NewService(nil)
	_, err := svc.Issue(context.Background(), "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user_id required")
}

// TestService_JWTExpiration JWT 过期时间约 5 分钟
func TestService_JWTExpiration(t *testing.T) {
	svc := NewService(nil)
	token, err := svc.Issue(context.Background(), "user-exp")
	require.NoError(t, err)

	claims, err := svc.Validate(token)
	require.NoError(t, err)

	expTime := claims.ExpiresAt.Time
	expected := time.Now().Add(5 * time.Minute)
	delta := expTime.Sub(expected)
	assert.True(t, delta < time.Minute && delta > -time.Minute, "exp should be ~5min, delta=%v", delta)
}

// TestService_Get 等同 Issue
func TestService_Get(t *testing.T) {
	svc := NewService(nil)
	token, err := svc.Get(context.Background(), "user-get")
	require.NoError(t, err)
	require.NotEmpty(t, token)
}

// ===================== Redis 集成测试 =====================

func TestService_Redis_Idempotent(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("redis not available, skip")
	}
	rdb.Del(ctx, keyPrefix+"user-redis-test")

	svc := NewService(rdb)
	token1, err := svc.Issue(ctx, "user-redis-test")
	require.NoError(t, err)
	token2, err := svc.Issue(ctx, "user-redis-test")
	require.NoError(t, err)
	assert.Equal(t, token1, token2) // 幂等

	rdb.Del(ctx, keyPrefix+"user-redis-test")
}

func TestService_Redis_TTL(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("redis not available, skip")
	}
	rdb.Del(ctx, keyPrefix+"user-ttl-test")

	svc := NewService(rdb)
	_, err := svc.Issue(ctx, "user-ttl-test")
	require.NoError(t, err)

	ttl, err := rdb.TTL(ctx, keyPrefix+"user-ttl-test").Result()
	require.NoError(t, err)
	assert.True(t, ttl > 4*time.Minute && ttl <= 5*time.Minute, "TTL=%v", ttl)

	rdb.Del(ctx, keyPrefix+"user-ttl-test")
}
