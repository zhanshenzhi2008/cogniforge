package nodes

import (
	"encoding/json"
	"fmt"
	"strings"

	"cogniforge/internal/orchestration/engine/core"
)

// StartVariable 开始节点入参定义（流程契约）
type StartVariable struct {
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"` // string | number | boolean
	Required    bool   `json:"required,omitempty"`
	Default     any    `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
}

// StartNodeConfig 开始节点：声明本流程要收哪些变量
type StartNodeConfig struct {
	Variables []StartVariable `json:"variables,omitempty"`
}

type StartNodeExecutor struct{}

func (e *StartNodeExecutor) Execute(ctx *core.ExecutionContext, config json.RawMessage) (any, error) {
	var cfg StartNodeConfig
	if len(config) > 0 {
		_ = json.Unmarshal(config, &cfg)
	}

	applied := map[string]any{}
	if ctx.Input == nil {
		ctx.Input = map[string]any{}
	}

	for _, v := range cfg.Variables {
		name := strings.TrimSpace(v.Name)
		if name == "" {
			continue
		}
		cur, has := ctx.Input[name]
		empty := !has || isEmptyInputValue(cur)
		if empty && v.Default != nil {
			cur = coerceStartValue(v.Type, v.Default)
			ctx.Input[name] = cur
			ctx.SetVariable(name, cur)
			empty = isEmptyInputValue(cur)
			applied[name] = cur
		} else if has {
			cur = coerceStartValue(v.Type, cur)
			ctx.Input[name] = cur
			ctx.SetVariable(name, cur)
		}
		if v.Required && empty {
			return nil, fmt.Errorf("Start node: required input %q is missing", name)
		}
		if !empty {
			applied[name] = ctx.Input[name]
		}
	}

	return map[string]any{
		"started":   true,
		"variables": applied,
	}, nil
}

func isEmptyInputValue(v any) bool {
	if v == nil {
		return true
	}
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x) == ""
	case []any:
		return len(x) == 0
	case map[string]any:
		return len(x) == 0
	default:
		return false
	}
}

func coerceStartValue(typ string, v any) any {
	t := strings.ToLower(strings.TrimSpace(typ))
	if t == "" || t == "string" {
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprint(v)
	}
	if t == "number" {
		switch x := v.(type) {
		case float64, float32, int, int64, int32:
			return x
		case json.Number:
			f, err := x.Float64()
			if err == nil {
				return f
			}
		case string:
			var f float64
			if _, err := fmt.Sscanf(strings.TrimSpace(x), "%f", &f); err == nil {
				return f
			}
		}
		return v
	}
	if t == "boolean" || t == "bool" {
		switch x := v.(type) {
		case bool:
			return x
		case string:
			s := strings.ToLower(strings.TrimSpace(x))
			return s == "true" || s == "1" || s == "yes"
		case float64:
			return x != 0
		}
		return v
	}
	return v
}

type EndNodeExecutor struct{}

func (e *EndNodeExecutor) Execute(ctx *core.ExecutionContext, config json.RawMessage) (any, error) {
	// 只回传轻量结果；完整输出已在 ctx.Output / execution.output，避免 node_run jsonb 过大失败
	keys := 0
	if ctx != nil && ctx.Output != nil {
		keys = len(ctx.Output)
	}
	return map[string]any{
		"finished":    true,
		"output_keys": keys,
	}, nil
}
