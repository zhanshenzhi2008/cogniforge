package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

var AppVersion = "1.0.0"

// Health 健康检查接口
// @Summary 健康检查
// @Description 检查服务是否正常运行
// @Tags system
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func Health(c *gin.Context) {
	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   AppVersion,
	}
	c.JSON(http.StatusOK, response)
}

// Ready 就绪检查接口 (用于 K8s readiness probe)
// @Summary 就绪检查
// @Description 检查服务是否准备好接收流量
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /ready [get]
func Ready(c *gin.Context) {
	// TODO: 添加依赖检查（数据库、Redis 等）
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

// Live 存活检查接口 (用于 K8s liveness probe)
// @Summary 存活检查
// @Description 检查服务是否存活
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Router /live [get]
func Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "alive"})
}
