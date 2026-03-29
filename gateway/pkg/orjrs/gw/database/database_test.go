package database

import (
	"testing"

	"github.com/orjrs/gateway/pkg/orjrs/gw/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// InitTestDB creates an in-memory SQLite database for testing.
// Call this at the start of each test that needs database access.
func InitTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	if err := db.AutoMigrate(&model.User{}, &model.ApiKey{}); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	DB = db
	return db
}
