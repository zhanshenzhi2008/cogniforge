package model

import (
	"time"

	"gorm.io/gorm"
)

// AIProvider AI供应商配置表
type AIProvider struct {
	ID           string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	Name         string         `gorm:"type:varchar(100);not null" json:"name"`          // 配置名称
	Provider     string         `gorm:"type:varchar(50);not null;index" json:"provider"` // openai | anthropic | openrouter | azure | gemini | siliconeglow | deepseek
	BaseURL      string         `gorm:"type:varchar(500)" json:"base_url"`               // API端点
	APIKey       string         `gorm:"type:varchar(500);not null" json:"api_key"`       // API密钥
	DefaultModel string         `gorm:"type:varchar(100)" json:"default_model"`          // 默认模型
	ExtraHeaders JSONBMap       `gorm:"type:jsonb" json:"extra_headers"`                 // 额外请求头（如OpenRouter的HTTP-Referer）
	IsEnabled    bool           `gorm:"default:false" json:"is_enabled"`                 // 是否启用
	IsDefault    bool           `gorm:"default:false;index" json:"is_default"`           // 是否默认
	Priority     int            `gorm:"default:0" json:"priority"`                       // 优先级
	Status       string         `gorm:"type:varchar(20);default:'active'" json:"status"` // active | error | testing
	LastTestAt   *time.Time     `json:"last_test_at"`                                    // 上次测试时间
	LastError    string         `gorm:"type:text" json:"last_error"`                     // 上次错误信息
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
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
	ProviderGroq        ProviderType = "groq"
	ProviderOllama      ProviderType = "ollama"
)

// DefaultProviders 返回系统默认供应商配置（不含API Key）
var DefaultProviders = []AIProvider{
	{
		ID:           "openai",
		Provider:     string(ProviderOpenAI),
		Name:         "OpenAI",
		BaseURL:      "https://api.openai.com/v1",
		DefaultModel: "gpt-4o",
		IsEnabled:    false,
		IsDefault:    false,
		Priority:     10,
		Status:       "active",
	},
	{
		ID:           "anthropic",
		Provider:     string(ProviderAnthropic),
		Name:         "Anthropic",
		BaseURL:      "https://api.anthropic.com",
		DefaultModel: "claude-3-5-sonnet-20241022",
		IsEnabled:    false,
		IsDefault:    false,
		Priority:     20,
		Status:       "active",
	},
	{
		ID:           "openrouter",
		Provider:     string(ProviderOpenRouter),
		Name:         "OpenRouter",
		BaseURL:      "https://openrouter.ai/api/v1",
		DefaultModel: "openai/gpt-4o",
		IsEnabled:    false,
		IsDefault:    false,
		Priority:     30,
		Status:       "active",
	},
	{
		ID:           "siliconeglow",
		Provider:     string(ProviderSiliconGlow),
		Name:         "硅基流动 SiliconGlow",
		BaseURL:      "https://api.siliconflow.cn/v1",
		DefaultModel: "Qwen/Qwen2.5-7B-Instruct",
		IsEnabled:    false,
		IsDefault:    false,
		Priority:     40,
		Status:       "active",
	},
	{
		ID:           "deepseek",
		Provider:     string(ProviderDeepSeek),
		Name:         "DeepSeek",
		BaseURL:      "https://api.deepseek.com/v1",
		DefaultModel: "deepseek-chat",
		IsEnabled:    false,
		IsDefault:    false,
		Priority:     50,
		Status:       "active",
	},
	{
		ID:           "groq",
		Provider:     string(ProviderGroq),
		Name:         "Groq",
		BaseURL:      "https://api.groq.com/openai/v1",
		DefaultModel: "llama-3.3-70b-versatile",
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
		IsEnabled:    false,
		IsDefault:    false,
		Priority:     5,
		Status:       "active",
	},
	{
		ID:           "aicore",
		Provider:     string(ProviderOpenAI),
		Name:         "AI Core",
		BaseURL:      "https://api.xty.app/v1",
		DefaultModel: "gpt-4o",
		IsEnabled:    false,
		IsDefault:    false,
		Priority:     5,
		Status:       "active",
	},
}

// MigrateProviders 迁移 AIProvider 表
func MigrateProviders(db *gorm.DB) error {
	return db.AutoMigrate(&AIProvider{})
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
