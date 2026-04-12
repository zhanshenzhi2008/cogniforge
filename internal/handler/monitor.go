package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"cogniforge/internal/database"
	"cogniforge/internal/model"
)

// RequestLogQuery 请求日志查询参数
type RequestLogQuery struct {
	Page       int    `form:"page" json:"page"`
	PageSize   int    `form:"page_size" json:"page_size"`
	Method     string `form:"method" json:"method"`
	Path       string `form:"path" json:"path"`
	StatusCode int    `form:"status_code" json:"status_code"`
	StartTime  string `form:"start_time" json:"start_time"`
	EndTime    string `form:"end_time" json:"end_time"`
	TraceID    string `form:"trace_id" json:"trace_id"`
}

// ListRequestLogs 获取请求日志列表
func ListRequestLogs(c *gin.Context) {
	var query RequestLogQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		model.BadRequest(c, "参数解析失败")
		return
	}

	// 默认分页参数
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	// 获取当前用户 ID
	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)

	// 构建查询 - 支持过滤 user_id 为空的日志（用于调试）
	db := database.DB.Model(&model.RequestLog{}).
		Where("user_id = ? OR user_id = ''", userIDStr).
		Order("created_at DESC")

	// 应用过滤条件
	if query.Method != "" {
		db = db.Where("method = ?", strings.ToUpper(query.Method))
	}
	if query.Path != "" {
		db = db.Where("path LIKE ?", "%"+query.Path+"%")
	}
	if query.StatusCode > 0 {
		db = db.Where("status_code = ?", query.StatusCode)
	}
	if query.TraceID != "" {
		db = db.Where("trace_id = ?", query.TraceID)
	}
	if query.StartTime != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", query.StartTime); err == nil {
			db = db.Where("created_at >= ?", t)
		}
	}
	if query.EndTime != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", query.EndTime); err == nil {
			db = db.Where("created_at <= ?", t)
		}
	}

	// 获取总数
	var total int64
	db.Count(&total)

	// 分页查询
	offset := (query.Page - 1) * query.PageSize
	var logs []model.RequestLog
	if err := db.Offset(offset).Limit(query.PageSize).Find(&logs).Error; err != nil {
		model.InternalError(c, "查询失败")
		return
	}

	model.Success(c, gin.H{
		"logs":        logs,
		"total":       total,
		"page":        query.Page,
		"page_size":   query.PageSize,
		"total_pages": (total + int64(query.PageSize) - 1) / int64(query.PageSize),
	})
}

// GetRequestLog 获取请求日志详情
func GetRequestLog(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		model.BadRequest(c, "日志 ID 不能为空")
		return
	}

	// 获取当前用户 ID
	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)

	var log model.RequestLog
	if err := database.DB.Where("id = ? AND user_id = ?", id, userIDStr).First(&log).Error; err != nil {
		model.NotFound(c, "日志不存在")
		return
	}

	model.Success(c, log)
}

// GetUsageStats 获取用量统计
func GetUsageStats(c *gin.Context) {
	// 获取当前用户 ID
	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)

	// 默认查询最近 7 天
	days := 7
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -days)

	// 统计总请求数
	var totalRequests int64
	database.DB.Model(&model.RequestLog{}).
		Where("user_id = ? OR user_id = ''", userIDStr).
		Count(&totalRequests)

	// 统计总耗时
	var totalDuration int64
	database.DB.Model(&model.RequestLog{}).
		Select("COALESCE(SUM(duration), 0)").
		Where("user_id = ? OR user_id = ''", userIDStr).
		Scan(&totalDuration)

	// 统计错误请求数
	var errorRequests int64
	database.DB.Model(&model.RequestLog{}).
		Where("(user_id = ? OR user_id = '') AND status_code >= 400", userIDStr).
		Count(&errorRequests)

	// 按天统计请求量
	type DailyStat struct {
		Date         string `json:"date"`
		RequestCount int64  `json:"request_count"`
		AvgDuration  int64  `json:"avg_duration"`
		ErrorCount   int64  `json:"error_count"`
	}

	var dailyStats []DailyStat
	database.DB.Model(&model.RequestLog{}).
		Select("DATE(created_at) as date, COUNT(*) as request_count, COALESCE(AVG(duration), 0) as avg_duration, SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) as error_count").
		Where("(user_id = ? OR user_id = '') AND created_at >= ? AND created_at <= ?", userIDStr, startTime, endTime).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&dailyStats)

	// 按状态码统计
	type StatusStat struct {
		StatusCode int     `json:"status_code"`
		Count      int64   `json:"count"`
		Percentage float64 `json:"percentage"`
	}

	var statusStats []StatusStat
	database.DB.Model(&model.RequestLog{}).
		Select("status_code, COUNT(*) as count").
		Where("user_id = ? OR user_id = ''", userIDStr).
		Group("status_code").
		Order("count DESC").
		Scan(&statusStats)

	// 计算百分比和派生值
	avgDuration := calculateAvgDuration(totalRequests, totalDuration)
	errorRate := calculateErrorRate(totalRequests, errorRequests)

	// 计算百分比
	for i := range statusStats {
		if totalRequests > 0 {
			statusStats[i].Percentage = float64(statusStats[i].Count) / float64(totalRequests) * 100
		}
	}

	// 按方法统计
	type MethodStat struct {
		Method      string `json:"method"`
		Count       int64  `json:"count"`
		AvgDuration int64  `json:"avg_duration"`
	}

	var methodStats []MethodStat
	database.DB.Model(&model.RequestLog{}).
		Select("method, COUNT(*) as count, COALESCE(AVG(duration), 0) as avg_duration").
		Where("user_id = ? OR user_id = ''", userIDStr).
		Group("method").
		Order("count DESC").
		Scan(&methodStats)

	// 按路径统计（TOP 10）
	type PathStat struct {
		Path        string `json:"path"`
		Count       int64  `json:"count"`
		AvgDuration int64  `json:"avg_duration"`
	}

	var pathStats []PathStat
	database.DB.Model(&model.RequestLog{}).
		Select("path, COUNT(*) as count, COALESCE(AVG(duration), 0) as avg_duration").
		Where("user_id = ? OR user_id = ''", userIDStr).
		Group("path").
		Order("count DESC").
		Limit(10).
		Scan(&pathStats)

	model.Success(c, gin.H{
		"period":         days,
		"total_requests": totalRequests,
		"total_duration": totalDuration,
		"avg_duration":   avgDuration,
		"error_requests": errorRequests,
		"error_rate":     errorRate,
		"daily_stats":    dailyStats,
		"status_stats":   statusStats,
		"method_stats":   methodStats,
		"path_stats":     pathStats,
	})
}

// CalculateAvgDuration 计算平均耗时
func calculateAvgDuration(totalReq, totalDur int64) int64 {
	if totalReq > 0 {
		return totalDur / totalReq
	}
	return 0
}

// CalculateErrorRate 计算错误率
func calculateErrorRate(totalReq, errorReq int64) float64 {
	if totalReq > 0 {
		return float64(errorReq) / float64(totalReq) * 100
	}
	return 0
}

// GetRealtimeStats 获取实时统计（最近 5 分钟）
func GetRealtimeStats(c *gin.Context) {
	// 获取当前用户 ID
	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)

	// 最近 5 分钟
	endTime := time.Now()
	startTime := endTime.Add(-5 * time.Minute)

	var reqCount int64
	var avgDuration int64
	var errorCount int64

	database.DB.Model(&model.RequestLog{}).
		Select("COUNT(*) as count").
		Where("user_id = ? AND created_at >= ?", userIDStr, startTime).
		Scan(&reqCount)

	database.DB.Model(&model.RequestLog{}).
		Select("COALESCE(AVG(duration), 0) as avg_duration").
		Where("user_id = ? AND created_at >= ?", userIDStr, startTime).
		Scan(&avgDuration)

	database.DB.Model(&model.RequestLog{}).
		Select("COUNT(*) as count").
		Where("user_id = ? AND created_at >= ? AND status_code >= 400", userIDStr, startTime).
		Scan(&errorCount)

	// 每分钟请求数
	type MinuteStat struct {
		Minute       string `json:"minute"`
		RequestCount int64  `json:"request_count"`
	}

	var minuteStats []MinuteStat
	database.DB.Model(&model.RequestLog{}).
		Select("DATE_FORMAT(created_at, '%H:%i') as minute, COUNT(*) as request_count").
		Where("user_id = ? AND created_at >= ?", userIDStr, startTime).
		Group("DATE_FORMAT(created_at, '%H:%i')").
		Order("minute ASC").
		Scan(&minuteStats)

	model.Success(c, gin.H{
		"req_count":    reqCount,
		"avg_duration": avgDuration,
		"error_count":  errorCount,
		"minute_stats": minuteStats,
		"timestamp":    time.Now().Unix(),
	})
}

// ParseInt 解析整数
func ParseInt(s string, defaultVal int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return defaultVal
}
