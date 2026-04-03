package database

import (
	"testing"

	"cogniforge/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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

func InitTestDBForPkg() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to open test database: " + err.Error())
	}

	if err := db.AutoMigrate(&model.User{}, &model.ApiKey{}); err != nil {
		panic("failed to migrate test database: " + err.Error())
	}

	DB = db
	return db
}

func GetTestDB() *gorm.DB {
	return DB
}
