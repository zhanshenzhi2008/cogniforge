package provider

import (
	"cogniforge/internal/model"
)

// ===================== 请求结构 =====================

type CreateProviderRequest struct {
	ID             string                 `json:"id"`                      // 可选，手动指定 ID
	Name           string                 `json:"name" binding:"required"` // 配置名称
	Provider       string                 `json:"provider" binding:"required"`
	BaseURL        string                 `json:"base_url"`
	APIKey         string                 `json:"api_key" binding:"required"`
	DefaultModel   string                 `json:"default_model"`
	EmbeddingModel string                 `json:"embedding_model"`
	Capabilities   []string               `json:"capabilities"`
	ExtraHeaders   map[string]interface{} `json:"extra_headers"`
	IsEnabled      bool                   `json:"is_enabled"`
	Priority       int                    `json:"priority"`
	Scope          string                 `json:"scope"` // private | public；非 admin 强制 private
}

type UpdateProviderRequest struct {
	Name           *string                `json:"name"`
	Provider       *string                `json:"provider"` // 厂商键。决定卡片小字（zhipu→智谱 BigModel）。可与 Base URL 分开改。
	BaseURL        *string                `json:"base_url"`
	APIKey         *string                `json:"api_key"`
	DefaultModel   *string                `json:"default_model"`
	EmbeddingModel *string                `json:"embedding_model"`
	Capabilities   []string               `json:"capabilities"`
	ExtraHeaders   map[string]interface{} `json:"extra_headers"`
	IsEnabled      *bool                  `json:"is_enabled"`
	Priority       *int                   `json:"priority"`
	Scope          *string                `json:"scope"` // 仅 admin 可改为 public
}

type SetDefaultRequest struct {
	Purpose string `json:"purpose"` // chat | embedding，空则对话
	// Personal=true：写入用户偏好，不改全站 is_default（普通用户默认走这条）
	Personal *bool `json:"personal,omitempty"`
}

// ===================== 响应结构 =====================

type ProviderResponse struct {
	ID                 string                 `json:"id"`
	Name               string                 `json:"name"`
	Provider           string                 `json:"provider"`
	BaseURL            string                 `json:"base_url"`
	DefaultModel       string                 `json:"default_model"`
	EmbeddingModel     string                 `json:"embedding_model"`
	Capabilities       []string               `json:"capabilities"`
	ExtraHeaders       map[string]interface{} `json:"extra_headers"`
	OwnerUserID        string                 `json:"owner_user_id"`
	Scope              string                 `json:"scope"`
	CanWrite           bool                   `json:"can_write"`
	IsPreferredChat    bool                   `json:"is_preferred_chat,omitempty"`
	IsPreferredEmbed   bool                   `json:"is_preferred_embedding,omitempty"`
	IsEnabled          bool                   `json:"is_enabled"`
	IsDefault          bool                   `json:"is_default"`
	IsDefaultEmbedding bool                   `json:"is_default_embedding"`
	Priority           int                    `json:"priority"`
	Status             string                 `json:"status"`
	LastTestAt         *string                `json:"last_test_at"`
	LastError          string                 `json:"last_error,omitempty"`
	CreatedAt          string                 `json:"created_at"`
	UpdatedAt          string                 `json:"updated_at"`
	// API Key 安全处理：创建/更新时返回，后续列表接口不返回原始 Key
	APIKey string `json:"api_key,omitempty"`
}

type TestResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	LatencyMs int64  `json:"latency_ms"`
}

// ===================== 辅助函数 =====================

func toResponse(p *model.AIProvider) ProviderResponse {
	return ToResponseFor(p, Actor{}, "", "")
}

// ToResponseFor 组装对外 Provider 响应（按 actor 脱敏）。
func ToResponseFor(p *model.AIProvider, actor Actor, prefChat, prefEmbed string) ProviderResponse {
	scope := p.Scope
	if scope == "" {
		scope = model.ProviderScopePublic
	}
	resp := ProviderResponse{
		ID:                 p.ID,
		Name:               p.Name,
		Provider:           p.Provider,
		BaseURL:            p.BaseURL,
		DefaultModel:       p.DefaultModel,
		EmbeddingModel:     p.EmbeddingModel,
		Capabilities:       model.SplitCapabilities(p.Capabilities),
		ExtraHeaders:       p.ExtraHeaders,
		OwnerUserID:        p.OwnerUserID,
		Scope:              scope,
		CanWrite:           model.ProviderWritableBy(p, actor.UserID, actor.IsAdmin),
		IsPreferredChat:    prefChat != "" && prefChat == p.ID,
		IsPreferredEmbed:   prefEmbed != "" && prefEmbed == p.ID,
		IsEnabled:          p.IsEnabled,
		IsDefault:          p.IsDefault,
		IsDefaultEmbedding: p.IsDefaultEmbedding,
		Priority:           p.Priority,
		Status:             p.Status,
		LastError:          p.LastError,
		CreatedAt:          p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:          p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if p.APIKey != "" {
		resp.APIKey = p.APIKey
	}
	if p.LastTestAt != nil {
		s := p.LastTestAt.Format("2006-01-02T15:04:05Z")
		resp.LastTestAt = &s
	}
	return resp
}

func toResponseList(providers []model.AIProvider) []ProviderResponse {
	result := make([]ProviderResponse, len(providers))
	for i, p := range providers {
		result[i] = toResponse(&p)
	}
	return result
}
