package model

import (
	"time"
)

// =============================================================================
// Monitor Models - 请求日志
// =============================================================================

type RequestLog struct {
	ID           string    `gorm:"primaryKey;type:varchar(64)" json:"id"`
	TraceID      string    `gorm:"type:varchar(64);index" json:"trace_id"`
	UserID       string    `gorm:"type:varchar(64);index" json:"user_id"`
	Method       string    `gorm:"type:varchar(10);not null" json:"method"`    // GET, POST, PUT, DELETE
	Path         string    `gorm:"type:varchar(500);not null" json:"path"`     // 请求路径
	Query        string    `gorm:"type:text" json:"query"`                     // 查询参数
	RequestBody  string    `gorm:"type:text" json:"request_body"`              // 请求体
	StatusCode   int       `gorm:"type:smallint;default:0" json:"status_code"` // 响应状态码
	ResponseBody string    `gorm:"type:text" json:"response_body"`             // 响应体
	Duration     int64     `gorm:"type:bigint;default:0" json:"duration"`      // 耗时(毫秒)
	UserAgent    string    `gorm:"type:varchar(500)" json:"user_agent"`        // 用户代理
	IP           string    `gorm:"type:varchar(50)" json:"ip"`                 // IP 地址
	Error        string    `gorm:"type:text" json:"error"`                     // 错误信息
	CreatedAt    time.Time `json:"created_at"`
}

func (RequestLog) TableName() string {
	return "request_logs"
}
