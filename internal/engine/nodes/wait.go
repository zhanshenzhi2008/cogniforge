package nodes

import (
	"encoding/json"
	"fmt"

	"cogniforge/internal/engine/core"
)

// WaitNodeConfig 人工/外部等待节点（阶段十八 W2）
type WaitNodeConfig struct {
	Reason  string `json:"reason,omitempty"`  // approval / external / manual
	Message string `json:"message,omitempty"` // 给人看的提示
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
	// 特殊标记：引擎检测后进入 waiting
	return map[string]any{
		"__wait__": true,
		"reason":   reason,
		"message":  msg,
	}, nil
}

// IsWaitOutput 判断节点输出是否为等待信号
func IsWaitOutput(output any) (reason, message string, ok bool) {
	m, ok := output.(map[string]any)
	if !ok {
		return "", "", false
	}
	flag, _ := m["__wait__"].(bool)
	if !flag {
		return "", "", false
	}
	reason, _ = m["reason"].(string)
	message, _ = m["message"].(string)
	if reason == "" {
		reason = "approval"
	}
	return reason, message, true
}

// ErrWaitRequested 仅作文档占位；实际用 IsWaitOutput
var ErrWaitRequested = fmt.Errorf("wait requested")
