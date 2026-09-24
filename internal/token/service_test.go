package token

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cogniforge/internal/middleware"
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

// ===================== HTTP Handler 测试 =====================

func setupRouter(svc *Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	auth := r.Group("/api/v1")
	auth.Use(func(c *gin.Context) {
		c.Set("user_id", "user-test")
		c.Next()
	})
	NewHandler(svc).RegisterRoutes(auth)
	return r
}

func TestHandler_Issue_LLM(t *testing.T) {
	svc := NewService(nil)
	r := setupRouter(svc)

	req := httptest.NewRequest("GET", "/api/v1/token/llm", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"token"`)
	assert.Contains(t, w.Body.String(), `"expires_in":300`)
	assert.Contains(t, w.Body.String(), `"scope":"llm_call"`)
}

func TestHandler_Issue_NoAuth(t *testing.T) {
	svc := NewService(nil)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	NewHandler(svc).RegisterRoutes(r.Group("/api/v1"))

	req := httptest.NewRequest("GET", "/api/v1/token/llm", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_Validate_OK(t *testing.T) {
	svc := NewService(nil)
	r := setupRouter(svc)

	// 先拿 token
	issueReq := httptest.NewRequest("GET", "/api/v1/token/llm", nil)
	issueW := httptest.NewRecorder()
	r.ServeHTTP(issueW, issueReq)
	token := extractToken(t, issueW.Body.String())

	// 验证
	body := `{"token":"` + token + `"}`
	validateReq := httptest.NewRequest("POST", "/api/v1/token/validate", strings.NewReader(body))
	validateReq.Header.Set("Content-Type", "application/json")
	validateW := httptest.NewRecorder()
	r.ServeHTTP(validateW, validateReq)

	assert.Equal(t, http.StatusOK, validateW.Code)
	assert.Contains(t, validateW.Body.String(), `"valid":true`)
	assert.Contains(t, validateW.Body.String(), `"user_id":"user-test"`)
}

func TestHandler_Validate_Invalid(t *testing.T) {
	svc := NewService(nil)
	r := setupRouter(svc)

	body := `{"token":"invalid-token"}`
	req := httptest.NewRequest("POST", "/api/v1/token/validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"invalid token"`)
}

func TestHandler_Validate_MissingToken(t *testing.T) {
	svc := NewService(nil)
	r := setupRouter(svc)

	req := httptest.NewRequest("POST", "/api/v1/token/validate", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
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

// ===================== 辅助函数 =====================

func extractToken(t *testing.T, body string) string {
	t.Helper()
	start := strings.Index(body, `"token":"`) + len(`"token":"`)
	end := strings.Index(body[start:], `"`)
	return body[start : start+end]
}
