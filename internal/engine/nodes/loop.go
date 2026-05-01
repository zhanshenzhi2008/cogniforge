package nodes

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"cogniforge/internal/engine/core"
)

type LoopNodeConfig struct {
	Type      string `json:"type"`
	Count     int    `json:"count"`
	ArrayVar  string `json:"array_var"`
	ItemVar   string `json:"item_var"`
	KeyVar    string `json:"key_var"`
	MaxIter   int    `json:"max_iter"`
	Condition string `json:"condition"`
}

type LoopNodeExecutor struct{}

func (e *LoopNodeExecutor) Execute(ctx *core.ExecutionContext, config json.RawMessage) (any, error) {
	var cfg LoopNodeConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("invalid loop config: %w", err)
	}

	if cfg.MaxIter == 0 {
		cfg.MaxIter = 100
	}

	var iterations []map[string]any

	switch cfg.Type {
	case "count", "times":
		iterations = e.executeTimesLoop(ctx, cfg.Count, cfg.MaxIter)
	case "while":
		iterations = e.executeWhileLoop(ctx, cfg.Condition, cfg.MaxIter)
	case "for_each":
		iterations = e.executeForEachLoop(ctx, cfg.ArrayVar, cfg.ItemVar, cfg.KeyVar, cfg.MaxIter)
	default:
		return nil, fmt.Errorf("unknown loop type: %s", cfg.Type)
	}

	return map[string]any{
		"type":       cfg.Type,
		"iterations": iterations,
		"total":      len(iterations),
	}, nil
}

func (e *LoopNodeExecutor) executeTimesLoop(ctx *core.ExecutionContext, count, maxIter int) []map[string]any {
	iterations := make([]map[string]any, 0)
	actualCount := count
	if actualCount > maxIter {
		actualCount = maxIter
		ctx.Logger.Logf(ctx.NodeID, "warn", "Loop count %d exceeds max_iter %d, capped", count, maxIter)
	}

	for i := 0; i < actualCount; i++ {
		ctx.SetVariable("loop.index", i)
		ctx.SetVariable("loop.count", count)
		iterations = append(iterations, map[string]any{
			"index": i,
			"count": count,
		})
		ctx.Logger.Logf(ctx.NodeID, "info", "Loop iteration %d/%d", i+1, count)
	}

	return iterations
}

func (e *LoopNodeExecutor) executeWhileLoop(ctx *core.ExecutionContext, condition string, maxIter int) []map[string]any {
	iterations := make([]map[string]any, 0)
	i := 0

	for i < maxIter {
		ctx.SetVariable("loop.index", i)

		if !e.evaluateCondition(ctx, condition) {
			break
		}

		iterations = append(iterations, map[string]any{
			"index":     i,
			"condition": condition,
			"continues": true,
		})
		ctx.Logger.Logf(ctx.NodeID, "info", "While loop iteration %d, condition satisfied", i+1)
		i++
	}

	if i >= maxIter {
		ctx.Logger.Logf(ctx.NodeID, "warn", "While loop reached max_iter %d", maxIter)
	}

	return iterations
}

func (e *LoopNodeExecutor) executeForEachLoop(ctx *core.ExecutionContext, arrayVar, itemVar, keyVar string, maxIter int) []map[string]any {
	iterations := make([]map[string]any, 0)

	arrayVal, ok := ctx.GetVariable(arrayVar)
	if !ok {
		arr, ok := ctx.Input[arrayVar]
		if !ok {
			ctx.Logger.Logf(ctx.NodeID, "warn", "Array variable %s not found", arrayVar)
			return iterations
		}
		arrayVal = arr
	}

	arr, ok := arrayVal.([]any)
	if !ok {
		ctx.Logger.Logf(ctx.NodeID, "warn", "Variable %s is not an array", arrayVar)
		return iterations
	}

	count := len(arr)
	if count > maxIter {
		count = maxIter
		ctx.Logger.Logf(ctx.NodeID, "warn", "Array length %d exceeds max_iter %d, capped", len(arr), maxIter)
	}

	for i := 0; i < count; i++ {
		ctx.SetVariable("loop.index", i)
		ctx.SetVariable("loop.count", count)
		if itemVar != "" {
			ctx.SetVariable(itemVar, arr[i])
		}
		if keyVar != "" {
			ctx.SetVariable(keyVar, i)
		}

		iterations = append(iterations, map[string]any{
			"index": i,
			"count": count,
			"item":  arr[i],
		})
		ctx.Logger.Logf(ctx.NodeID, "info", "ForEach iteration %d/%d", i+1, count)
	}

	return iterations
}

func (e *LoopNodeExecutor) evaluateCondition(ctx *core.ExecutionContext, condition string) bool {
	condExpr := strings.ReplaceAll(condition, "&&", " AND ")
	condExpr = strings.ReplaceAll(condExpr, "||", " OR ")

	parts := splitByOperators(condExpr, []string{" AND ", " OR "})
	if len(parts) == 0 {
		return false
	}

	if strings.Contains(condExpr, " OR ") {
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if e.evaluateSingleCondition(ctx, part) {
				return true
			}
		}
		return false
	}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if !e.evaluateSingleCondition(ctx, part) {
			return false
		}
	}
	return true
}

func splitByOperators(s string, operators []string) []string {
	result := []string{s}
	for _, op := range operators {
		var newResult []string
		for _, part := range result {
			splitParts := strings.Split(part, op)
			newResult = append(newResult, splitParts...)
		}
		result = newResult
	}
	return result
}

func (e *LoopNodeExecutor) evaluateSingleCondition(ctx *core.ExecutionContext, condition string) bool {
	condition = strings.TrimSpace(condition)

	for _, op := range []string{"!=", "==", ">=", "<=", ">", "<"} {
		parts := strings.SplitN(condition, op, 2)
		if len(parts) == 2 {
			field := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			value = strings.Trim(value, "\"'")

			fieldValue := e.getFieldValue(field, ctx)

			switch op {
			case "==":
				return fmt.Sprintf("%v", fieldValue) == value
			case "!=":
				return fmt.Sprintf("%v", fieldValue) != value
			case ">", "<", ">=", "<=":
				return e.compareNumeric(fieldValue, op, value, ctx)
			}
		}
	}

	return false
}

func (e *LoopNodeExecutor) getFieldValue(field string, ctx *core.ExecutionContext) any {
	if val, ok := ctx.GetVariable(field); ok {
		return val
	}
	if val, ok := ctx.Input[field]; ok {
		return val
	}
	if val, ok := ctx.Output[field]; ok {
		return val
	}
	return nil
}

func (e *LoopNodeExecutor) compareNumeric(fieldValue any, operator string, expectedValue string, ctx *core.ExecutionContext) bool {
	fieldStr := fmt.Sprintf("%v", fieldValue)
	fieldNum, err1 := strconv.ParseFloat(fieldStr, 64)
	expectedNum, err2 := strconv.ParseFloat(expectedValue, 64)

	if err1 != nil || err2 != nil {
		return false
	}

	switch operator {
	case ">":
		return fieldNum > expectedNum
	case "<":
		return fieldNum < expectedNum
	case ">=":
		return fieldNum >= expectedNum
	case "<=":
		return fieldNum <= expectedNum
	default:
		return false
	}
}
