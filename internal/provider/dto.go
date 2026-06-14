package provider

import (
	"cogniforge/internal/model"
)

// ===================== 请求结构 =====================

type CreateProviderRequest struct {
	ID           string                 `json:"id"`                      // 可选，手动指定 ID
	Name         string                 `json:"name" binding:"required"` // 配置名称
	Provider     string                 `json:"provider" binding:"required"`
	BaseURL      string                 `json:"base_url"`
	APIKey       string                 `json:"api_key" binding:"required"`
	DefaultModel string                 `json:"default_model"`
	ExtraHeaders map[string]interface{} `json:"extra_headers"`
	IsEnabled    bool                   `json:"is_enabled"`
	Priority     int                    `json:"priority"`
}

type UpdateProviderRequest struct {
	Name         *string                `json:"name"`
	BaseURL      *string                `json:"base_url"`
	APIKey       *string                `json:"api_key"`
	DefaultModel *string                `json:"default_model"`
	ExtraHeaders map[string]interface{} `json:"extra_headers"`
	IsEnabled    *bool                  `json:"is_enabled"`
	Priority     *int                   `json:"priority"`
}

// ===================== 响应结构 =====================

type ProviderResponse struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Provider     string                 `json:"provider"`
	BaseURL      string                 `json:"base_url"`
	DefaultModel string                 `json:"default_model"`
	ExtraHeaders map[string]interface{} `json:"extra_headers"`
	IsEnabled    bool                   `json:"is_enabled"`
	IsDefault    bool                   `json:"is_default"`
	Priority     int                    `json:"priority"`
	Status       string                 `json:"status"`
	LastTestAt   *string                `json:"last_test_at"`
	LastError    string                 `json:"last_error,omitempty"`
	CreatedAt    string                 `json:"created_at"`
	UpdatedAt    string                 `json:"updated_at"`
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
	resp := ProviderResponse{
		ID:           p.ID,
		Name:         p.Name,
		Provider:     p.Provider,
		BaseURL:      p.BaseURL,
		DefaultModel: p.DefaultModel,
		ExtraHeaders: p.ExtraHeaders,
		IsEnabled:    p.IsEnabled,
		IsDefault:    p.IsDefault,
		Priority:     p.Priority,
		Status:       p.Status,
		LastError:    p.LastError,
		CreatedAt:    p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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
