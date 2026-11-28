package nodes

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"cogniforge/internal/capability/mcp"
	"cogniforge/internal/orchestration/engine/core"
)

type CodeNodeConfig struct {
	Language string          `json:"language"`
	Code     string          `json:"code"`
	Timeout  json.RawMessage `json:"timeout,omitempty"` // 数字或 "{{变量}}"；沙箱用，空/0→默认
}

type CodeNodeExecutor struct{}

func (e *CodeNodeExecutor) Execute(ctx *core.ExecutionContext, config json.RawMessage) (any, error) {
	var cfg CodeNodeConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("invalid code config: %w", err)
	}

	lang := strings.ToLower(strings.TrimSpace(cfg.Language))
	if lang == "" {
		lang = "expression"
	}
	ctx.Logger.Log("code", "info", fmt.Sprintf("Executing code (language: %s)", lang))

	timeoutSec := resolveFlexibleInt(cfg.Timeout, ctx)

	var out any
	var err error
	switch lang {
	case "expression":
		out, err = e.executeExpression(ctx, cfg.Code)
	case "javascript", "js", "node", "python":
		out, err = e.executeSandboxed(ctx, lang, cfg.Code, timeoutSec)
	default:
		return nil, fmt.Errorf("unsupported language: %s", cfg.Language)
	}
	if err != nil {
		return nil, err
	}
	storeLastCodeOutput(ctx, out)
	return out, nil
}

// storeLastCodeOutput 供下游 {{last_code_output}} 引用
func storeLastCodeOutput(ctx *core.ExecutionContext, out any) {
	if ctx == nil || out == nil {
		return
	}
	if m, ok := out.(map[string]any); ok {
		if s, ok := m["output"].(string); ok {
			ctx.SetVariable("last_code_output", s)
			return
		}
	}
	ctx.SetVariable("last_code_output", fmt.Sprint(out))
}

func (e *CodeNodeExecutor) executeExpression(ctx *core.ExecutionContext, code string) (any, error) {
	code = strings.TrimSpace(code)

	for _, op := range []string{"+", "-", "*", "/", "%"} {
		pattern := regexp.MustCompile(fmt.Sprintf(`(\w+)\s*\%s\s*(\w+)`, regexp.QuoteMeta(op)))
		matches := pattern.FindAllStringSubmatch(code, -1)
		for _, match := range matches {
			left := e.resolveValue(match[1], ctx)
			right := e.resolveValue(match[2], ctx)
			var result any
			if num1, ok1 := toFloat(left); ok1 {
				if num2, ok2 := toFloat(right); ok2 {
					switch op {
					case "+":
						result = num1 + num2
					case "-":
						result = num1 - num2
					case "*":
						result = num1 * num2
					case "/":
						if num2 != 0 {
							result = num1 / num2
						} else {
							return nil, fmt.Errorf("division by zero")
						}
					case "%":
						result = float64(int(num1) % int(num2))
					}
				}
			}
			if result != nil {
				code = strings.Replace(code, match[0], fmt.Sprintf("%v", result), 1)
			}
		}
	}

	return e.resolveValue(strings.TrimSpace(code), ctx), nil
}

func (e *CodeNodeExecutor) executeSandboxed(ctx *core.ExecutionContext, language, code string, timeoutSec int) (any, error) {
	wrapped, err := wrapCodeWithContext(language, code, ctx)
	if err != nil {
		return nil, err
	}
	sandboxLang := language
	if sandboxLang == "js" || sandboxLang == "node" {
		sandboxLang = "javascript"
	}
	res, err := mcp.ExecuteSandboxedCode(sandboxLang, wrapped, timeoutSec)
	if err != nil {
		return nil, err
	}
	status, _ := res["status"].(string)
	out, _ := res["output"].(string)
	if status == "disabled" {
		return nil, fmt.Errorf("code sandbox disabled (CODE_EXECUTE_DISABLED)")
	}
	if status != "ok" {
		if out == "" {
			out = status
		}
		return nil, fmt.Errorf("sandbox %s: %s", status, out)
	}
	return map[string]any{
		"language":   sandboxLang,
		"output":     out,
		"status":     status,
		"exit_code":  res["exit_code"],
		"elapsed_ms": res["elapsed_ms"],
	}, nil
}

func wrapCodeWithContext(language, code string, ctx *core.ExecutionContext) (string, error) {
	vars := map[string]any{}
	for k, v := range ctx.Variables {
		vars[k] = v
	}
	payload := map[string]any{
		"input":     ctx.Input,
		"variables": vars,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	quoted := strconv.Quote(string(raw))
	code = strings.TrimSpace(code)
	switch strings.ToLower(language) {
	case "python":
		return fmt.Sprintf(
			"import json\n_cf_context = json.loads(%s)\nvariables = _cf_context.get('variables') or {}\ninput = _cf_context.get('input') or {}\n%s\n",
			quoted, code,
		), nil
	default:
		// javascript / node
		return fmt.Sprintf(
			"const _cf_context = JSON.parse(%s);\nconst variables = _cf_context.variables || {};\nconst input = _cf_context.input || {};\n%s\n",
			quoted, code,
		), nil
	}
}

func (e *CodeNodeExecutor) resolveValue(name string, ctx *core.ExecutionContext) any {
	if val, ok := ctx.GetVariable(name); ok {
		return val
	}
	if val, ok := ctx.Input[name]; ok {
		return val
	}
	if val, ok := ctx.LookupOutput(name); ok {
		return val
	}

	if num, err := strconv.ParseFloat(name, 64); err == nil {
		return num
	}

	if name == "true" {
		return true
	}
	if name == "false" {
		return false
	}

	if (strings.HasPrefix(name, "\"") && strings.HasSuffix(name, "\"")) ||
		(strings.HasPrefix(name, "'") && strings.HasSuffix(name, "'")) {
		return name[1 : len(name)-1]
	}

	return name
}

func toFloat(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

