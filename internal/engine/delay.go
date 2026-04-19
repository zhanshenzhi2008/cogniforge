package engine

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type DelayNodeConfig struct {
	Duration int    `json:"duration"` // 延迟时间（秒）
	Unit     string `json:"unit"`     // seconds, minutes, hours
}

type DelayNodeExecutor struct{}

func (e *DelayNodeExecutor) Execute(ctx *ExecutionContext, config json.RawMessage) (any, error) {
	var cfg DelayNodeConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("invalid delay config: %w", err)
	}

	if cfg.Duration <= 0 {
		cfg.Duration = 1
	}

	var duration time.Duration
	switch strings.ToLower(cfg.Unit) {
	case "minutes", "minute", "m":
		duration = time.Duration(cfg.Duration) * time.Minute
	case "hours", "hour", "h":
		duration = time.Duration(cfg.Duration) * time.Hour
	default:
		duration = time.Duration(cfg.Duration) * time.Second
	}

	maxDelay := 5 * time.Minute
	if duration > maxDelay {
		ctx.Logger.Logf(ctx.NodeID, "warn", "Delay capped at %v (requested: %v)", maxDelay, duration)
		duration = maxDelay
	}

	ctx.Logger.Logf(ctx.NodeID, "info", "Delaying for %v", duration)
	time.Sleep(duration)

	return map[string]any{
		"delayed":   true,
		"duration":  cfg.Duration,
		"unit":      cfg.Unit,
		"actual_ms": duration.Milliseconds(),
	}, nil
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
