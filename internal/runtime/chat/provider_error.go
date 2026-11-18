package chat

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"cogniforge/internal/shared/response"
)

// WriteNoActiveProvider 未配置活跃模型时写统一错误响应。
func WriteNoActiveProvider(c *gin.Context, err error) bool {
	if err == nil || !errors.Is(err, ErrNoActiveProvider) || c.Writer.Written() {
		return false
	}
	response.FailWithHTTPStatus(c, http.StatusServiceUnavailable, response.CodeNoActiveProvider, err.Error())
	return true
}
