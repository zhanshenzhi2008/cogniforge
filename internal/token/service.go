package token

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"

	"cogniforge/internal/middleware"
)

const (
	// Redis 键前缀
	keyPrefix = "cogniforge:llm_token:"
	// TTL 5 分钟
	tokenTTL = 5 * time.Minute
)

// ErrTokenNotFound token 不存在或已过期
var ErrTokenNotFound = errors.New("token not found or expired")

// LLMTokenClaims JWT 临时凭证声明，供 Python 验证用。
type LLMTokenClaims struct {
	UserID string `json:"user_id"`
	Scope  string `json:"scope"` // "llm_call"
	jwt.RegisteredClaims
}

// Service 签发/查询 LLM 临时凭证。
// 凭证本身是 JWT（防篡改），Redis 只做 TTL 缓存（幂等）。
type Service struct {
	rdb *redis.Client
}

// NewService 创建 token service。rdb 可为 nil（走仅 JWT 模式，TTL 由 JWT 自己控制）。
func NewService(rdb *redis.Client) *Service {
	return &Service{rdb: rdb}
}

// Issue 签发 LLM 临时凭证。
// - Redis 有未过期缓存 → 直接返回已有 token（幂等）
// - Redis 无 → 签发新 JWT，写入 Redis（TTL 5min），返回 token
func (s *Service) Issue(ctx context.Context, userID string) (string, error) {
	if userID == "" {
		return "", errors.New("user_id required")
	}

	key := keyPrefix + userID

	// 幂等：Redis 有未过期 token，直接返回
	if s.rdb != nil {
		existing, err := s.rdb.Get(ctx, key).Result()
		if err == nil && existing != "" {
			return existing, nil
		}
	}

	// 签发新 JWT
	token, err := s.sign(userID)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	// 写入 Redis（TTL 5min）
	if s.rdb != nil {
		_ = s.rdb.Set(ctx, key, token, tokenTTL).Err()
	}

	return token, nil
}

// Get 取当前有效 token（Redis 有则返回，无则签发）。
// 用于屏障降级路径：Python 请求时 Go 需要拿到 token。
func (s *Service) Get(ctx context.Context, userID string) (string, error) {
	return s.Issue(ctx, userID)
}

// Validate 验证 LLM token 是否有效（仅 JWT 签名校验，Redis TTL 由签发端控制）。
// Python 侧用此校验。
func (s *Service) Validate(tokenString string) (*LLMTokenClaims, error) {
	claims := &LLMTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return middleware.JWTSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.Scope != "llm_call" {
		return nil, errors.New("invalid scope")
	}
	return claims, nil
}

// sign 生成 JWT（TTL 由 JWT exp 字段控制，写入 Redis 不影响 JWT 有效性）。
func (s *Service) sign(userID string) (string, error) {
	claims := &LLMTokenClaims{
		UserID: userID,
		Scope:  "llm_call",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(middleware.JWTSecret)
}

// IssueHandler HTTP handler：GET /api/v1/token/llm
// 从上下文取已认证 user_id，签发 LLM 临时凭证。
func (s *Service) IssueHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, gin.H{"error": "unauthorized"})
			return
		}

		token, err := s.Issue(c.Request.Context(), userID.(string))
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"token":     token,
			"expires_in": int(tokenTTL.Seconds()),
			"scope":     "llm_call",
		})
	}
}

// ValidateHandler HTTP handler：POST /api/v1/token/validate
// Python 回调 Go 验证 token（Go 侧验证 JWT 签名）。
func (s *Service) ValidateHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Token string `json:"token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "token required"})
			return
		}

		claims, err := s.Validate(req.Token)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid token"})
			return
		}

		c.JSON(200, gin.H{
			"valid":   true,
			"user_id": claims.UserID,
			"scope":   claims.Scope,
		})
	}
}
