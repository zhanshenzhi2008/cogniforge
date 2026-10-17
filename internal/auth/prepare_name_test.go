package auth_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cogniforge/internal/auth"
	"cogniforge/internal/database"
	"cogniforge/internal/model"
)

func TestPrepareUserNameUniqueness_PurgeSeedAndFreeSoftDeletedName(t *testing.T) {
	db := database.DB
	require.NoError(t, db.Unscoped().Where("1 = 1").Delete(&model.User{}).Error)

	keep := model.User{
		ID: "keep-admin", Email: "keep@example.com", Name: "admin",
		Password: "x", Role: "admin", Status: "active",
		CreatedAt: time.Now().Add(-time.Hour), UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(&keep).Error)

	seed := model.User{
		ID: "seed-admin", Email: "admin@cogniforge.local", Name: "seed-only-name",
		Password: "x", Role: "admin", Status: "active",
		CreatedAt: time.Now().Add(-2 * time.Hour), UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(&seed).Error)

	soft := model.User{
		ID: "soft-admin", Email: "soft@example.com", Name: "soft-held",
		Password: "x", Role: "user", Status: "active",
		CreatedAt: time.Now().Add(-30 * time.Minute), UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(&soft).Error)
	require.NoError(t, db.Delete(&soft).Error)

	require.NoError(t, auth.PrepareUserNameUniqueness(db))

	var seedCount int64
	db.Unscoped().Model(&model.User{}).Where("email = ?", "admin@cogniforge.local").Count(&seedCount)
	assert.Equal(t, int64(0), seedCount)

	var softRow model.User
	require.NoError(t, db.Unscoped().Where("id = ?", "soft-admin").First(&softRow).Error)
	assert.Contains(t, softRow.Name, "__del_")

	var kept model.User
	require.NoError(t, db.Where("id = ?", "keep-admin").First(&kept).Error)
	assert.Equal(t, "admin", kept.Name)
}
