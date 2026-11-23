package model

import (
	"time"

	"gorm.io/gorm"
)

// Team 协作模式（与 docs/01-requirements/11-multi-agent-framework-design.md 对齐）
const (
	TeamModePipeline   = "pipeline"
	TeamModeSupervisor = "supervisor"
	TeamModeHandoff    = "handoff"
)

// AgentTeam 多 Agent 小队说明书（逻辑实体；执行真相仍走工作流 checkpoint / TeamRunState）
type AgentTeam struct {
	ID                string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID            string         `gorm:"type:varchar(64);not null;index:idx_agent_teams_user_id" json:"user_id"`
	Name              string         `gorm:"type:varchar(255);not null" json:"name"`
	Description       string         `gorm:"type:text" json:"description"`
	Mode              string         `gorm:"type:varchar(32);not null;default:'supervisor'" json:"mode"` // pipeline|supervisor|handoff
	SupervisorAgentID string         `gorm:"type:varchar(64);index:idx_agent_teams_supervisor" json:"supervisor_agent_id"`
	MemberAgentIDs    JSONBArray     `gorm:"type:jsonb" json:"member_agent_ids"` // 工人 / 流水线成员（含或不含主管，由产品约定）
	WorkflowID        string         `gorm:"type:varchar(64);index:idx_agent_teams_workflow" json:"workflow_id"` // 可选绑定展开图
	Config            JSONBMap       `gorm:"type:jsonb" json:"config"`                                          // max_depth、max_steps 等
	Status            string         `gorm:"type:varchar(50);default:'active';index:idx_agent_teams_status" json:"status"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AgentTeam) TableName() string {
	return "agent_teams"
}

// TeamConfigDefaults 写入 Config 时的推荐默认（代码侧校验也用同一套）
func TeamConfigDefaults() JSONBMap {
	return JSONBMap{
		"max_depth": 8,
		"max_steps": 40,
	}
}
