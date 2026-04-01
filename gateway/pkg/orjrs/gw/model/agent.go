package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Agent represents an AI agent
type Agent struct {
	ID           string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID       string         `gorm:"type:varchar(64);not null;index" json:"user_id"`
	Name         string         `gorm:"type:varchar(255);not null" json:"name"`
	Description  string         `gorm:"type:text" json:"description"`
	Model        string         `gorm:"type:varchar(100);not null" json:"model"`
	SystemPrompt string         `gorm:"type:text" json:"system_prompt"`
	Tools        JSONBArray     `gorm:"type:jsonb;default:'[]'" json:"tools"`
	MemoryType   string         `gorm:"type:varchar(50);default:'short_term'" json:"memory_type"`
	MemoryTurns  int            `gorm:"type:int;default:10" json:"memory_turns"`
	InputFilter  bool           `gorm:"type:bool;default:true" json:"input_filter"`
	OutputFilter bool           `gorm:"type:bool;default:true" json:"output_filter"`
	Status       string         `gorm:"type:varchar(20);default:'active'" json:"status"`
	Metadata     JSONBMap       `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Agent) TableName() string {
	return "agents"
}

// JSONBArray is a custom type for handling JSONB arrays in GORM
type JSONBArray []string

func (j JSONBArray) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	return json.Marshal(j)
}

func (j *JSONBArray) Scan(value interface{}) error {
	if value == nil {
		*j = []string{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan JSONBArray")
	}

	return json.Unmarshal(bytes, j)
}

// JSONBMap is a custom type for handling JSONB objects in GORM
type JSONBMap map[string]interface{}

func (j JSONBMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	return json.Marshal(j)
}

func (j *JSONBMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(map[string]interface{})
		return nil
	}

	// Handle both []byte and string
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	default:
		return errors.New("failed to scan JSONBMap: unsupported type")
	}
}
