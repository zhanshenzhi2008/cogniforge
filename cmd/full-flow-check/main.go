package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"cogniforge/internal/auth"
	"cogniforge/internal/config"
	"cogniforge/internal/mail"
	"cogniforge/internal/model"
)

func main() {
	fmt.Println("🔍 完整流程诊断测试")
	fmt.Println("====================")

	// 1. 加载配置
	cfg := config.Load()
	fmt.Println("✅ 1. 配置加载成功")
	fmt.Printf("   SMTP_HOST: %s\n", cfg.Mail.SMTPHost)
	fmt.Printf("   SMTP_USER: %s\n", cfg.Mail.SMTPUser)
	fmt.Printf("   SMTP_FROM: %s\n", cfg.Mail.From)

	// 2. 连接数据库
	dsn := cfg.DSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("❌ 2. 数据库连接失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ 2. 数据库连接成功")

	// 3. 连接 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		fmt.Printf("❌ 3. Redis 连接失败: %v\n", err)
		os.Exit(1)
	}
	defer rdb.Close()
	fmt.Println("✅ 3. Redis 连接成功")

	// 4. 构建 mailer
	mailer := buildMailer(cfg)
	fmt.Printf("✅ 4. Mailer 构建成功: %T\n", mailer)
	fmt.Printf("   Mailer.Enabled() = %v\n", mailer.Enabled())

	// 5. 构建 AuthService
	authSvc := auth.NewAuthServiceWithDeps(db, rdb, mailer)

	// 6. 检查 PasswordResetOptions
	opts := authSvc.PasswordResetOptions()
	fmt.Printf("✅ 5. PasswordResetOptions: email_enabled=%v, admin_reset=%v\n", opts.EmailEnabled, opts.AdminReset)

	// 7. 测试查询用户
	email := os.Getenv("TEST_EMAIL")
	if email == "" {
		email = "zhanshenzhi2008@foxmail.com"
	}
	fmt.Printf("\n📧 6. 测试查询用户: %s\n", email)

	var user model.User
	result := db.Where("email = ? AND status = ?", email, "active").First(&user)
	if result.Error != nil {
		fmt.Printf("   ❌ 查询失败: %v\n", result.Error)
		if result.Error == gorm.ErrRecordNotFound {
			fmt.Println("   ℹ️  用户不存在或状态不是 active")
		}
	} else {
		fmt.Printf("   ✅ 找到用户: ID=%s, Name=%s, Email=%s\n", user.ID, user.Name, user.Email)
	}

	// 8. 测试发送邮件（完整流程）
	fmt.Println("\n📤 7. 测试完整请求流程...")
	err = authSvc.RequestPasswordReset(ctx, email, "http://localhost:3000")
	if err != nil {
		fmt.Printf("   ❌ 请求失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 请求成功（无论用户是否存在，都返回成功以防枚举）")
	}

	// 9. 检查 Redis 中的 token
	time.Sleep(100 * time.Millisecond)
	keys, _ := rdb.Keys(ctx, "cogniforge:pwdreset:*").Result()
	fmt.Printf("\n🔑 8. Redis 中的密码重置 key: %d 个\n", len(keys))
	for _, k := range keys {
		fmt.Printf("   - %s\n", k)
	}

	// 10. 检查频率限制
	rateKey := "cogniforge:pwdreset:rl:" + email
	count, _ := rdb.Get(ctx, rateKey).Int()
	ttl, _ := rdb.TTL(ctx, rateKey).Result()
	fmt.Printf("\n⏱️  9. 频率限制状态:")
	if count > 0 {
		fmt.Printf(" count=%d, ttl=%v\n", count, ttl)
	} else {
		fmt.Println(" 未设置（可正常请求）")
	}

	fmt.Println("\n====================")
	fmt.Println("诊断完成")
}

// 从 router.go 复制的 buildMailer
func buildMailer(cfg *config.Config) mail.Sender {
	if cfg == nil {
		return mail.Nop{}
	}
	provider := cfg.Mail.Provider
	switch provider {
	case "", "smtp", "qq", "163":
		s := mail.NewSMTP(
			cfg.Mail.SMTPHost,
			cfg.Mail.SMTPPort,
			cfg.Mail.SMTPUser,
			cfg.Mail.SMTPPassword,
			cfg.Mail.From,
		)
		if s.Enabled() {
			return s
		}
	case "resend":
		r := mail.NewResend(cfg.Mail.APIKey, cfg.Mail.From)
		if r.Enabled() {
			return r
		}
	}
	// 兜底
	s := mail.NewSMTP(cfg.Mail.SMTPHost, cfg.Mail.SMTPPort, cfg.Mail.SMTPUser, cfg.Mail.SMTPPassword, cfg.Mail.From)
	if s.Enabled() {
		return s
	}
	r := mail.NewResend(cfg.Mail.APIKey, cfg.Mail.From)
	if r.Enabled() {
		return r
	}
	return mail.Nop{}
}
