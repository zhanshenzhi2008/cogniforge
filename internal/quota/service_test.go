package quota

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"cogniforge/internal/model"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.QuotaPolicy{}, &model.LLMUsageEvent{}))
	return db
}

func seedUser(t *testing.T, db *gorm.DB, id, role string) {
	t.Helper()
	require.NoError(t, db.Create(&model.User{ID: id, Email: id + "@t.local", Name: id, Password: "x", Role: role, Status: "active"}).Error)
}

func TestAllow_BlocksAfterDailyRequests(t *testing.T) {
	db := testDB(t)
	seedUser(t, db, "u1", "user")
	svc := New(db, NewMemoryStore())
	svc.EnsureDefaultPolicy()
	p, err := svc.GetDefaultPolicy()
	require.NoError(t, err)
	p.DailyRequests = 2
	p.RPM = 100
	_, err = svc.UpdateDefaultPolicy(*p)
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, svc.Allow(ctx, "u1", "playground"))
	require.NoError(t, svc.Allow(ctx, "u1", "playground"))
	err = svc.Allow(ctx, "u1", "playground")
	require.ErrorIs(t, err, ErrExceeded)
}

func TestAllow_AdminUnlimited(t *testing.T) {
	db := testDB(t)
	seedUser(t, db, "admin1", "admin")
	svc := New(db, NewMemoryStore())
	svc.EnsureDefaultPolicy()
	p, err := svc.GetDefaultPolicy()
	require.NoError(t, err)
	p.DailyRequests = 1
	_, err = svc.UpdateDefaultPolicy(*p)
	require.NoError(t, err)

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		require.NoError(t, svc.Allow(ctx, "admin1", "playground"))
	}
}

func TestAllow_RPM(t *testing.T) {
	db := testDB(t)
	seedUser(t, db, "u2", "user")
	svc := New(db, NewMemoryStore())
	svc.EnsureDefaultPolicy()
	p, err := svc.GetDefaultPolicy()
	require.NoError(t, err)
	p.RPM = 2
	p.DailyRequests = 100
	_, err = svc.UpdateDefaultPolicy(*p)
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, svc.Allow(ctx, "u2", "playground"))
	require.NoError(t, svc.Allow(ctx, "u2", "playground"))
	err = svc.Allow(ctx, "u2", "playground")
	require.ErrorIs(t, err, ErrRateLimited)
}

func TestAllow_FailClosedWithoutStore(t *testing.T) {
	db := testDB(t)
	seedUser(t, db, "u3", "user")
	svc := New(db, nil)
	err := svc.Allow(context.Background(), "u3", "playground")
	require.ErrorIs(t, err, ErrUnavailable)
}

func TestRefundRequest(t *testing.T) {
	db := testDB(t)
	seedUser(t, db, "u4", "user")
	svc := New(db, NewMemoryStore())
	svc.EnsureDefaultPolicy()
	p, err := svc.GetDefaultPolicy()
	require.NoError(t, err)
	p.DailyRequests = 1
	p.RPM = 100
	_, err = svc.UpdateDefaultPolicy(*p)
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, svc.Allow(ctx, "u4", "playground"))
	require.ErrorIs(t, svc.Allow(ctx, "u4", "playground"), ErrExceeded)
	svc.RefundRequest(ctx, "u4")
	require.NoError(t, svc.Allow(ctx, "u4", "playground"))
}

func TestMe_WarnAt80Percent(t *testing.T) {
	db := testDB(t)
	seedUser(t, db, "u5", "user")
	store := NewMemoryStore()
	svc := New(db, store)
	svc.EnsureDefaultPolicy()
	p, err := svc.GetDefaultPolicy()
	require.NoError(t, err)
	p.DailyRequests = 10
	_, err = svc.UpdateDefaultPolicy(*p)
	require.NoError(t, err)

	ctx := context.Background()
	for i := 0; i < 8; i++ {
		require.NoError(t, svc.Allow(ctx, "u5", "playground"))
	}
	snap, err := svc.Me(ctx, "u5")
	require.NoError(t, err)
	assert.True(t, snap.Warn)
	assert.Equal(t, int64(8), snap.Day.RequestsUsed)
}

func TestCommit_AddsTokens(t *testing.T) {
	db := testDB(t)
	seedUser(t, db, "u6", "user")
	svc := New(db, NewMemoryStore())
	svc.EnsureDefaultPolicy()
	ctx := context.Background()
	require.NoError(t, svc.Allow(ctx, "u6", "playground"))
	svc.Commit(ctx, CommitInput{UserID: "u6", Source: "playground", Model: "deepseek-chat", TotalTokens: 100, Status: statusOK})
	snap, err := svc.Me(ctx, "u6")
	require.NoError(t, err)
	assert.Equal(t, int64(100), snap.Day.TokensUsed)
	assert.Equal(t, int64(100), snap.Month.TokensUsed)

	var n int64
	db.Model(&model.LLMUsageEvent{}).Count(&n)
	assert.Equal(t, int64(1), n)
}
