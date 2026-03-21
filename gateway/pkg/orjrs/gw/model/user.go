package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	Email     string         `gorm:"uniqueIndex;type:varchar(255);not null" json:"email"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	Password  string         `gorm:"type:varchar(255);not null" json:"-"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}
