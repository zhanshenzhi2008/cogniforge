package mail

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

// SMTP 通用 SMTP 发信（QQ / 163 / 企业邮箱等）。
// QQ：host=smtp.qq.com port=465，密码填「授权码」不是登录密码。
type SMTP struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func NewSMTP(host string, port int, username, password, from string) *SMTP {
	if port <= 0 {
		port = 465
	}
	from = strings.TrimSpace(from)
	if from == "" {
		from = strings.TrimSpace(username)
	}
	return &SMTP{
		Host:     strings.TrimSpace(host),
		Port:     port,
		Username: strings.TrimSpace(username),
		Password: strings.TrimSpace(password),
		From:     from,
	}
}

func (s *SMTP) Enabled() bool {
	return s != nil && s.Host != "" && s.Username != "" && s.Password != "" && s.From != ""
}

func (s *SMTP) Send(ctx context.Context, msg Message) error {
	if !s.Enabled() {
		return fmt.Errorf("mail not configured")
	}
	to := strings.TrimSpace(msg.To)
	if to == "" {
		return fmt.Errorf("empty recipient")
	}

	fromAddr := extractEmail(s.From)
	body := buildMIME(s.From, to, msg.Subject, msg.HTML, msg.Text)
	addr := net.JoinHostPort(s.Host, strconv.Itoa(s.Port))
	auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)

	done := make(chan error, 1)
	go func() {
		if s.Port == 465 {
			done <- s.sendSMTPS(addr, auth, fromAddr, to, body)
			return
		}
		done <- s.sendSTARTTLS(addr, auth, fromAddr, to, body)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

func (s *SMTP) sendSMTPS(addr string, auth smtp.Auth, from, to string, body []byte) error {
	tlsCfg := &tls.Config{ServerName: s.Host, MinVersion: tls.VersionTLS12}
	dialer := &net.Dialer{Timeout: 15 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, tlsCfg)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.Host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	return s.transmit(client, auth, from, to, body)
}

func (s *SMTP) sendSTARTTLS(addr string, auth smtp.Auth, from, to string, body []byte) error {
	dialer := &net.Dialer{Timeout: 15 * time.Second}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.Host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	tlsCfg := &tls.Config{ServerName: s.Host, MinVersion: tls.VersionTLS12}
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("smtp starttls: %w", err)
		}
	}
	return s.transmit(client, auth, from, to, body)
}

func (s *SMTP) transmit(client *smtp.Client, auth smtp.Auth, from, to string, body []byte) error {
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write(body); err != nil {
		_ = w.Close()
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close: %w", err)
	}
	return client.Quit()
}

func extractEmail(from string) string {
	from = strings.TrimSpace(from)
	if i := strings.LastIndex(from, "<"); i >= 0 {
		if j := strings.Index(from[i:], ">"); j > 0 {
			return strings.TrimSpace(from[i+1 : i+j])
		}
	}
	return from
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
