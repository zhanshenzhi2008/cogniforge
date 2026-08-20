package auth

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"cogniforge/internal/mail"
	"cogniforge/internal/model"
)

type captureMailer struct {
	last mail.Message
	n    int
}

func (c *captureMailer) Enabled() bool { return true }

func (c *captureMailer) Send(_ context.Context, msg mail.Message) error {
	c.last = msg
	c.n++
	return nil
}

func setupResetTest(t *testing.T) (*AuthService, *captureMailer, *miniredis.Miniredis) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}))

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	m := &captureMailer{}
	svc := NewAuthServiceWithDeps(db, rdb, m, "http://localhost:3000")
	require.NoError(t, db.Create(&model.User{
		ID:       "u-reset-1",
		Email:    "reset@example.com",
		Name:     "Reset User",
		Password: "old",
		Role:     "user",
		Status:   "active",
	}).Error)
	return svc, m, mr
}

func TestRequestPasswordReset_SendsMail(t *testing.T) {
	svc, m, _ := setupResetTest(t)
	require.NoError(t, svc.RequestPasswordReset(context.Background(), "reset@example.com"))
	require.Equal(t, 1, m.n)
	require.Equal(t, "reset@example.com", m.last.To)
	require.Contains(t, m.last.HTML, "/reset-password?token=")
}

func TestRequestPasswordReset_UnknownEmailSilent(t *testing.T) {
	svc, m, _ := setupResetTest(t)
	require.NoError(t, svc.RequestPasswordReset(context.Background(), "nobody@example.com"))
	require.Equal(t, 0, m.n)
}

func TestResetPasswordWithToken_OK(t *testing.T) {
	svc, m, mr := setupResetTest(t)
	require.NoError(t, svc.RequestPasswordReset(context.Background(), "reset@example.com"))
	require.Equal(t, 1, m.n)

	// 从 redis 取 token
	keys := mr.Keys()
	var token string
	for _, k := range keys {
		if len(k) > len(pwdResetTokPrefix) && k[:len(pwdResetTokPrefix)] == pwdResetTokPrefix {
			token = k[len(pwdResetTokPrefix):]
			break
		}
	}
	require.NotEmpty(t, token)

	require.NoError(t, svc.ResetPasswordWithToken(context.Background(), token, "NewPass1!"))
	require.Error(t, svc.ResetPasswordWithToken(context.Background(), token, "NewPass2!"))
}

func TestPasswordResetOptions_RequiresMailAndRedis(t *testing.T) {
	svc := NewAuthServiceWithDeps(nil, nil, mail.Nop{}, "")
	opts := svc.PasswordResetOptions()
	require.False(t, opts.EmailEnabled)
	require.True(t, opts.AdminReset)
}

func TestRequestPasswordReset_RateLimit(t *testing.T) {
	svc, _, _ := setupResetTest(t)
	ctx := context.Background()
	for i := 0; i < pwdResetRateMax; i++ {
		require.NoError(t, svc.RequestPasswordReset(ctx, "reset@example.com"))
	}
	err := svc.RequestPasswordReset(ctx, "reset@example.com")
	require.Error(t, err)
	require.Contains(t, err.Error(), "频繁")
}
