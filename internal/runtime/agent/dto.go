package agent

import "cogniforge/internal/model"

// ============ 请求结构 ============

type CreateAgentRequest struct {
	Name             string   `json:"name" binding:"required"`
	Description      string   `json:"description"`
	Model            string   `json:"model" binding:"required"`
	SystemPrompt     string   `json:"system_prompt"`
	PromptID         string   `json:"prompt_id"` // 绑定提示词资产（优先于内联 system_prompt）
	OpeningMessage      string   `json:"opening_message"`
	SuggestedPrompts    []string `json:"suggested_prompts"`
	RAGKnowledgeBaseIDs []string `json:"rag_kb_ids"`
	AllowAttachments    bool     `json:"allow_attachments"`
	Tools               []string `json:"tools"`
	SkillID             string   `json:"skill_id"` // 阶段十五：从 SKILL 创建
	McpServers          []string `json:"mcp_servers"`
}

type UpdateAgentRequest struct {
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	Model               string    `json:"model"`
	SystemPrompt        string    `json:"system_prompt"`
	PromptID            *string   `json:"prompt_id"` // nil=不改；""=解绑
	OpeningMessage      *string   `json:"opening_message"`   // nil=不改；""=清空
	SuggestedPrompts    *[]string `json:"suggested_prompts"` // nil=不改；[]=清空
	RAGKnowledgeBaseIDs *[]string `json:"rag_kb_ids"`        // nil=不改；[]=清空
	AllowAttachments    *bool     `json:"allow_attachments"` // nil=不改
	Tools               []string  `json:"tools"`
	SkillID             string    `json:"skill_id"`
	McpServers          []string  `json:"mcp_servers"`
	Status              string    `json:"status"`
}

// ============ 响应结构 ============

type AgentResponse struct {
	Agent *model.Agent `json:"agent"`
}

// ============ 辅助函数 ============

func ToAgentResponse(agent *model.Agent) *AgentResponse {
	return &AgentResponse{Agent: agent}
}
