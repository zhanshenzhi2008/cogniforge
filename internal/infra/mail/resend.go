package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Sender 发信接口，便于测试替换。
type Sender interface {
	Enabled() bool
	Send(ctx context.Context, msg Message) error
}

type Message struct {
	To      string
	Subject string
	HTML    string
	Text    string
}

// Resend 调用 https://api.resend.com/emails
type Resend struct {
	APIKey string
	From   string
	Client *http.Client
}

func NewResend(apiKey, from string) *Resend {
	return &Resend{
		APIKey: strings.TrimSpace(apiKey),
		From:   strings.TrimSpace(from),
		Client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (r *Resend) Enabled() bool {
	return r != nil && r.APIKey != "" && r.From != ""
}

type resendPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html,omitempty"`
	Text    string   `json:"text,omitempty"`
}

func (r *Resend) Send(ctx context.Context, msg Message) error {
	if !r.Enabled() {
		return fmt.Errorf("mail not configured")
	}
	to := strings.TrimSpace(msg.To)
	if to == "" {
		return fmt.Errorf("empty recipient")
	}
	body, err := json.Marshal(resendPayload{
		From:    r.From,
		To:      []string{to},
		Subject: msg.Subject,
		HTML:    msg.HTML,
		Text:    msg.Text,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+r.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.Client.Do(req)
	if err != nil {
		return fmt.Errorf("resend request: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("resend status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	return nil
}

// Nop 未配置时的空实现。
type Nop struct{}

func (Nop) Enabled() bool { return false }

func (Nop) Send(context.Context, Message) error {
	return fmt.Errorf("mail not configured")
}
