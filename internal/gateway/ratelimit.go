package gateway

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"

	"cogniforge/internal/governance/quota"
	"cogniforge/internal/shared/response"
)

// RateLimiter 边缘限流最小接口。
// 实现一般由 governance/quota.Service 经 WithSource 适配；策略与计数仍在治理层。
type RateLimiter interface {
	Allow(ctx context.Context, userID string) error
}

// quotaAllow 适配 quota.Service.Allow(ctx, userID, source)。
type quotaAllow struct {
	svc    *quota.Service
	source string
}

func (q quotaAllow) Allow(ctx context.Context, userID string) error {
	if q.svc == nil {
		return nil
	}
	return q.svc.Allow(ctx, userID, q.source)
}

// RateLimitFromQuota 用治理层配额服务做网关限流（固定 source，如 "http"）。
//
// 注意：chat/agent 路径已在业务内调用 Allow，勿再对本路由挂本中间件，否则会双重扣次。
// 适用于尚未内嵌配额检查的 HTTP 路由（防刷、公开接口等）。
func RateLimitFromQuota(svc *quota.Service, source string) gin.HandlerFunc {
	if source == "" {
		source = "http"
	}
	return RateLimit(quotaAllow{svc: svc, source: source})
}

// RateLimit 在网关层做请求限流。
// userID 取自 JWT 中间件写入的 "user_id"；未登录则跳过。
func RateLimit(rl RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rl == nil {
			c.Next()
			return
		}
		uid, _ := c.Get("user_id")
		userID, _ := uid.(string)
		if userID == "" {
			c.Next()
			return
		}
		if err := rl.Allow(c.Request.Context(), userID); err != nil {
			switch {
			case errors.Is(err, quota.ErrRateLimited):
				response.FailWithHTTPStatus(c, 429, response.CodeRateLimitExceeded, "")
			case errors.Is(err, quota.ErrExceeded):
				response.FailWithHTTPStatus(c, 429, response.CodeUserQuotaExceeded, "")
			default:
				response.InternalError(c, "rate limit check failed")
			}
			c.Abort()
			return
		}
		c.Next()
	}
}
