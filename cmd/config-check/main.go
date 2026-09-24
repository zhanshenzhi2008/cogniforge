package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

func main() {
	// 模拟 main.go 的配置加载流程
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("configs")
	viper.AddConfigPath(".")

	// 模拟 expandEnvVarsInFile
	data, _ := os.ReadFile("configs/config.yaml")
	replacer := strings.NewReplacer(
		"${PORT:-8080}", getEnv("PORT", "8080"),
		"${GIN_MODE:-debug}", getEnv("GIN_MODE", "debug"),
		"${PGSQL_HOST:-localhost}", getEnv("PGSQL_HOST", "localhost"),
		"${PGSQL_PORT:-5433}", getEnv("PGSQL_PORT", "5433"),
		"${PGSQL_USERNAME:-postgres}", getEnv("PGSQL_USERNAME", "postgres"),
		"${PGSQL_PASSWORD:-postgres-root}", getEnv("PGSQL_PASSWORD", "postgres-root"),
		"${PGSQL_DB:-cogniforge}", getEnv("PGSQL_DB", "cogniforge"),
		"${REDIS_HOST:-localhost}", getEnv("REDIS_HOST", "localhost"),
		"${REDIS_PORT:-6379}", getEnv("REDIS_PORT", "6379"),
		"${REDIS_PASSWORD:-}", getEnv("REDIS_PASSWORD", ""),
		"${ENCRYPTION_KEY:-}", getEnv("ENCRYPTION_KEY", ""),
		"${RAG_PYTHON_URL:-http://localhost:8086}", getEnv("RAG_PYTHON_URL", "http://localhost:8086"),
		"${JWT_SECRET:-cogniforge-secret-key}", getEnv("JWT_SECRET", "cogniforge-secret-key"),
		"${MAIL_PROVIDER:-smtp}", getEnv("MAIL_PROVIDER", "smtp"),
		"${RESEND_API_KEY:-}", getEnv("RESEND_API_KEY", ""),
		"${MAIL_FROM:-}", getEnv("MAIL_FROM", ""),
		"${SMTP_HOST:-smtp.qq.com}", getEnv("SMTP_HOST", "smtp.qq.com"),
		"${SMTP_PORT:-465}", getEnv("SMTP_PORT", "465"),
		"${SMTP_USER:-}", getEnv("SMTP_USER", ""),
		"${SMTP_PASSWORD:-}", getEnv("SMTP_PASSWORD", ""),
		"${LOG_LEVEL:-info}", getEnv("LOG_LEVEL", "info"),
		"${LOG_FORMAT:-json}", getEnv("LOG_FORMAT", "json"),
	)
	expanded := replacer.Replace(string(data))
	viper.ReadConfig(strings.NewReader(expanded))

	fmt.Println("📋 当前邮件配置：")
	fmt.Printf("  Provider: %q\n", viper.Get("mail.provider"))
	fmt.Printf("  SMTPHost: %q\n", viper.Get("mail.smtp_host"))
	fmt.Printf("  SMTPPort: %d\n", viper.Get("mail.smtp_port"))
	fmt.Printf("  SMTPUser: %q\n", viper.Get("mail.smtp_user"))
	fmt.Printf("  SMTPPassword: %q\n", maskPassword(viper.GetString("mail.smtp_password")))
	fmt.Printf("  From: %q\n", viper.Get("mail.from"))

	// 检查是否所有字段都正确加载
	host := viper.GetString("mail.smtp_host")
	user := viper.GetString("mail.smtp_user")
	password := viper.GetString("mail.smtp_password")
	from := viper.GetString("mail.from")

	fmt.Println()
	if host == "" {
		fmt.Println("❌ SMTP_HOST 未正确加载")
	}
	if user == "" {
		fmt.Println("❌ SMTP_USER 未正确加载")
	}
	if password == "" {
		fmt.Println("❌ SMTP_PASSWORD 未正确加载")
	}
	if from == "" {
		fmt.Println("⚠️ MAIL_FROM 为空（会使用 SMTP_USER 作为发件人）")
	}
	if host != "" && user != "" && password != "" {
		fmt.Println("✅ 邮件配置完整")
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func maskPassword(s string) string {
	if len(s) > 4 {
		return strings.Repeat("*", len(s)-4) + s[len(s)-4:]
	}
	return s
}
