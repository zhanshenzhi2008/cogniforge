package database

import (
	"fmt"
	"log/slog"
	"testing"

	"cogniforge/internal/config"
	"cogniforge/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	slog.Info("database connected successfully")
	DB = db
	return db, nil
}

// InitTestDB creates an in-memory SQLite database for testing
func InitTestDB(t testing.TB) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	DB = db
	return db
}

// InitTestDBForPkg creates an in-memory SQLite database for package-level testing
func InitTestDBForPkg() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic("failed to open test database: " + err.Error())
	}

	DB = db
	return db
}

// GetTestDB returns the current test database instance
func GetTestDB() *gorm.DB {
	return DB
}

// MigrateTestDB migrates the test database with common models
func MigrateTestDB(db *gorm.DB) {
	db.AutoMigrate(&model.User{}, &model.ApiKey{}, &model.UserSession{})
}
