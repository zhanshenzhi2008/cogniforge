package model

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// AIProvider AI供应商配置表
type AIProvider struct {
	ID                 string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	Name               string         `gorm:"type:varchar(100);not null" json:"name"`            // 配置名称
	Provider           string         `gorm:"type:varchar(50);not null;index" json:"provider"`   // 厂商键，卡片小字用它：openai | zhipu | deepseek | custom 等。聊天请求仍走 OpenAI 兼容路径。
	BaseURL            string         `gorm:"type:varchar(500)" json:"base_url"`                 // API端点
	APIKey             string         `gorm:"type:varchar(500);not null" json:"api_key"`         // API密钥
	DefaultModel       string         `gorm:"type:varchar(100)" json:"default_model"`            // 对话默认模型
	EmbeddingModel     string         `gorm:"type:varchar(100)" json:"embedding_model"`          // 向量模型；空则对话模型不当向量用
	Capabilities       string         `gorm:"type:varchar(64);default:chat" json:"capabilities"` // chat,embedding
	ExtraHeaders       JSONBMap       `gorm:"type:jsonb" json:"extra_headers"`                   // 额外请求头（如OpenRouter的HTTP-Referer）
	IsEnabled          bool           `gorm:"default:false" json:"is_enabled"`                   // 是否启用
	IsDefault          bool           `gorm:"default:false;index" json:"is_default"`             // 对话默认
	IsDefaultEmbedding bool           `gorm:"default:false;index" json:"is_default_embedding"`   // 向量默认
	Priority           int            `gorm:"default:0" json:"priority"`                         // 优先级
	Status             string         `gorm:"type:varchar(20);default:'active'" json:"status"`   // active | error | testing
	LastTestAt         *time.Time     `json:"last_test_at"`                                      // 上次测试时间
	LastError          string         `gorm:"type:text" json:"last_error"`                       // 上次错误信息
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AIProvider) TableName() string {
	return "ai_providers"
}

// ProviderType 供应商类型常量
type ProviderType string

const (
	ProviderOpenAI      ProviderType = "openai"
	ProviderAnthropic   ProviderType = "anthropic"
	ProviderOpenRouter  ProviderType = "openrouter"
	ProviderAzure       ProviderType = "azure"
	ProviderGemini      ProviderType = "gemini"
	ProviderSiliconGlow ProviderType = "siliconeglow"
	ProviderDeepSeek    ProviderType = "deepseek"
	ProviderZhipu       ProviderType = "zhipu" // 智谱 BigModel（OpenAI 兼容 /v4）
	ProviderGroq        ProviderType = "groq"
	ProviderOllama      ProviderType = "ollama"
	ProviderAICore      ProviderType = "aicore"
	ProviderCustom      ProviderType = "custom" // 列表外的 OpenAI 兼容端点，卡片小字为「自定义兼容」
)

// DefaultProviders 返回系统默认供应商配置（不含API Key）
var DefaultProviders = []AIProvider{
	{
		ID:             "openai",
		Provider:       string(ProviderOpenAI),
		Name:           "OpenAI",
		BaseURL:        "https://api.openai.com/v1",
		DefaultModel:   "gpt-4o",
		EmbeddingModel: "text-embedding-3-small",
		Capabilities:   CapChat + "," + CapEmbedding,
		IsEnabled:      false,
		IsDefault:      false,
		Priority:       10,
		Status:         "active",
	},
	{
		ID:           "anthropic",
		Provider:     string(ProviderAnthropic),
		Name:         "Anthropic",
		BaseURL:      "https://api.anthropic.com",
		DefaultModel: "claude-3-5-sonnet-20241022",
		Capabilities: CapChat,
		IsEnabled:    false,
		IsDefault:    false,
		Priority:     20,
		Status:       "active",
	},
	{
		ID:             "openrouter",
		Provider:       string(ProviderOpenRouter),
		Name:           "OpenRouter",
		BaseURL:        "https://openrouter.ai/api/v1",
		DefaultModel:   "openai/gpt-4o",
		EmbeddingModel: "openai/text-embedding-3-small",
		Capabilities:   CapChat + "," + CapEmbedding,
		IsEnabled:      false,
		IsDefault:      false,
		Priority:       30,
		Status:         "active",
	},
	{
		ID:             "siliconeglow",
		Provider:       string(ProviderSiliconGlow),
		Name:           "硅基流动 SiliconGlow",
		BaseURL:        "https://api.siliconflow.cn/v1",
		DefaultModel:   "Qwen/Qwen2.5-7B-Instruct",
		EmbeddingModel: "BAAI/bge-m3",
		Capabilities:   CapChat + "," + CapEmbedding,
		IsEnabled:      false,
		IsDefault:      false,
		Priority:       40,
		Status:         "active",
	},
	{
		ID:           "deepseek",
		Provider:     string(ProviderDeepSeek),
		Name:         "DeepSeek",
		BaseURL:      "https://api.deepseek.com/v1",
		DefaultModel: "deepseek-chat",
		Capabilities: CapChat,
		IsEnabled:    false,
		IsDefault:    false,
		Priority:     50,
		Status:       "active",
	},
	{
		ID:             "zhipu",
		Provider:       string(ProviderZhipu),
		Name:           "智谱 BigModel",
		BaseURL:        "https://open.bigmodel.cn/api/paas/v4",
		DefaultModel:   "glm-4.7-flash", // 平台标注免费；资源包模型另选
		EmbeddingModel: "embedding-3",
		Capabilities:   CapChat + "," + CapEmbedding,
		IsEnabled:      false,
		IsDefault:      false,
		Priority:       45,
		Status:         "active",
	},
	{
		ID:           "groq",
		Provider:     string(ProviderGroq),
		Name:         "Groq",
		BaseURL:      "https://api.groq.com/openai/v1",
		DefaultModel: "llama-3.3-70b-versatile",
		Capabilities: CapChat,
		IsEnabled:    false,
		IsDefault:    false,
		Priority:     15,
		Status:       "active",
	},
	{
		ID:           "ollama",
		Provider:     string(ProviderOllama),
		Name:         "Ollama (本地)",
		BaseURL:      "http://localhost:11434/v1",
		DefaultModel: "llama3",
		Capabilities: CapChat,
		IsEnabled:    false,
		IsDefault:    false,
		Priority:     5,
		Status:       "active",
	},
	{
		ID:             "aicore",
		Provider:       string(ProviderAICore),
		Name:           "AI Core",
		BaseURL:        "https://api.xty.app/v1",
		DefaultModel:   "gpt-4o",
		EmbeddingModel: "text-embedding-3-small",
		Capabilities:   CapChat + "," + CapEmbedding,
		IsEnabled:      false,
		IsDefault:      false,
		Priority:       5,
		Status:         "active",
	},
}

// MigrateProviders 迁移 AIProvider 表
func MigrateProviders(db *gorm.DB) error {
	if err := db.AutoMigrate(&AIProvider{}); err != nil {
		return err
	}
	return db.Model(&AIProvider{}).Where("capabilities IS NULL OR capabilities = ?", "").Update("capabilities", CapChat).Error
}

// InitDefaultProviders 初始化默认供应商记录（如果不存在）
func InitDefaultProviders(db *gorm.DB) error {
	for _, p := range DefaultProviders {
		var existing AIProvider
		err := db.Where("id = ?", p.ID).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&p).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// RestoreSeededDeepSeekChatDefault 撤销误把系统自带 DeepSeek 默认改成 V4 的操作；自定义供应商不动。
func RestoreSeededDeepSeekChatDefault(db *gorm.DB) error {
	return db.Model(&AIProvider{}).
		Where("id = ? AND default_model IN ?", "deepseek", []string{"deepseek-v4-flash", "deepseek-v4-pro"}).
		Update("default_model", "deepseek-chat").Error
}

// VendorURLHint 用 Base URL 里的主机片段认出厂商键。顺序：更具体的放前面。
type VendorURLHint struct {
	Host     string
	Provider string
}

// VendorURLHints 与前端 PROVIDER_META.hostIncludes 保持一致。
func VendorURLHints() []VendorURLHint {
	return []VendorURLHint{
		{Host: "open.bigmodel.cn", Provider: string(ProviderZhipu)},
		{Host: "api.deepseek.com", Provider: string(ProviderDeepSeek)},
		{Host: "api.siliconflow.cn", Provider: string(ProviderSiliconGlow)},
		{Host: "openrouter.ai", Provider: string(ProviderOpenRouter)},
		{Host: "api.anthropic.com", Provider: string(ProviderAnthropic)},
		{Host: "api.groq.com", Provider: string(ProviderGroq)},
		{Host: "generativelanguage.googleapis.com", Provider: string(ProviderGemini)},
		{Host: "openai.azure.com", Provider: string(ProviderAzure)},
		{Host: "localhost:11434", Provider: string(ProviderOllama)},
		{Host: "127.0.0.1:11434", Provider: string(ProviderOllama)},
		{Host: "api.xty.app", Provider: string(ProviderAICore)},
		{Host: "api.openai.com", Provider: string(ProviderOpenAI)},
	}
}

// VendorFromBaseURL 从地址认出厂商键。认不出时返回空字符串。
func VendorFromBaseURL(base string) string {
	b := strings.ToLower(strings.TrimSpace(base))
	if b == "" {
		return ""
	}
	for _, h := range VendorURLHints() {
		if h.Host != "" && strings.Contains(b, h.Host) {
			return h.Provider
		}
	}
	return ""
}

// CatalogModels 某类供应商在对话下拉里应出现的模型 ID（免费/便宜的排前面）。
func CatalogModels(provider string) []string {
	switch provider {
	case string(ProviderDeepSeek):
		return []string{"deepseek-chat", "deepseek-reasoner", "deepseek-v4-flash", "deepseek-v4-pro"}
	case string(ProviderZhipu):
		// 免费 Flash 优先；其后为资源包/付费常用文本与视觉
		return []string{
			"glm-4.7-flash",
			"glm-4.5-flash",
			"glm-4-flash-250414",
			"glm-4.6v-flash",
			"glm-4.5-air",
			"glm-4.6",
			"glm-4.7",
			"glm-5.3-flash",
		}
	default:
		return nil
	}
}
