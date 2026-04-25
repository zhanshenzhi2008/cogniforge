package handler_test

import (
	"testing"

	"cogniforge/internal/database"
	"cogniforge/internal/model"
)

func TestMain(m *testing.M) {
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
	)
	m.Run()
}
