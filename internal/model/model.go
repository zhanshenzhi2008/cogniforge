package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	Email     string         `gorm:"uniqueIndex;type:varchar(255);not null" json:"email"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	Password  string         `gorm:"type:varchar(255);not null" json:"-"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}

type ApiKey struct {
	ID        string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID    string         `gorm:"type:varchar(64);not null;index" json:"user_id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	Key       string         `gorm:"type:varchar(255);not null" json:"key"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ApiKey) TableName() string {
	return "api_keys"
}

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

type WorkflowNode struct {
	ID          string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	WorkflowID  string         `gorm:"type:varchar(64);not null;index" json:"workflow_id"`
	Type        string         `gorm:"type:varchar(50)" json:"type"`
	Name        string         `gorm:"type:varchar(255)" json:"name"`
	Config      string         `gorm:"type:text" json:"config"`
	PositionX   float64        `json:"position_x"`
	PositionY   float64        `json:"position_y"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (WorkflowNode) TableName() string {
	return "workflow_nodes"
}

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

type WorkflowExecution struct {
	ID          string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	WorkflowID  string         `gorm:"type:varchar(64);not null;index" json:"workflow_id"`
	UserID      string         `gorm:"type:varchar(64);not null;index" json:"user_id"`
	Status      string         `gorm:"type:varchar(50);default:'pending'" json:"status"`
	Input       JSONBMap       `gorm:"type:jsonb" json:"input"`
	Output      string         `gorm:"type:text" json:"output"`
	Error       string         `gorm:"type:text" json:"error"`
	StartedAt   *time.Time     `json:"started_at"`
	CompletedAt *time.Time     `json:"completed_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (WorkflowExecution) TableName() string {
	return "workflow_executions"
}

type JSONBArray []string

type JSONBMap map[string]any
