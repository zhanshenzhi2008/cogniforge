package user_test

import (
	"os"
	"testing"

	"cogniforge/internal/database"
	"cogniforge/internal/model"
)

func TestMain(m *testing.M) {
	// Initialize test database
	db := database.InitTestDBForPkg()
	database.DB = db
	db.AutoMigrate(
		&model.User{},
		&model.UserSettings{},
		&model.UserSession{},
		&model.ApiKey{},
		&model.Agent{},
		&model.Workflow{},
		&model.KnowledgeBase{},
		&model.Document{},
		&model.Permission{},
		&model.Role{},
		&model.RolePermission{},
	)
	os.Exit(m.Run())
}
