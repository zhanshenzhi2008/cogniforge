package nodes

import (
	"encoding/json"
	"fmt"

	"cogniforge/internal/orchestration/engine/core"
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
	reason := resolveTemplate(cfg.Reason, ctx)
	if reason == "" {
		reason = "approval"
	}
	msg := resolveTemplate(cfg.Message, ctx)
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
	storeLastWait(ctx, out)
	return out, nil
}

// storeLastWait 挂起前写入，供下游/审批后续步骤 {{last_wait_reason}} 等引用
func storeLastWait(ctx *core.ExecutionContext, out map[string]any) {
	if ctx == nil || out == nil {
		return
	}
	ctx.SetVariable("last_wait_reason", out["reason"])
	ctx.SetVariable("last_wait_message", out["message"])
	if sec, ok := out["timeout_sec"]; ok {
		ctx.SetVariable("last_wait_timeout_sec", sec)
	} else {
		ctx.SetVariable("last_wait_timeout_sec", 0)
	}
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
