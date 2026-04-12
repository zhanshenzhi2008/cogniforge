package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"cogniforge/internal/database"
	"cogniforge/internal/model"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 记录请求路径和方法
		path := c.Request.URL.Path
		method := c.Request.Method

		// 跳过日志记录的健康检查端点
		if path == "/health" || path == "/ready" || path == "/live" {
			c.Next()
			return
		}

		// 生成 Trace ID
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = strings.Split(uuid.New().String(), "-")[0]
		}

		// 读取请求体
		var requestBody string
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			if len(bodyBytes) > 0 {
				// 敏感信息脱敏
				requestBody = maskSensitiveData(string(bodyBytes))
			}
		}

		// 创建响应写入器
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// 获取用户 ID
		userID, _ := c.Get("user_id")
		userIDStr := ""
		if userID != nil {
			userIDStr = userID.(string)
		}

		// 处理请求
		c.Next()

		// 计算耗时
		duration := time.Since(start).Milliseconds()

		// 记录日志
		logEntry := model.RequestLog{
			ID:           generateLogID(),
			TraceID:      traceID,
			UserID:       userIDStr,
			Method:       method,
			Path:         path,
			Query:        c.Request.URL.RawQuery,
			RequestBody:  requestBody,
			StatusCode:   blw.Status(),
			ResponseBody: truncateString(blw.body.String(), 2000),
			Duration:     duration,
			UserAgent:    c.Request.UserAgent(),
			IP:           c.ClientIP(),
			Error:        "",
			CreatedAt:    start,
		}

		// 如果有错误，记录错误信息
		if len(c.Errors) > 0 {
			logEntry.Error = c.Errors.String()
		}

		// 异步写入数据库
		go func(log model.RequestLog) {
			if err := database.DB.Create(&log).Error; err != nil {
				// 日志写入失败不应该影响主流程
			}
		}(logEntry)
	}
}

// bodyLogWriter 捕获响应体
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// generateLogID 生成日志 ID
func generateLogID() string {
	return "log_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:28]
}

// maskSensitiveData 脱敏敏感数据
func maskSensitiveData(data string) string {
	// 脱敏密码字段
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &jsonData); err != nil {
		return data
	}

	sensitiveFields := []string{"password", "token", "secret", "key", "api_key", "apiKey"}
	for key, value := range jsonData {
		for _, field := range sensitiveFields {
			if strings.Contains(strings.ToLower(key), field) {
				if str, ok := value.(string); ok && len(str) > 0 {
					jsonData[key] = maskString(str)
				}
			}
		}
	}

	result, _ := json.Marshal(jsonData)
	return string(result)
}

// maskString 掩码字符串
func maskString(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + strings.Repeat("*", min(len(s)-4, 8)) + s[len(s)-2:]
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "...[truncated]"
}
