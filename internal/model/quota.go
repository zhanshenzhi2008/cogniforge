package model

import "time"

// QuotaPolicy 配额策略。user_id 为空表示全站默认，有值则覆盖该用户。
type QuotaPolicy struct {
	ID             string    `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID         *string   `gorm:"type:varchar(64);index" json:"user_id"`
	DailyRequests  int       `gorm:"not null;default:30" json:"daily_requests"`
	DailyTokens    int64     `gorm:"not null;default:100000" json:"daily_tokens"`
	MonthlyTokens  int64     `gorm:"not null;default:1000000" json:"monthly_tokens"`
	RPM            int       `gorm:"not null;default:8" json:"rpm"`
	AdminUnlimited bool      `gorm:"not null;default:true" json:"admin_unlimited"`
	Enabled        bool      `gorm:"not null;default:true" json:"enabled"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (QuotaPolicy) TableName() string {
	return "quota_policies"
}

// LLMUsageEvent 每次模型调用的用量明细（图表原料，与 request_logs 分开）。
type LLMUsageEvent struct {
	ID               string    `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID           string    `gorm:"type:varchar(64);index:idx_llm_usage_user_created,priority:1" json:"user_id"`
	Source           string    `gorm:"type:varchar(32);not null" json:"source"`
	Model            string    `gorm:"type:varchar(100)" json:"model"`
	PromptTokens     int       `gorm:"not null;default:0" json:"prompt_tokens"`
	CompletionTokens int       `gorm:"not null;default:0" json:"completion_tokens"`
	TotalTokens      int       `gorm:"not null;default:0" json:"total_tokens"`
	TokensEstimated  bool      `gorm:"not null;default:false" json:"tokens_estimated"`
	Status           string    `gorm:"type:varchar(32);not null" json:"status"`
	LatencyMS        int64     `json:"latency_ms"`
	TraceID          string    `gorm:"type:varchar(64)" json:"trace_id"`
	CreatedAt        time.Time `gorm:"index:idx_llm_usage_user_created,priority:2;index:idx_llm_usage_created" json:"created_at"`
}

func (LLMUsageEvent) TableName() string {
	return "llm_usage_events"
}
