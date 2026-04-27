package agent

import "cogniforge/internal/model"

// ============ 请求结构 ============

type CreateAgentRequest struct {
	Name         string   `json:"name" binding:"required"`
	Description  string   `json:"description"`
	Model        string   `json:"model" binding:"required"`
	SystemPrompt string   `json:"system_prompt"`
	Tools        []string `json:"tools"`
}

type UpdateAgentRequest struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Model        string   `json:"model"`
	SystemPrompt string   `json:"system_prompt"`
	Tools        []string `json:"tools"`
	Status       string   `json:"status"`
}

// ============ 响应结构 ============

type AgentResponse struct {
	Agent *model.Agent `json:"agent"`
}

// ============ 辅助函数 ============

func ToAgentResponse(agent *model.Agent) *AgentResponse {
	return &AgentResponse{Agent: agent}
}
