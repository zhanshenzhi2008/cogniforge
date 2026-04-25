package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户表
type User struct {
	ID        string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	Email     string         `gorm:"uniqueIndex;type:varchar(255);not null" json:"email"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	Password  string         `gorm:"type:varchar(255);not null" json:"-"`
	AvatarURL string         `gorm:"type:varchar(500)" json:"avatar_url"`             // 头像地址
	Status    string         `gorm:"type:varchar(50);default:'active'" json:"status"` // active, disabled, locked
	Role      string         `gorm:"type:varchar(100);default:'user'" json:"role"`    // 用户角色（简单方案）
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}

// UserSettings 用户设置表
type UserSettings struct {
	ID        string    `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID    string    `gorm:"type:varchar(64);not null;uniqueIndex" json:"user_id"`
	AvatarURL string    `gorm:"type:varchar(500)" json:"avatar_url"`           // 头像地址
	Theme     string    `gorm:"type:varchar(50);default:'light'" json:"theme"` // 主题：light/dark
	Language  string    `gorm:"type:varchar(20);default:'zh-CN'" json:"language"`
	Timezone  string    `gorm:"type:varchar(50);default:'Asia/Shanghai'" json:"timezone"`
	Metadata  JSONBMap  `gorm:"type:jsonb" json:"metadata"` // 其他设置（通知偏好等）
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (UserSettings) TableName() string {
	return "user_settings"
}

// UserSession 用户会话表（记录登录设备）
type UserSession struct {
	ID        string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID    string         `gorm:"type:varchar(64);not null;index:idx_user_active,priority:1" json:"user_id"`
	TokenID   string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"token_id"`         // JWT ID (jti)
	UserAgent string         `gorm:"type:varchar(500)" json:"user_agent"`                            // 用户代理
	IPAddress string         `gorm:"type:varchar(50)" json:"ip_address"`                             // IP 地址
	Device    string         `gorm:"type:varchar(100)" json:"device"`                                // 设备信息
	Location  string         `gorm:"type:varchar(255)" json:"location"`                              // 登录地点
	ExpiresAt time.Time      `json:"expires_at"`                                                     // 过期时间
	LastUsed  time.Time      `json:"last_used"`                                                      // 最后使用时间
	IsActive  bool           `gorm:"default:true;index:idx_user_active,priority:2" json:"is_active"` // 是否活跃
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (UserSession) TableName() string {
	return "user_sessions"
}

// ApiKey API Key 表
type ApiKey struct {
	ID        string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID    string         `gorm:"type:varchar(64);not null;index" json:"user_id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	Key       string         `gorm:"type:varchar(255);not null" json:"key"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ApiKey) TableName() string {
	return "api_keys"
}
