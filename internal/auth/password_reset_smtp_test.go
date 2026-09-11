package auth

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/smtp"
	"os"
	"strings"
	"testing"
	"time"
)

// TestRealSMTPSend_Diagnose 真实 SMTP 发送测试，带详细诊断信息。
// 运行方式：
//
//	SMTP_HOST=smtp.qq.com SMTP_PORT=465 SMTP_USER=xxx@qq.com SMTP_PASSWORD=xxx go test -v -run TestRealSMTPSend
func TestRealSMTPSend_Diagnose(t *testing.T) {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASSWORD")
	to := os.Getenv("SMTP_TEST_TO") // 测试收件人，可与 SMTP_USER 不同

	if host == "" {
		t.Skip("跳过：未设置 SMTP_HOST 环境变量")
	}
	if user == "" || password == "" {
		t.Skip("跳过：未设置 SMTP_USER 或 SMTP_PASSWORD 环境变量")
	}
	if to == "" {
		to = user // 默认发给自己测试
	}

	if port == "" {
		port = "465"
	}

	// ─── 诊断步骤 ───────────────────────────────────────────────
	t.Run("诊断1_连接SMTP服务器", func(t *testing.T) {
		addr := net.JoinHostPort(host, port)
		conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
		if err != nil {
			t.Fatalf("❌ 连接失败: %v\n可能原因：\n  - 网络不通（防火墙/代理）\n  - SMTP 端口被封\n  - 主机名错误: %s", err, host)
		}
		conn.Close()
		t.Logf("✅ 连接成功: %s", addr)
	})

	t.Run("诊断2_SMTP认证", func(t *testing.T) {
		addr := net.JoinHostPort(host, port)
		auth := smtp.PlainAuth("", user, password, host)

		done := make(chan error, 1)
		go func() {
			done <- testSMTPAuthTLS(addr, host, user, auth)
		}()

		select {
		case authErr := <-done:
			if authErr != nil {
				t.Fatalf("❌ SMTP 认证失败: %v\n可能原因：\n  - 授权码错误（不是 QQ 密码！）\n  - 邮箱未开启 SMTP 服务\n  - 授权码已过期（需重新生成）\n  - 部分邮箱需要开启 IMAP/SMTP 服务", authErr)
			}
			t.Log("✅ SMTP 认证成功")
		case <-time.After(15 * time.Second):
			t.Fatal("❌ SMTP 认证超时（15秒）")
		}
	})

	t.Run("诊断3_发送测试邮件", func(t *testing.T) {
		addr := net.JoinHostPort(host, port)
		auth := smtp.PlainAuth("", user, password, host)
		from := user

		subject := "CogniForge 邮件发送测试"
		html := fmt.Sprintf(`<p>这是一封测试邮件。</p><p>发送时间: %s</p>`, time.Now().Format(time.RFC1123))
		body := buildTestMIME(from, to, subject, html, "")

		done := make(chan error, 1)
		go func() {
			done <- sendMailTLS(addr, host, auth, from, to, body)
		}()

		select {
		case err := <-done:
			if err != nil {
				t.Fatalf("❌ 邮件发送失败: %v\n可能原因：\n  - 发件人邮箱未验证\n  - QQ/163 等邮箱的每日发送限制\n  - 收件人邮箱地址无效或不存在\n  - 邮件被收件方服务器拒绝（垃圾邮件/黑名单）", err)
			}
			t.Logf("✅ 邮件发送成功！请检查 %s 的收件箱（包括垃圾箱）", to)
		case <-time.After(15 * time.Second):
			t.Fatal("❌ 邮件发送超时（15秒）")
		}
	})
}

// testSMTPAuthTLS 使用 TLS 测试 SMTP 认证
func testSMTPAuthTLS(addr, host string, from string, auth smtp.Auth) error {
	tlsCfg := &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}

	conn, err := tls.Dial("tcp", addr, tlsCfg)
	if err != nil {
		return fmt.Errorf("TLS 连接失败: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("创建 SMTP 客户端失败: %w", err)
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("认证失败: %w", err)
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}

	return nil
}

// sendMailTLS 使用 TLS 发送邮件
func sendMailTLS(addr, host string, auth smtp.Auth, from, to string, body []byte) error {
	tlsCfg := &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}

	conn, err := tls.Dial("tcp", addr, tlsCfg)
	if err != nil {
		return fmt.Errorf("TLS 连接失败: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("创建 SMTP 客户端失败: %w", err)
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("认证失败: %w", err)
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("设置收件人失败: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("获取数据 writer 失败: %w", err)
	}

	if _, err := w.Write(body); err != nil {
		w.Close()
		return fmt.Errorf("写入邮件内容失败: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("关闭 writer 失败: %w", err)
	}

	return client.Quit()
}

// buildTestMIME 构建测试邮件
func buildTestMIME(from, to, subject, html, text string) []byte {
	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + to + "\r\n")
	b.WriteString("Subject: =?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(subject)) + "?=\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	if html != "" {
		b.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		b.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
		b.WriteString(html)
	} else {
		b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		b.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
		b.WriteString(text)
	}
	b.WriteString("\r\n")
	return []byte(b.String())
}

// TestMailConfigValidation 验证 .env 配置是否完整
func TestMailConfigValidation(t *testing.T) {
	required := []struct {
		name     string
		envKey   string
		fallback string
	}{
		{"SMTP_HOST", "SMTP_HOST", ""},
		{"SMTP_USER", "SMTP_USER", ""},
		{"SMTP_PASSWORD", "SMTP_PASSWORD", ""},
		{"MAIL_FROM", "MAIL_FROM", ""},
	}

	missing := []string{}
	for _, r := range required {
		val := os.Getenv(r.envKey)
		if val == "" {
			if r.fallback != "" {
				t.Logf("⚠️ %s 未设置，使用默认值: %s", r.name, r.fallback)
			} else {
				missing = append(missing, r.name)
			}
		} else {
			t.Logf("✅ %s = %s", r.name, maskPassword(r.name, val))
		}
	}

	if len(missing) > 0 {
		t.Fatalf("❌ 缺少必需配置: %v\n请检查 .env 文件", missing)
	}
}

func maskPassword(name, val string) string {
	if name == "SMTP_PASSWORD" && len(val) > 4 {
		return strings.Repeat("*", len(val)-4) + val[len(val)-4:]
	}
	return val
}

// TestSMTPSendViaGoCode 用 Go 代码直接发送（测试你的代码是否正确）
func TestSMTPSendViaGoCode(t *testing.T) {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		t.Skip("跳过：未设置 SMTP_HOST")
	}

	user := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASSWORD")
	to := os.Getenv("SMTP_TEST_TO")
	if to == "" {
		to = user
	}

	// 复用项目里的 mail.NewSMTP
	port := 465
	from := user

	t.Log("使用 Go 代码发送邮件...")
	err := sendViaSMTP(host, port, user, password, from, to)
	if err != nil {
		t.Fatalf("❌ 发送失败: %v", err)
	}
	t.Logf("✅ 发送成功！请检查 %s 的收件箱/垃圾箱", to)
}

// sendViaSMTP 封装发送逻辑
func sendViaSMTP(host string, port int, user, password, from, to string) error {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	tlsCfg := &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}

	conn, err := tls.Dial("tcp", addr, tlsCfg)
	if err != nil {
		return fmt.Errorf("连接: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("客户端: %w", err)
	}
	defer client.Close()

	auth := smtp.PlainAuth("", user, password, host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("认证: %w", err)
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("发件人: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("收件人: %w", err)
	}

	body := buildTestMIME(from, to, "测试邮件", "<p>测试</p>", "")
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("数据: %w", err)
	}
	if _, err := w.Write(body); err != nil {
		w.Close()
		return fmt.Errorf("写入: %w", err)
	}
	w.Close()
	return client.Quit()
}

// TestResponseReader 读取服务器响应用于调试
func TestSMTPDebugResponse(t *testing.T) {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	if host == "" {
		t.Skip("跳过")
	}
	if port == "" {
		port = "465"
	}

	addr := net.JoinHostPort(host, port)
	tlsCfg := &tls.Config{ServerName: host, InsecureSkipVerify: true}

	conn, err := tls.Dial("tcp", addr, tlsCfg)
	if err != nil {
		t.Fatalf("连接失败: %v", err)
	}
	defer conn.Close()

	reader := io.TeeReader(conn, os.Stderr)
	buf := make([]byte, 1024)
	n, _ := reader.Read(buf)
	t.Logf("服务器响应:\n%s", string(buf[:n]))
}
