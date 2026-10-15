package model

import (
	"time"

	"gorm.io/gorm"
)

// =============================================================================
// Eval Models - 设计评估体系（阶段十六）
// =============================================================================

const (
	EvalJudgeContains = "contains"
	EvalJudgeExact    = "exact"
	EvalJudgeLLM      = "llm_judge"

	EvalTargetModel  = "model"
	EvalTargetAgent  = "agent"
	EvalTargetPrompt = "prompt" // 提示词版本 id；需配合 CreateRunRequest.Model

	EvalRunPending   = "pending"
	EvalRunRunning   = "running"
	EvalRunCompleted = "completed"
	EvalRunFailed    = "failed"

	EvalResultPass  = "pass"
	EvalResultFail  = "fail"
	EvalResultSkip  = "skip"
	EvalResultError = "error"
)

// EvalSuite 评测集
type EvalSuite struct {
	ID          string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID      string         `gorm:"type:varchar(64);not null;index:idx_eval_suites_user_id" json:"user_id"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (EvalSuite) TableName() string { return "eval_suites" }

// EvalCase 评测用例
type EvalCase struct {
	ID          string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	SuiteID     string         `gorm:"type:varchar(64);not null;index:idx_eval_cases_suite_id" json:"suite_id"`
	UserID      string         `gorm:"type:varchar(64);not null;index:idx_eval_cases_user_id" json:"user_id"`
	Name        string         `gorm:"type:varchar(255)" json:"name"`
	Input       string         `gorm:"type:text;not null" json:"input"`
	Expected    string         `gorm:"type:text" json:"expected"`
	JudgeType   string         `gorm:"type:varchar(32);not null;default:contains" json:"judge_type"`
	JudgeConfig JSONBMap       `gorm:"type:jsonb" json:"judge_config"`
	Weight      float64        `gorm:"default:1" json:"weight"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (EvalCase) TableName() string { return "eval_cases" }

// EvalRun 一次评测运行
type EvalRun struct {
	ID           string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	SuiteID      string         `gorm:"type:varchar(64);not null;index:idx_eval_runs_suite_id" json:"suite_id"`
	UserID       string         `gorm:"type:varchar(64);not null;index:idx_eval_runs_user_id" json:"user_id"`
	TargetType   string         `gorm:"type:varchar(32);not null" json:"target_type"`
	TargetID     string         `gorm:"type:varchar(255);not null" json:"target_id"`
	Model        string         `gorm:"type:varchar(100)" json:"model,omitempty"`          // target=prompt 时使用的聊天模型
	SystemPrompt string         `gorm:"type:text" json:"system_prompt,omitempty"`          // 实际使用的 system（便于回看）
	Status       string         `gorm:"type:varchar(32);not null;default:pending" json:"status"`
	PassCount    int            `json:"pass_count"`
	FailCount    int            `json:"fail_count"`
	SkipCount    int            `json:"skip_count"`
	Score        float64        `json:"score"`
	ErrorMessage string         `gorm:"type:text" json:"error_message,omitempty"`
	StartedAt    *time.Time     `json:"started_at,omitempty"`
	FinishedAt   *time.Time     `json:"finished_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (EvalRun) TableName() string { return "eval_runs" }

// EvalResult 单题结果
type EvalResult struct {
	ID           string    `gorm:"primaryKey;type:varchar(64)" json:"id"`
	RunID        string    `gorm:"type:varchar(64);not null;index:idx_eval_results_run_id" json:"run_id"`
	CaseID       string    `gorm:"type:varchar(64);not null;index:idx_eval_results_case_id" json:"case_id"`
	Status       string    `gorm:"type:varchar(32);not null" json:"status"`
	Score        float64   `json:"score"`
	ActualOutput string    `gorm:"type:text" json:"actual_output"`
	LatencyMS    int64     `json:"latency_ms"`
	Detail       JSONBMap  `gorm:"type:jsonb" json:"detail"`
	CreatedAt    time.Time `json:"created_at"`
}

func (EvalResult) TableName() string { return "eval_results" }
