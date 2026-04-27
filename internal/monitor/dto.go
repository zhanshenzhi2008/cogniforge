package monitor

// ============ 请求结构 ============

type RequestLogQuery struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	Method     string `form:"method"`
	Path       string `form:"path"`
	StatusCode int    `form:"status_code"`
	StartTime  string `form:"start_time"`
	EndTime    string `form:"end_time"`
	TraceID    string `form:"trace_id"`
}

// ============ 响应结构 ============

type DailyStat struct {
	Date         string `json:"date"`
	RequestCount int64  `json:"request_count"`
	AvgDuration  int64  `json:"avg_duration"`
	ErrorCount   int64  `json:"error_count"`
}

type StatusStat struct {
	StatusCode int     `json:"status_code"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

type MethodStat struct {
	Method      string `json:"method"`
	Count       int64  `json:"count"`
	AvgDuration int64  `json:"avg_duration"`
}

type PathStat struct {
	Path        string `json:"path"`
	Count       int64  `json:"count"`
	AvgDuration int64  `json:"avg_duration"`
}

type MinuteStat struct {
	Minute       string `json:"minute"`
	RequestCount int64  `json:"request_count"`
}

type UsageStatsResponse struct {
	Period        int          `json:"period"`
	TotalRequests int64        `json:"total_requests"`
	TotalDuration int64        `json:"total_duration"`
	AvgDuration   int64        `json:"avg_duration"`
	ErrorRequests int64        `json:"error_requests"`
	ErrorRate     float64      `json:"error_rate"`
	DailyStats    []DailyStat  `json:"daily_stats"`
	StatusStats   []StatusStat `json:"status_stats"`
	MethodStats   []MethodStat `json:"method_stats"`
	PathStats     []PathStat   `json:"path_stats"`
}

type RealtimeStatsResponse struct {
	ReqCount    int64        `json:"req_count"`
	AvgDuration int64        `json:"avg_duration"`
	ErrorCount  int64        `json:"error_count"`
	MinuteStats []MinuteStat `json:"minute_stats"`
	Timestamp   int64        `json:"timestamp"`
}

type PaginatedLogsResponse struct {
	Logs       any   `json:"logs"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int64 `json:"total_pages"`
}
