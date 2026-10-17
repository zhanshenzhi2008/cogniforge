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
	require.NoError(t, db.AutoMigrate(&model.Prompt{}, &model.PromptVersion{}, &model.Agent{}))
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
	assert.Equal(t, model.PromptVersionPublished, v1.Status)
	assert.Equal(t, v1.ID, p.CurrentVersionID)

	v2, err := svc.CreateVersion("u1", p.ID, &CreateVersionRequest{
		Content:   "你是客服 v2，更简洁",
		Changelog: "收紧语气",
		Publish:   true,
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

func TestPrompt_DraftPublish_AndSetCurrent(t *testing.T) {
	db := setupPromptDB(t)
	svc := NewService(db)

	p, v1, err := svc.Create("u1", &CreatePromptRequest{Name: "客服", Content: "v1"})
	require.NoError(t, err)

	draft, err := svc.CreateVersion("u1", p.ID, &CreateVersionRequest{Content: "draft"})
	require.NoError(t, err)
	assert.Equal(t, model.PromptVersionDraft, draft.Status)
	p, _, err = svc.Get("u1", p.ID)
	require.NoError(t, err)
	assert.Equal(t, v1.ID, p.CurrentVersionID)

	_, err = svc.SetCurrentVersion("u1", p.ID, draft.ID)
	require.Error(t, err)

	_, pub, err := svc.PublishVersion("u1", draft.ID)
	require.NoError(t, err)
	assert.Equal(t, model.PromptVersionPublished, pub.Status)
	p, _, err = svc.Get("u1", p.ID)
	require.NoError(t, err)
	assert.Equal(t, draft.ID, p.CurrentVersionID)
}

func TestPrompt_DeleteVersion(t *testing.T) {
	db := setupPromptDB(t)
	svc := NewService(db)

	p, v1, err := svc.Create("u1", &CreatePromptRequest{Name: "客服", Content: "v1"})
	require.NoError(t, err)
	v2, err := svc.CreateVersion("u1", p.ID, &CreateVersionRequest{Content: "v2", Publish: true})
	require.NoError(t, err)

	_, vOnly, err := svc.Create("u1", &CreatePromptRequest{Name: "单版", Content: "only"})
	require.NoError(t, err)
	require.Error(t, svc.DeleteVersion("u1", vOnly.ID))

	require.NoError(t, svc.DeleteVersion("u1", v2.ID))
	got, cur, err := svc.Get("u1", p.ID)
	require.NoError(t, err)
	assert.Equal(t, v1.ID, got.CurrentVersionID)
	require.NotNil(t, cur)
	assert.Equal(t, 1, cur.Version)
}

func TestPrompt_Delete(t *testing.T) {
	db := setupPromptDB(t)
	svc := NewService(db)
	p, _, err := svc.Create("u1", &CreatePromptRequest{Name: "A", Content: "x"})
	require.NoError(t, err)
	_, err = svc.CreateVersion("u1", p.ID, &CreateVersionRequest{Content: "y", Publish: true})
	require.NoError(t, err)
	require.NoError(t, svc.Delete("u1", p.ID))
	_, _, err = svc.Get("u1", p.ID)
	require.Error(t, err)
	list, err := svc.List("u1")
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestPrompt_DuplicateName(t *testing.T) {
	db := setupPromptDB(t)
	svc := NewService(db)
	_, _, err := svc.Create("u1", &CreatePromptRequest{Name: "客服", Content: "a"})
	require.NoError(t, err)
	_, _, err = svc.Create("u1", &CreatePromptRequest{Name: "客服", Content: "b"})
	require.Error(t, err)
	// 不同用户可同名
	_, _, err = svc.Create("u2", &CreatePromptRequest{Name: "客服", Content: "c"})
	require.NoError(t, err)
}

func TestPrompt_Import_SameAgent_ReimportsAsNewVersion(t *testing.T) {
	db := setupPromptDB(t)
	svc := NewService(db)
	ag := model.Agent{
		ID: "a1", UserID: "u1", Name: "助手",
		SystemPrompt: "第一版正文", Status: "active",
	}
	require.NoError(t, db.Create(&ag).Error)

	p1, v1, err := svc.ImportFromAgent("u1", &ImportRequest{AgentID: "a1"})
	require.NoError(t, err)
	assert.Equal(t, 1, v1.Version)
	assert.Equal(t, model.PromptSourceAgent, p1.SourceType)
	require.NotNil(t, p1.SourceID)
	assert.Equal(t, "a1", *p1.SourceID)
	assert.Equal(t, "助手", p1.SourceName)

	ag.SystemPrompt = "第二版正文"
	require.NoError(t, db.Save(&ag).Error)
	p2, v2, err := svc.ImportFromAgent("u1", &ImportRequest{AgentID: "a1"})
	require.NoError(t, err)
	assert.Equal(t, p1.ID, p2.ID, "同一 Agent 应落到同一条提示词")
	assert.Equal(t, 2, v2.Version)
	assert.Contains(t, v2.Content, "第二版")

	list, err := svc.List("u1")
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestPrompt_Delete_FreesName(t *testing.T) {
	db := setupPromptDB(t)
	svc := NewService(db)
	p, _, err := svc.Create("u1", &CreatePromptRequest{Name: "可复用", Content: "x"})
	require.NoError(t, err)
	require.NoError(t, svc.Delete("u1", p.ID))
	_, _, err = svc.Create("u1", &CreatePromptRequest{Name: "可复用", Content: "y"})
	require.NoError(t, err)
}

func TestPrompt_Render_AndImport(t *testing.T) {
	db := setupPromptDB(t)
	svc := NewService(db)

	p, _, err := svc.Create("u1", &CreatePromptRequest{
		Name:    "带变量",
		Content: "你好 {{brand}}",
		Variables: model.PromptVarDefs{
			{Name: "brand", Default: "CogniForge", Required: true},
		},
	})
	require.NoError(t, err)

	out, err := svc.Render("u1", p.ID, &RenderRequest{})
	require.NoError(t, err)
	assert.Equal(t, "你好 CogniForge", out.Content)

	ag := model.Agent{
		ID: "a1", UserID: "u1", Name: "助手",
		SystemPrompt: "你是导入来的助手", Status: "active",
	}
	require.NoError(t, db.Create(&ag).Error)
	ip, iv, err := svc.ImportFromAgent("u1", &ImportRequest{AgentID: "a1"})
	require.NoError(t, err)
	assert.Contains(t, iv.Content, "导入来的")
	assert.NotEmpty(t, ip.ID)

	ag.PromptID = p.ID
	require.NoError(t, db.Save(&ag).Error)
	usages, err := svc.ListUsages("u1", p.ID)
	require.NoError(t, err)
	require.Len(t, usages, 1)
	assert.Equal(t, "a1", usages[0].AgentID)
}
