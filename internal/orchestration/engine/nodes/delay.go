package nodes

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cogniforge/internal/orchestration/engine/core"
)

type DelayNodeConfig struct {
	Duration json.RawMessage `json:"duration"` // 数字或 "{{变量}}" 字符串
	Unit     string          `json:"unit"`
}

type DelayNodeExecutor struct{}

func (e *DelayNodeExecutor) Execute(ctx *core.ExecutionContext, config json.RawMessage) (any, error) {
	var cfg DelayNodeConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("invalid delay config: %w", err)
	}

	durN := resolveDelayDuration(cfg.Duration, ctx)
	if durN <= 0 {
		durN = 1
	}

	var duration time.Duration
	switch strings.ToLower(cfg.Unit) {
	case "minutes", "minute", "m":
		duration = time.Duration(durN) * time.Minute
	case "hours", "hour", "h":
		duration = time.Duration(durN) * time.Hour
	default:
		duration = time.Duration(durN) * time.Second
	}

	maxDelay := 5 * time.Minute
	if duration > maxDelay {
		ctx.Logger.Logf(ctx.NodeID, "warn", "Delay capped at %v (requested: %v)", maxDelay, duration)
		duration = maxDelay
	}

	ctx.Logger.Logf(ctx.NodeID, "info", "Delaying for %v", duration)
	time.Sleep(duration)

	out := map[string]any{
		"delayed":   true,
		"duration":  durN,
		"unit":      cfg.Unit,
		"actual_ms": duration.Milliseconds(),
	}
	storeLastDelay(ctx, out)
	return out, nil
}

// resolveDelayDuration 支持 JSON 数字，或经 {{变量}} 展开后的整数字符串
func resolveDelayDuration(raw json.RawMessage, ctx *core.ExecutionContext) int {
	if len(raw) == 0 || string(raw) == "null" {
		return 1
	}
	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		return n
	}
	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		return int(f)
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return 1
	}
	s = strings.TrimSpace(resolveTemplate(s, ctx))
	if s == "" {
		return 1
	}
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	if fv, err := strconv.ParseFloat(s, 64); err == nil {
		return int(fv)
	}
	return 1
}

// storeLastDelay 供下游 {{last_delay_ms}} / {{last_delay_duration}} / {{last_delay_unit}} 引用
func storeLastDelay(ctx *core.ExecutionContext, out map[string]any) {
	if ctx == nil || out == nil {
		return
	}
	ctx.SetVariable("last_delay_ms", out["actual_ms"])
	ctx.SetVariable("last_delay_duration", out["duration"])
	ctx.SetVariable("last_delay_unit", out["unit"])
}

func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)

	numberStr := ""
	unit := ""
	for i, ch := range s {
		if ch >= '0' && ch <= '9' || ch == '.' {
			numberStr += string(ch)
		} else {
			unit = strings.TrimSpace(s[i:])
			break
		}
	}

	if numberStr == "" {
		return 0, fmt.Errorf("invalid duration: %s", s)
	}

	number, err := strconv.ParseFloat(numberStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid duration number: %s", s)
	}

	unit = strings.ToLower(unit)
	switch unit {
	case "", "s", "sec", "second", "seconds":
		return time.Duration(number * float64(time.Second)), nil
	case "m", "min", "minute", "minutes":
		return time.Duration(number * float64(time.Minute)), nil
	case "h", "hr", "hour", "hours":
		return time.Duration(number * float64(time.Hour)), nil
	default:
		return 0, fmt.Errorf("unknown duration unit: %s", unit)
	}
}
