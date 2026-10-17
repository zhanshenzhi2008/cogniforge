package prompt

import "cogniforge/internal/model"

type CreatePromptRequest struct {
	Name        string              `json:"name" binding:"required"`
	Description string              `json:"description"`
	Content     string              `json:"content"` // 初版正文，可空
	Changelog   string              `json:"changelog"`
	Variables   model.PromptVarDefs `json:"variables"`
	Tags        []string            `json:"tags"`
}

type UpdatePromptRequest struct {
	Name        *string              `json:"name"`
	Description *string              `json:"description"`
	Variables   *model.PromptVarDefs `json:"variables"`
	Tags        *[]string            `json:"tags"`
}

type CreateVersionRequest struct {
	Content   string `json:"content" binding:"required"`
	Changelog string `json:"changelog"`
	// Publish=true：直接发布并设为当前；默认 false 存为 draft
	Publish bool `json:"publish"`
}

type SetCurrentVersionRequest struct {
	VersionID string `json:"version_id" binding:"required"`
}

type RenderRequest struct {
	VersionID string            `json:"version_id"` // 空则用当前版
	Variables map[string]string `json:"variables"`
}

type RenderResponse struct {
	Content   string   `json:"content"`
	VarNames  []string `json:"var_names"`
	VersionID string   `json:"version_id"`
	Version   int      `json:"version"`
}

type ImportRequest struct {
	AgentID string `json:"agent_id"` // 从 Agent.system_prompt 导入
	Name    string `json:"name"`     // 可选；默认用 Agent 名
}

type PromptUsage struct {
	AgentID   string `json:"agent_id"`
	AgentName string `json:"agent_name"`
}

type PromptDetail struct {
	Prompt         interface{} `json:"prompt"`
	CurrentVersion interface{} `json:"current_version,omitempty"`
}
