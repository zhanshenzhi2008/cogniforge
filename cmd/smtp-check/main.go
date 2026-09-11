package main

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"
)

// 快速诊断工具 - 直接用 .env 的配置测试 SMTP
func main() {
	// 从环境变量读取（需要先 source .env）
	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASSWORD")
	to := os.Getenv("SMTP_TEST_TO")

	if host == "" || user == "" || password == "" {
		fmt.Println("❌ 缺少环境变量，请先运行: source .env")
		fmt.Println("或手动设置: SMTP_HOST=smtp.qq.com SMTP_USER=xxx SMTP_PASSWORD=xxx")
		os.Exit(1)
	}

	port, _ := strconv.Atoi(portStr)
	if port == 0 {
		port = 465
	}

	if to == "" {
		to = user
	}

	fmt.Printf("📧 诊断 SMTP 配置...\n")
	fmt.Printf("   Host: %s\n", host)
	fmt.Printf("   Port: %d\n", port)
	fmt.Printf("   User: %s\n", user)
	fmt.Printf("   To:   %s\n\n", to)

	// 1. 连接测试
	fmt.Print("1. 连接服务器... ")
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		fmt.Println("   可能原因: 网络不通 / 端口被封 / 防火墙")
		os.Exit(1)
	}
	conn.Close()
	fmt.Println("✅")

	// 2. TLS 连接测试
	fmt.Print("2. TLS 连接... ")
	tlsCfg := &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
	tlsConn, err := tls.Dial("tcp", addr, tlsCfg)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		fmt.Println("   可能原因: TLS 版本不兼容")
		os.Exit(1)
	}
	defer tlsConn.Close()
	fmt.Println("✅")

	// 3. SMTP 认证测试
	fmt.Print("3. SMTP 认证... ")
	client, err := smtp.NewClient(tlsConn, host)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	auth := smtp.PlainAuth("", user, password, host)
	if err := client.Auth(auth); err != nil {
		fmt.Printf("❌ %v\n", err)
		fmt.Println("   🚨 授权码可能已过期！请去 QQ 邮箱重新生成")
		os.Exit(1)
	}
	fmt.Println("✅")

	// 4. 发送测试邮件
	fmt.Print("4. 发送测试邮件... ")
	if err := client.Mail(user); err != nil {
		fmt.Printf("❌ 发件人设置失败: %v\n", err)
		os.Exit(1)
	}
	if err := client.Rcpt(to); err != nil {
		fmt.Printf("❌ 收件人设置失败: %v\n", err)
		os.Exit(1)
	}

	w, err := client.Data()
	if err != nil {
		fmt.Printf("❌ 获取数据 writer 失败: %v\n", err)
		os.Exit(1)
	}

	subject := "CogniForge SMTP 诊断测试"
	html := fmt.Sprintf(`<p>这是一封诊断测试邮件。</p><p>时间: %s</p>`, time.Now().Format(time.RFC1123))
	body := buildMIME(user, to, subject, html, "")

	if _, err := w.Write(body); err != nil {
		w.Close()
		fmt.Printf("❌ 写入邮件内容失败: %v\n", err)
		os.Exit(1)
	}
	w.Close()
	client.Quit()

	fmt.Printf("✅\n\n")
	fmt.Printf("📬 邮件已发送到: %s\n", to)
	fmt.Println("   请检查收件箱和垃圾箱！")
}

func buildMIME(from, to, subject, html, text string) []byte {
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
