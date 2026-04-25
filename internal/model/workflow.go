package model

import (
	"time"

	"gorm.io/gorm"
)

// =============================================================================
// Workflow Models - 工作流相关
// =============================================================================

// Workflow 工作流定义表
type Workflow struct {
	ID          string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID      string         `gorm:"type:varchar(64);not null;index" json:"user_id"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Definition  string         `gorm:"type:text" json:"definition"`
	Status      string         `gorm:"type:varchar(50);default:'draft'" json:"status"`
	Version     int            `gorm:"default:1" json:"version"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Workflow) TableName() string {
	return "workflows"
}

// WorkflowNode 工作流节点表
type WorkflowNode struct {
	ID         string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	WorkflowID string         `gorm:"type:varchar(64);not null;index" json:"workflow_id"`
	Type       string         `gorm:"type:varchar(50)" json:"type"`
	Name       string         `gorm:"type:varchar(255)" json:"name"`
	Config     string         `gorm:"type:text" json:"config"`
	PositionX  float64        `json:"position_x"`
	PositionY  float64        `json:"position_y"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (WorkflowNode) TableName() string {
	return "workflow_nodes"
}

// WorkflowEdge 工作流边表（节点连接关系）
type WorkflowEdge struct {
	ID         string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	WorkflowID string         `gorm:"type:varchar(64);not null;index" json:"workflow_id"`
	SourceID   string         `gorm:"type:varchar(64)" json:"source_id"`
	TargetID   string         `gorm:"type:varchar(64)" json:"target_id"`
	Config     string         `gorm:"type:text" json:"config"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (WorkflowEdge) TableName() string {
	return "workflow_edges"
}

// WorkflowExecution 工作流执行记录表
type WorkflowExecution struct {
	ID          string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	WorkflowID  string         `gorm:"type:varchar(64);not null;index" json:"workflow_id"`
	UserID      string         `gorm:"type:varchar(64);not null;index" json:"user_id"`
	Status      string         `gorm:"type:varchar(50);default:'pending'" json:"status"`
	Input       JSONBMap       `gorm:"type:jsonb" json:"input"`
	Output      string         `gorm:"type:text" json:"output"`
	Error       string         `gorm:"type:text" json:"error"`
	CurrentNode string         `gorm:"type:varchar(64)" json:"current_node"` // 当前正在执行的节点（用于调试）
	StartedAt   *time.Time     `json:"started_at"`
	CompletedAt *time.Time     `json:"completed_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (WorkflowExecution) TableName() string {
	return "workflow_executions"
}

// WorkflowSchedule 工作流定时调度表
type WorkflowSchedule struct {
	ID             string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	WorkflowID     string         `gorm:"type:varchar(64);not null;index" json:"workflow_id"`
	UserID         string         `gorm:"type:varchar(64);not null;index" json:"user_id"`
	Name           string         `gorm:"type:varchar(255)" json:"name"`
	CronExpression string         `gorm:"type:varchar(100)" json:"cron_expression"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	DefaultInput   JSONBMap       `gorm:"type:jsonb" json:"default_input"`
	LastRun        *time.Time     `json:"last_run"`
	LastError      string         `gorm:"type:text" json:"last_error"`
	NextRun        *time.Time     `json:"next_run"`
	RunCount       int            `gorm:"default:0" json:"run_count"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (WorkflowSchedule) TableName() string {
	return "workflow_schedules"
}

// =============================================================================
// Agent Models - AI Agent
// =============================================================================

type Agent struct {
	ID           string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID       string         `gorm:"type:varchar(64);not null;index" json:"user_id"`
	Name         string         `gorm:"type:varchar(255);not null" json:"name"`
	Description  string         `gorm:"type:text" json:"description"`
	Model        string         `gorm:"type:varchar(100)" json:"model"`
	SystemPrompt string         `gorm:"type:text" json:"system_prompt"`
	Tools        JSONBArray     `gorm:"type:jsonb" json:"tools"`
	MemoryType   string         `gorm:"type:varchar(50)" json:"memory_type"`
	MemoryTurns  int            `gorm:"default:10" json:"memory_turns"`
	InputFilter  bool           `gorm:"default:true" json:"input_filter"`
	OutputFilter bool           `gorm:"default:true" json:"output_filter"`
	Status       string         `gorm:"type:varchar(50);default:'active'" json:"status"`
	Metadata     JSONBMap       `gorm:"type:jsonb" json:"metadata"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Agent) TableName() string {
	return "agents"
}
