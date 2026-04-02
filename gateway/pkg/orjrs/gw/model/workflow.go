package model

import (
	"time"

	"gorm.io/gorm"
)

// Workflow represents an AI workflow
type Workflow struct {
	ID          string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID      string         `gorm:"type:varchar(64);not null;index" json:"user_id"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Status      string         `gorm:"type:varchar(50);default:'draft'" json:"status"` // draft, published, archived
	Definition  string         `gorm:"type:text" json:"definition"`                    // JSON workflow definition
	Version     int            `gorm:"type:int;default:1" json:"version"`
	Metadata    JSONBMap       `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Workflow) TableName() string {
	return "workflows"
}

// WorkflowNode represents a node in a workflow
type WorkflowNode struct {
	ID         string    `gorm:"primaryKey;type:varchar(64)" json:"id"`
	WorkflowID string    `gorm:"type:varchar(64);not null;index" json:"workflow_id"`
	Type       string    `gorm:"type:varchar(50);not null" json:"type"` // start, end, agent, tool, condition, loop
	Name       string    `gorm:"type:varchar(255)" json:"name"`
	Config     JSONBMap  `gorm:"type:jsonb;default:'{}'" json:"config"`
	PositionX  float64   `gorm:"type:float;default:0" json:"position_x"`
	PositionY  float64   `gorm:"type:float;default:0" json:"position_y"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (WorkflowNode) TableName() string {
	return "workflow_nodes"
}

// WorkflowEdge represents a connection between nodes
type WorkflowEdge struct {
	ID         string    `gorm:"primaryKey;type:varchar(64)" json:"id"`
	WorkflowID string    `gorm:"type:varchar(64);not null;index" json:"workflow_id"`
	SourceID   string    `gorm:"type:varchar(64);not null" json:"source_id"`
	TargetID   string    `gorm:"type:varchar(64);not null" json:"target_id"`
	Label      string    `gorm:"type:varchar(255)" json:"label"`
	Condition  string    `gorm:"type:text" json:"condition"` // optional condition for conditional edges
	CreatedAt  time.Time `json:"created_at"`
}

func (WorkflowEdge) TableName() string {
	return "workflow_edges"
}

// WorkflowExecution represents a workflow execution instance
type WorkflowExecution struct {
	ID          string     `gorm:"primaryKey;type:varchar(64)" json:"id"`
	WorkflowID  string     `gorm:"type:varchar(64);not null;index" json:"workflow_id"`
	UserID      string     `gorm:"type:varchar(64);not null" json:"user_id"`
	Status      string     `gorm:"type:varchar(50)" json:"status"` // pending, running, completed, failed, cancelled
	Input       JSONBMap   `gorm:"type:jsonb;default:'{}'" json:"input"`
	Output      JSONBMap   `gorm:"type:jsonb;default:'{}'" json:"output"`
	Error       string     `gorm:"type:text" json:"error"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

func (WorkflowExecution) TableName() string {
	return "workflow_executions"
}
