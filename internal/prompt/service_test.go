package prompt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"cogniforge/internal/model"
)

func setupPromptDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Prompt{}, &model.PromptVersion{}))
	return db
}

func TestPromptCreate_AndNewVersion(t *testing.T) {
	db := setupPromptDB(t)
	svc := NewService(db)

	p, v1, err := svc.Create("u1", &CreatePromptRequest{
		Name:    "客服",
		Content: "你是客服 v1",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, v1.Version)
	assert.Equal(t, v1.ID, p.CurrentVersionID)

	v2, err := svc.CreateVersion("u1", p.ID, &CreateVersionRequest{
		Content:   "你是客服 v2，更简洁",
		Changelog: "收紧语气",
	})
	require.NoError(t, err)
	assert.Equal(t, 2, v2.Version)

	got, cur, err := svc.Get("u1", p.ID)
	require.NoError(t, err)
	assert.Equal(t, v2.ID, got.CurrentVersionID)
	require.NotNil(t, cur)
	assert.Contains(t, cur.Content, "v2")

	vers, err := svc.ListVersions("u1", p.ID)
	require.NoError(t, err)
	require.Len(t, vers, 2)
}

func TestPrompt_UserIsolation(t *testing.T) {
	db := setupPromptDB(t)
	svc := NewService(db)
	p, _, err := svc.Create("u1", &CreatePromptRequest{Name: "A", Content: "x"})
	require.NoError(t, err)
	_, _, err = svc.Get("u2", p.ID)
	require.Error(t, err)
}
