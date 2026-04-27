package monitor

import (
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"cogniforge/internal/database"
	"cogniforge/internal/model"
)

type MonitorService struct {
	db *gorm.DB
}

func NewMonitorService() *MonitorService {
	return &MonitorService{db: database.DB}
}

// ListRequestLogs 获取请求日志列表
func (s *MonitorService) ListRequestLogs(userID string, query *RequestLogQuery) (*PaginatedLogsResponse, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	db := s.db.Model(&model.RequestLog{}).
		Where("user_id = ? OR user_id = ''", userID).
		Order("created_at DESC")

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

	var total int64
	db.Count(&total)

	offset := (query.Page - 1) * query.PageSize
	var logs []model.RequestLog
	if err := db.Offset(offset).Limit(query.PageSize).Find(&logs).Error; err != nil {
		return nil, err
	}

	return &PaginatedLogsResponse{
		Logs:       logs,
		Total:      total,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: (total + int64(query.PageSize) - 1) / int64(query.PageSize),
	}, nil
}

// GetRequestLog 获取请求日志详情
func (s *MonitorService) GetRequestLog(userID, logID string) (*model.RequestLog, error) {
	var log model.RequestLog
	if err := s.db.Where("id = ? AND user_id = ?", logID, userID).First(&log).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

// GetUsageStats 获取用量统计
func (s *MonitorService) GetUsageStats(userID string, days int) (*UsageStatsResponse, error) {
	if days <= 0 {
		days = 7
	}

	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -days)

	// 统计总请求数
	var totalRequests int64
	s.db.Model(&model.RequestLog{}).
		Where("user_id = ? OR user_id = ''", userID).
		Count(&totalRequests)

	// 统计总耗时
	var totalDuration int64
	s.db.Model(&model.RequestLog{}).
		Select("COALESCE(SUM(duration), 0)").
		Where("user_id = ? OR user_id = ''", userID).
		Scan(&totalDuration)

	// 统计错误请求数
	var errorRequests int64
	s.db.Model(&model.RequestLog{}).
		Where("(user_id = ? OR user_id = '') AND status_code >= 400", userID).
		Count(&errorRequests)

	// 按天统计
	var dailyStats []DailyStat
	s.db.Model(&model.RequestLog{}).
		Select("DATE(created_at) as date, COUNT(*) as request_count, COALESCE(AVG(duration), 0) as avg_duration, SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) as error_count").
		Where("(user_id = ? OR user_id = '') AND created_at >= ? AND created_at <= ?", userID, startTime, endTime).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&dailyStats)

	// 按状态码统计
	var statusStats []StatusStat
	s.db.Model(&model.RequestLog{}).
		Select("status_code, COUNT(*) as count").
		Where("user_id = ? OR user_id = ''", userID).
		Group("status_code").
		Order("count DESC").
		Scan(&statusStats)

	avgDuration := calculateAvgDuration(totalRequests, totalDuration)
	errorRate := calculateErrorRate(totalRequests, errorRequests)

	for i := range statusStats {
		if totalRequests > 0 {
			statusStats[i].Percentage = float64(statusStats[i].Count) / float64(totalRequests) * 100
		}
	}

	// 按方法统计
	var methodStats []MethodStat
	s.db.Model(&model.RequestLog{}).
		Select("method, COUNT(*) as count, COALESCE(AVG(duration), 0) as avg_duration").
		Where("user_id = ? OR user_id = ''", userID).
		Group("method").
		Order("count DESC").
		Scan(&methodStats)

	// 按路径统计（TOP 10）
	var pathStats []PathStat
	s.db.Model(&model.RequestLog{}).
		Select("path, COUNT(*) as count, COALESCE(AVG(duration), 0) as avg_duration").
		Where("user_id = ? OR user_id = ''", userID).
		Group("path").
		Order("count DESC").
		Limit(10).
		Scan(&pathStats)

	return &UsageStatsResponse{
		Period:        days,
		TotalRequests: totalRequests,
		TotalDuration: totalDuration,
		AvgDuration:   avgDuration,
		ErrorRequests: errorRequests,
		ErrorRate:     errorRate,
		DailyStats:    dailyStats,
		StatusStats:   statusStats,
		MethodStats:   methodStats,
		PathStats:     pathStats,
	}, nil
}

// GetRealtimeStats 获取实时统计（最近 5 分钟）
func (s *MonitorService) GetRealtimeStats(userID string) (*RealtimeStatsResponse, error) {
	endTime := time.Now()
	startTime := endTime.Add(-5 * time.Minute)

	var reqCount int64
	s.db.Model(&model.RequestLog{}).
		Select("COUNT(*) as count").
		Where("user_id = ? AND created_at >= ?", userID, startTime).
		Scan(&reqCount)

	var avgDuration int64
	s.db.Model(&model.RequestLog{}).
		Select("COALESCE(AVG(duration), 0) as avg_duration").
		Where("user_id = ? AND created_at >= ?", userID, startTime).
		Scan(&avgDuration)

	var errorCount int64
	s.db.Model(&model.RequestLog{}).
		Select("COUNT(*) as count").
		Where("user_id = ? AND created_at >= ? AND status_code >= 400", userID, startTime).
		Scan(&errorCount)

	var minuteStats []MinuteStat
	s.db.Model(&model.RequestLog{}).
		Select("DATE_FORMAT(created_at, '%H:%i') as minute, COUNT(*) as request_count").
		Where("user_id = ? AND created_at >= ?", userID, startTime).
		Group("DATE_FORMAT(created_at, '%H:%i')").
		Order("minute ASC").
		Scan(&minuteStats)

	return &RealtimeStatsResponse{
		ReqCount:    reqCount,
		AvgDuration: avgDuration,
		ErrorCount:  errorCount,
		MinuteStats: minuteStats,
		Timestamp:   time.Now().Unix(),
	}, nil
}

// ============ 辅助函数 ============

func calculateAvgDuration(totalReq, totalDur int64) int64 {
	if totalReq > 0 {
		return totalDur / totalReq
	}
	return 0
}

func calculateErrorRate(totalReq, errorReq int64) float64 {
	if totalReq > 0 {
		return float64(errorReq) / float64(totalReq) * 100
	}
	return 0
}

func parseInt(s string, defaultVal int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return defaultVal
}
