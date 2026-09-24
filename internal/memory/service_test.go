package memory

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"cogniforge/internal/model"
)

func setupDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(&model.ChatMemory{})
	require.NoError(t, err)
	return db
}

func TestList_Empty(t *testing.T) {
	db := setupDB(t)
	svc := NewMemoryService(db)
	ctx := context.Background()

	rows, err := svc.List(ctx, "user1", "", "", 50)
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestList_WithData(t *testing.T) {
	db := setupDB(t)
	svc := NewMemoryService(db)
	ctx := context.Background()
	uid := "user1"

	now := time.Now()
	m := &model.ChatMemory{
		ID:         "mem1",
		UserID:     uid,
		Kind:       "profile",
		Content:    "我叫小明",
		Importance: 7,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	require.NoError(t, svc.Upsert(ctx, m))

	rows, err := svc.List(ctx, uid, "", "", 50)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "mem1", rows[0].ID)
	assert.Equal(t, "profile", rows[0].Kind)
}

func TestList_FilterByKind(t *testing.T) {
	db := setupDB(t)
	svc := NewMemoryService(db)
	ctx := context.Background()
	uid := "user1"

	now := time.Now()
	for i, kind := range []string{"profile", "preference", "profile"} {
		m := &model.ChatMemory{
			ID:         fmt.Sprintf("mem-%s-%d", kind, i),
			UserID:     uid,
			Kind:       kind,
			Content:    kind,
			Importance: 5,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		require.NoError(t, svc.Upsert(ctx, m))
	}

	rows, err := svc.List(ctx, uid, "profile", "", 50)
	require.NoError(t, err)
	assert.Len(t, rows, 2)
}

func TestList_UserIsolation(t *testing.T) {
	db := setupDB(t)
	svc := NewMemoryService(db)
	ctx := context.Background()

	now := time.Now()
	for _, uid := range []string{"user1", "user2"} {
		m := &model.ChatMemory{
			ID:         "mem-" + uid,
			UserID:     uid,
			Kind:       "profile",
			Content:    "content for " + uid,
			Importance: 5,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		require.NoError(t, svc.Upsert(ctx, m))
	}

	rows, err := svc.List(ctx, "user1", "", "", 50)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "user1", rows[0].UserID)
}

func TestGet_Found(t *testing.T) {
	db := setupDB(t)
	svc := NewMemoryService(db)
	ctx := context.Background()
	uid := "user1"

	now := time.Now()
	m := &model.ChatMemory{
		ID:         "mem-get",
		UserID:     uid,
		Kind:       "preference",
		Content:    "回复要简短",
		Importance: 8,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	require.NoError(t, svc.Upsert(ctx, m))

	got, err := svc.Get(ctx, uid, "mem-get")
	require.NoError(t, err)
	assert.Equal(t, "preference", got.Kind)
	assert.Equal(t, "回复要简短", got.Content)
}

func TestGet_NotFound(t *testing.T) {
	db := setupDB(t)
	svc := NewMemoryService(db)
	ctx := context.Background()

	_, err := svc.Get(ctx, "user1", "not-exist")
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestGet_WrongUser(t *testing.T) {
	db := setupDB(t)
	svc := NewMemoryService(db)
	ctx := context.Background()
	now := time.Now()

	m := &model.ChatMemory{
		ID:         "mem-cross",
		UserID:     "user1",
		Kind:       "profile",
		Content:    "only for user1",
		Importance: 5,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	require.NoError(t, svc.Upsert(ctx, m))

	_, err := svc.Get(ctx, "user2", "mem-cross")
	assert.Error(t, err)
}

func TestDelete_SoftDelete(t *testing.T) {
	db := setupDB(t)
	svc := NewMemoryService(db)
	ctx := context.Background()
	uid := "user1"
	now := time.Now()

	m := &model.ChatMemory{
		ID:         "mem-del",
		UserID:     uid,
		Kind:       "decision",
		Content:    "用 Redis 缓存",
		Importance: 5,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	require.NoError(t, svc.Upsert(ctx, m))

	err := svc.Delete(ctx, uid, "mem-del")
	require.NoError(t, err)

	// 软删除后查不到
	_, err = svc.Get(ctx, uid, "mem-del")
	assert.Error(t, err)

	// 但数据库里还在（gorm soft delete）
	var count int64
	db.Unscoped().Where("id = ?", "mem-del").Model(&model.ChatMemory{}).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestDelete_WrongUser_NoError(t *testing.T) {
	db := setupDB(t)
	svc := NewMemoryService(db)
	ctx := context.Background()
	now := time.Now()

	m := &model.ChatMemory{
		ID:         "mem-del2",
		UserID:     "user1",
		Kind:       "episode",
		Content:    "content",
		Importance: 5,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	require.NoError(t, svc.Upsert(ctx, m))

	// user2 无权删除
	err := svc.Delete(ctx, "user2", "mem-del2")
	assert.NoError(t, err) // 不报错但未删

	got, err := svc.Get(ctx, "user1", "mem-del2")
	require.NoError(t, err)
	assert.Equal(t, "mem-del2", got.ID)
}

func TestClearAll(t *testing.T) {
	db := setupDB(t)
	svc := NewMemoryService(db)
	ctx := context.Background()
	uid := "user1"
	now := time.Now()

	for i := 0; i < 5; i++ {
		m := &model.ChatMemory{
			ID:         "mem-clear" + string(rune('a'+i)),
			UserID:     uid,
			Kind:       "profile",
			Content:    "content",
			Importance: 5,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		require.NoError(t, svc.Upsert(ctx, m))
	}
	// 另一个用户的记忆
	m2 := &model.ChatMemory{
		ID:         "mem-other",
		UserID:    "user2",
		Kind:      "profile",
		Content:   "other",
		Importance: 5,
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, svc.Upsert(ctx, m2))

	err := svc.ClearAll(ctx, uid)
	require.NoError(t, err)

	rows, err := svc.List(ctx, uid, "", "", 50)
	require.NoError(t, err)
	assert.Empty(t, rows)

	// user2 不受影响
	rows2, err := svc.List(ctx, "user2", "", "", 50)
	require.NoError(t, err)
	assert.Len(t, rows2, 1)
}

func TestUpsert_UpdateExisting(t *testing.T) {
	db := setupDB(t)
	svc := NewMemoryService(db)
	ctx := context.Background()
	uid := "user1"
	now := time.Now()

	m1 := &model.ChatMemory{
		ID:         "mem-upsert",
		UserID:     uid,
		Kind:       "preference",
		Content:    "旧内容",
		Importance: 3,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	require.NoError(t, svc.Upsert(ctx, m1))

	// 同样 id 再次 upsert
	m2 := &model.ChatMemory{
		ID:         "mem-upsert",
		UserID:     uid,
		Kind:       "preference",
		Content:    "新内容",
		Importance: 9,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	require.NoError(t, svc.Upsert(ctx, m2))

	got, err := svc.Get(ctx, uid, "mem-upsert")
	require.NoError(t, err)
	assert.Equal(t, "新内容", got.Content)
	assert.Equal(t, 9, got.Importance)
}

func TestTouchAccessed(t *testing.T) {
	db := setupDB(t)
	svc := NewMemoryService(db)
	ctx := context.Background()
	uid := "user1"
	now := time.Now()

	m := &model.ChatMemory{
		ID:         "mem-touch",
		UserID:     uid,
		Kind:       "profile",
		Content:    "content",
		Importance: 5,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	require.NoError(t, svc.Upsert(ctx, m))

	err := svc.TouchAccessed(ctx, uid, []string{"mem-touch"})
	require.NoError(t, err)

	got, err := svc.Get(ctx, uid, "mem-touch")
	require.NoError(t, err)
	assert.NotNil(t, got.LastAccessedAt)
}

func TestCount(t *testing.T) {
	db := setupDB(t)
	svc := NewMemoryService(db)
	ctx := context.Background()
	uid := "user1"
	now := time.Now()

	for i := 0; i < 3; i++ {
		m := &model.ChatMemory{
			ID:         "mem-count" + string(rune('0'+i)),
			UserID:     uid,
			Kind:       "profile",
			Content:    "content",
			Importance: 5,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		require.NoError(t, svc.Upsert(ctx, m))
	}

	total, err := svc.Count(ctx, uid, "", "")
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)

	profileCount, err := svc.Count(ctx, uid, "profile", "")
	require.NoError(t, err)
	assert.Equal(t, int64(3), profileCount)

	otherCount, err := svc.Count(ctx, uid, "decision", "")
	require.NoError(t, err)
	assert.Equal(t, int64(0), otherCount)
}
