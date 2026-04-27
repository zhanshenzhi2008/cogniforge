package monitor

import (
	"github.com/gin-gonic/gin"

	"cogniforge/internal/response"
)

type MonitorHandler struct {
	service *MonitorService
}

func NewMonitorHandler() *MonitorHandler {
	return &MonitorHandler{
		service: NewMonitorService(),
	}
}

// ListRequestLogs 获取请求日志列表
func (h *MonitorHandler) ListRequestLogs(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)

	var query RequestLogQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "参数解析失败")
		return
	}

	result, err := h.service.ListRequestLogs(userIDStr, &query)
	if err != nil {
		response.InternalError(c, "查询失败")
		return
	}

	response.Success(c, result)
}

// GetRequestLog 获取请求日志详情
func (h *MonitorHandler) GetRequestLog(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)

	logID := c.Param("id")
	if logID == "" {
		response.BadRequest(c, "日志 ID 不能为空")
		return
	}

	log, err := h.service.GetRequestLog(userIDStr, logID)
	if err != nil {
		response.NotFound(c, "日志不存在")
		return
	}

	response.Success(c, log)
}

// GetUsageStats 获取用量统计
func (h *MonitorHandler) GetUsageStats(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)

	result, err := h.service.GetUsageStats(userIDStr, 7)
	if err != nil {
		response.InternalError(c, "查询失败")
		return
	}

	response.Success(c, result)
}

// GetRealtimeStats 获取实时统计
func (h *MonitorHandler) GetRealtimeStats(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)

	result, err := h.service.GetRealtimeStats(userIDStr)
	if err != nil {
		response.InternalError(c, "查询失败")
		return
	}

	response.Success(c, result)
}

// RegisterRoutes 注册路由
func (h *MonitorHandler) RegisterRoutes(rg *gin.RouterGroup) {
	monitor := rg.Group("/monitor")
	{
		monitor.GET("/logs", h.ListRequestLogs)
		monitor.GET("/logs/:id", h.GetRequestLog)
		monitor.GET("/stats", h.GetUsageStats)
		monitor.GET("/stats/realtime", h.GetRealtimeStats)
	}
}
