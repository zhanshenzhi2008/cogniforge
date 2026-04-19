package engine

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type CodeNodeConfig struct {
	Language string `json:"language"` // expression, javascript
	Code     string `json:"code"`
}

type CodeNodeExecutor struct{}

func (e *CodeNodeExecutor) Execute(ctx *ExecutionContext, config json.RawMessage) (any, error) {
	var cfg CodeNodeConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("invalid code config: %w", err)
	}

	ctx.Logger.Log("code", "info", fmt.Sprintf("Executing code (language: %s)", cfg.Language))

	switch cfg.Language {
	case "expression":
		return e.executeExpression(ctx, cfg.Code)
	case "javascript":
		return e.executeJavaScript(ctx, cfg.Code)
	default:
		return nil, fmt.Errorf("unsupported language: %s", cfg.Language)
	}
}

func (e *CodeNodeExecutor) executeExpression(ctx *ExecutionContext, code string) (any, error) {
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

func (e *CodeNodeExecutor) executeJavaScript(ctx *ExecutionContext, code string) (any, error) {
	code = strings.TrimSpace(code)

	if strings.HasPrefix(code, "return ") {
		code = code[7:]
	}

	for _, fn := range []struct {
		name string
		fn   func(args ...any) any
	}{
		{"Math.abs", func(args ...any) any { return math.Abs(toFloat64(args[0])) }},
		{"Math.round", func(args ...any) any { return math.Round(toFloat64(args[0])) }},
		{"Math.floor", func(args ...any) any { return math.Floor(toFloat64(args[0])) }},
		{"Math.ceil", func(args ...any) any { return math.Ceil(toFloat64(args[0])) }},
		{"Math.max", func(args ...any) any { return math.Max(toFloat64(args[0]), toFloat64(args[1])) }},
		{"Math.min", func(args ...any) any { return math.Min(toFloat64(args[0]), toFloat64(args[1])) }},
		{"Math.sqrt", func(args ...any) any { return math.Sqrt(toFloat64(args[0])) }},
		{"Math.pow", func(args ...any) any { return math.Pow(toFloat64(args[0]), toFloat64(args[1])) }},
		{"String.length", func(args ...any) any { return float64(len(toString(args[0]))) }},
		{"String.uppercase", func(args ...any) any { return strings.ToUpper(toString(args[0])) }},
		{"String.lowercase", func(args ...any) any { return strings.ToLower(toString(args[0])) }},
		{"String.trim", func(args ...any) any { return strings.TrimSpace(toString(args[0])) }},
		{"String.includes", func(args ...any) any { return strings.Contains(toString(args[0]), toString(args[1])) }},
		{"Array.length", func(args ...any) any { return float64(len(toSlice(args[0]))) }},
		{"Date.now", func(args ...any) any { return float64(time.Now().UnixMilli()) }},
	} {
		for _, placeholder := range []string{
			fmt.Sprintf("%s()", fn.name),
			fmt.Sprintf("%s(%s)", fn.name, "args"),
		} {
			if strings.Contains(code, placeholder) {
				ctx.Logger.Log("code", "warn", fmt.Sprintf("Function %s is simulated", fn.name))
			}
		}
	}

	result := code
	for key, val := range ctx.Variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", val))
	}

	for key, val := range ctx.Input {
		placeholder := fmt.Sprintf("{{input.%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", val))
	}

	return result, nil
}

func (e *CodeNodeExecutor) resolveValue(name string, ctx *ExecutionContext) any {
	if val, ok := ctx.GetVariable(name); ok {
		return val
	}
	if val, ok := ctx.Input[name]; ok {
		return val
	}
	if val, ok := ctx.Output[name]; ok {
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

func toFloat64(v any) float64 {
	if f, ok := toFloat(v); ok {
		return f
	}
	return 0
}

func toString(v any) string {
	return fmt.Sprintf("%v", v)
}

func toSlice(v any) []any {
	if s, ok := v.([]any); ok {
		return s
	}
	return nil
}
