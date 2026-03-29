package database

import (
	"github.com/orjrs/gateway/pkg/orjrs/gw/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// InitTestDBForPkg creates an in-memory SQLite database for use in TestMain.
// Does not take *testing.T so it can be called from TestMain in any test package.
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
