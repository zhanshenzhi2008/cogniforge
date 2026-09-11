package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"cogniforge/internal/mail"
	"cogniforge/internal/model"
	"cogniforge/internal/user"
)

const (
	pwdResetTTL       = 30 * time.Minute
	pwdResetRateTTL   = 15 * time.Minute
	pwdResetRateMax   = 3
	pwdResetTokPrefix = "cogniforge:pwdreset:tok:"
	pwdResetUIDPrefix = "cogniforge:pwdreset:uid:"
	pwdResetRLPrefix  = "cogniforge:pwdreset:rl:"
)

// PasswordResetOptions 前端用来决定展示邮箱表单还是仅管理员指引。
type PasswordResetOptions struct {
	EmailEnabled bool `json:"email_enabled"`
	AdminReset   bool `json:"admin_reset"`
}

func (s *AuthService) PasswordResetOptions() PasswordResetOptions {
	return PasswordResetOptions{
		EmailEnabled: s.mailer != nil && s.mailer.Enabled() && s.rdb != nil,
		AdminReset:   true,
	}
}

// RequestPasswordReset 请求重置：无论邮箱是否存在都返回同一提示（防枚举）。
func (s *AuthService) RequestPasswordReset(ctx context.Context, email string, publicURL string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || !isValidEmail(email) {
		return fmt.Errorf("请输入有效的邮箱地址")
	}

	opts := s.PasswordResetOptions()
	if !opts.EmailEnabled {
		return errMailDisabled
	}

	if err := s.checkResetRate(ctx, email); err != nil {
		return err
	}

	var u model.User
	err := s.db.Where("email = ? AND status = ?", email, "active").First(&u).Error
	if err == gorm.ErrRecordNotFound {
		slog.Info("password reset requested for unknown email")
		return nil
	}
	if err != nil {
		return fmt.Errorf("查询失败")
	}

	token, err := randomToken(32)
	if err != nil {
		return fmt.Errorf("生成令牌失败")
	}

	if err := s.storeResetToken(ctx, u.ID, token); err != nil {
		slog.Error("store reset token failed", "error", err)
		return fmt.Errorf("服务暂时不可用，请稍后再试")
	}

	// 优先使用传入的 publicURL，兜底使用配置的默认值
	baseURL := publicURL
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}
	link := strings.TrimRight(baseURL, "/") + "/reset-password?token=" + token
	html := fmt.Sprintf(`<p>你好，%s：</p>
<p>请在 30 分钟内点击下面的链接重置 CogniForge 密码：</p>
<p><a href="%s">%s</a></p>
<p>如果不是你本人操作，请忽略本邮件。</p>`, escapeHTML(u.Name), link, link)
	text := fmt.Sprintf("重置链接（30 分钟内有效）：\n%s\n", link)

	if err := s.mailer.Send(ctx, mail.Message{
		To:      u.Email,
		Subject: "重置你的 CogniForge 密码",
		HTML:    html,
		Text:    text,
	}); err != nil {
		slog.Error("send reset email failed", "error", err, "user_id", u.ID)
		return fmt.Errorf("邮件发送失败，请稍后再试或联系管理员")
	}
	return nil
}

// ResetPasswordWithToken 用邮件令牌设置新密码。
func (s *AuthService) ResetPasswordWithToken(ctx context.Context, token, newPassword string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("重置链接无效或已过期")
	}
	ok, msg := user.CheckPasswordStrength(newPassword)
	if !ok {
		return fmt.Errorf("%s", msg)
	}
	if s.rdb == nil {
		return fmt.Errorf("服务暂时不可用，请稍后再试")
	}

	userID, err := s.rdb.Get(ctx, pwdResetTokPrefix+token).Result()
	if err == redis.Nil {
		return fmt.Errorf("重置链接无效或已过期")
	}
	if err != nil {
		return fmt.Errorf("服务暂时不可用，请稍后再试")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败")
	}
	if err := s.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"password":   string(hashed),
		"updated_at": time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("密码更新失败")
	}

	_ = s.rdb.Del(ctx, pwdResetTokPrefix+token).Err()
	_ = s.rdb.Del(ctx, pwdResetUIDPrefix+userID).Err()
	return nil
}

var errMailDisabled = fmt.Errorf("邮件重置未配置")

func (s *AuthService) checkResetRate(ctx context.Context, email string) error {
	if s.rdb == nil {
		return fmt.Errorf("服务暂时不可用，请稍后再试")
	}
	key := pwdResetRLPrefix + email
	n, err := s.rdb.Incr(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("服务暂时不可用，请稍后再试")
	}
	if n == 1 {
		_ = s.rdb.Expire(ctx, key, pwdResetRateTTL).Err()
	}
	if n > pwdResetRateMax {
		return fmt.Errorf("请求过于频繁，请稍后再试")
	}
	return nil
}

func (s *AuthService) storeResetToken(ctx context.Context, userID, token string) error {
	// 作废该用户旧令牌
	if old, err := s.rdb.Get(ctx, pwdResetUIDPrefix+userID).Result(); err == nil && old != "" {
		_ = s.rdb.Del(ctx, pwdResetTokPrefix+old).Err()
	}
	pipe := s.rdb.Pipeline()
	pipe.Set(ctx, pwdResetTokPrefix+token, userID, pwdResetTTL)
	pipe.Set(ctx, pwdResetUIDPrefix+userID, token, pwdResetTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func randomToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func escapeHTML(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
	)
	return r.Replace(s)
}
