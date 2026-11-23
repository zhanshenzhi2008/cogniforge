package team

import "cogniforge/internal/model"

// CreateTeamRequest 创建小队
type CreateTeamRequest struct {
	Name              string         `json:"name" binding:"required"`
	Description       string         `json:"description"`
	Mode              string         `json:"mode"` // pipeline|supervisor|handoff；默认 supervisor
	SupervisorAgentID string         `json:"supervisor_agent_id"`
	MemberAgentIDs    []string       `json:"member_agent_ids"`
	WorkflowID        string         `json:"workflow_id"`
	Config            model.JSONBMap `json:"config"`
	Status            string         `json:"status"`
}

// UpdateTeamRequest 更新小队（空字段表示不改，除指针语义外用零值跳过）
type UpdateTeamRequest struct {
	Name              *string         `json:"name"`
	Description       *string         `json:"description"`
	Mode              *string         `json:"mode"`
	SupervisorAgentID *string         `json:"supervisor_agent_id"`
	MemberAgentIDs    *[]string       `json:"member_agent_ids"`
	WorkflowID        *string         `json:"workflow_id"`
	Config            *model.JSONBMap `json:"config"`
	Status            *string         `json:"status"`
}
