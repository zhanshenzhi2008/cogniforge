package nodes

import (
	"encoding/json"
	"fmt"

	"cogniforge/internal/orchestration/engine/core"
)

// SubflowNodeConfig 嵌套调用另一个工作流（同步等待完成）
type SubflowNodeConfig struct {
	WorkflowID string         `json:"workflow_id"`
	Input      map[string]any `json:"input,omitempty"`
}

// NestedRunner 由引擎注入，避免 nodes ↔ core 循环依赖
type NestedRunner func(parent *core.ExecutionContext, workflowID string, input map[string]any) (map[string]any, error)

// SubflowNodeExecutor 子流程节点
type SubflowNodeExecutor struct {
	Run NestedRunner
}

func (e *SubflowNodeExecutor) Execute(ctx *core.ExecutionContext, config json.RawMessage) (any, error) {
	var cfg SubflowNodeConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("invalid subflow config: %w", err)
	}
	if e == nil || e.Run == nil {
		return nil, fmt.Errorf("Subflow node: runner not configured")
	}
	if cfg.WorkflowID == "" {
		return nil, fmt.Errorf("Subflow node: workflow_id is required")
	}

	input := map[string]any{}
	if len(cfg.Input) > 0 {
		for k, v := range cfg.Input {
			if s, ok := v.(string); ok {
				input[k] = resolveTemplate(s, ctx)
			} else {
				input[k] = v
			}
		}
	} else {
		for k, v := range ctx.Variables {
			input[k] = v
		}
		for k, v := range ctx.Input {
			if _, exists := input[k]; !exists {
				input[k] = v
			}
		}
	}

	out, err := e.Run(ctx, cfg.WorkflowID, input)
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = map[string]any{}
	}
	// 供下游 {{last_subflow_output}}；若有 content/字符串则兼写 last_agent_content
	if raw, mErr := json.Marshal(out); mErr == nil {
		ctx.SetVariable("last_subflow_output", string(raw))
	}
	if c, ok := out["content"]; ok {
		if s := fmt.Sprint(c); s != "" && s != "<nil>" {
			ctx.SetVariable("last_agent_content", s)
		}
	}
	return map[string]any{
		"workflow_id": cfg.WorkflowID,
		"output":      out,
		"status":      "completed",
	}, nil
}
