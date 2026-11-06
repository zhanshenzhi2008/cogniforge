package nodes

import (
	"encoding/json"
	"fmt"

	"cogniforge/internal/engine/core"
)

// WaitNodeConfig 人工/外部等待节点（阶段十八 W2/W3）
type WaitNodeConfig struct {
	Reason     string `json:"reason,omitempty"`      // approval / external / manual
	Message    string `json:"message,omitempty"`     // 给人看的提示
	TimeoutSec int    `json:"timeout_sec,omitempty"` // >0 时写入 wait_until；超时扫描置 failed
}

// WaitNodeExecutor 命中后由引擎挂起执行（status=waiting），不在此阻塞
type WaitNodeExecutor struct{}

func (e *WaitNodeExecutor) Execute(ctx *core.ExecutionContext, config json.RawMessage) (any, error) {
	var cfg WaitNodeConfig
	if len(config) > 0 {
		_ = json.Unmarshal(config, &cfg)
	}
	reason := cfg.Reason
	if reason == "" {
		reason = "approval"
	}
	msg := cfg.Message
	if msg == "" {
		msg = "Waiting for human approval or external event"
	}
	out := map[string]any{
		"__wait__": true,
		"reason":   reason,
		"message":  msg,
	}
	if cfg.TimeoutSec > 0 {
		out["timeout_sec"] = cfg.TimeoutSec
	}
	return out, nil
}

// IsWaitOutput 判断节点输出是否为等待信号
func IsWaitOutput(output any) (reason, message string, timeoutSec int, ok bool) {
	m, ok := output.(map[string]any)
	if !ok {
		return "", "", 0, false
	}
	flag, _ := m["__wait__"].(bool)
	if !flag {
		return "", "", 0, false
	}
	reason, _ = m["reason"].(string)
	message, _ = m["message"].(string)
	if reason == "" {
		reason = "approval"
	}
	switch v := m["timeout_sec"].(type) {
	case float64:
		timeoutSec = int(v)
	case int:
		timeoutSec = v
	case json.Number:
		n, _ := v.Int64()
		timeoutSec = int(n)
	}
	return reason, message, timeoutSec, true
}

// ErrWaitRequested 仅作文档占位；实际用 IsWaitOutput
var ErrWaitRequested = fmt.Errorf("wait requested")
