package engine

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type ConditionNodeConfig struct {
	Conditions []Condition `json:"conditions"`
}

type Condition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"` // ==, !=, >, <, >=, <=, contains, starts_with, ends_with, is_empty, is_not_empty, matches
	Value    any    `json:"value"`
	Branch   string `json:"branch"` // true_branch or false_branch
}

type ConditionNodeExecutor struct{}

func (e *ConditionNodeExecutor) Execute(ctx *ExecutionContext, config json.RawMessage) (any, error) {
	var cfg ConditionNodeConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("invalid condition config: %w", err)
	}

	if len(cfg.Conditions) == 0 {
		return map[string]any{
			"result": true,
			"branch": "default",
		}, nil
	}

	for _, condition := range cfg.Conditions {
		fieldValue := e.getFieldValue(condition.Field, ctx)
		result := e.evaluateCondition(fieldValue, condition.Operator, condition.Value)

		ctx.Logger.Logf(ctx.NodeID, "info",
			"Condition: %s %s %v = %v (actual: %v)",
			condition.Field, condition.Operator, condition.Value, result, fieldValue)

		if result {
			return map[string]any{
				"result":   true,
				"branch":   condition.Branch,
				"field":    condition.Field,
				"operator": condition.Operator,
				"expected": condition.Value,
				"actual":   fieldValue,
			}, nil
		}
	}

	return map[string]any{
		"result": false,
		"branch": "default",
	}, nil
}

func (e *ConditionNodeExecutor) getFieldValue(field string, ctx *ExecutionContext) any {
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

func (e *ConditionNodeExecutor) evaluateCondition(fieldValue any, operator string, expectedValue any) bool {
	switch operator {
	case "==":
		return fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", expectedValue)
	case "!=":
		return fmt.Sprintf("%v", fieldValue) != fmt.Sprintf("%v", expectedValue)
	case ">", "<", ">=", "<=":
		return e.compareNumeric(fieldValue, operator, expectedValue)
	case "contains":
		return strings.Contains(fmt.Sprintf("%v", fieldValue), fmt.Sprintf("%v", expectedValue))
	case "starts_with":
		return strings.HasPrefix(fmt.Sprintf("%v", fieldValue), fmt.Sprintf("%v", expectedValue))
	case "ends_with":
		return strings.HasSuffix(fmt.Sprintf("%v", fieldValue), fmt.Sprintf("%v", expectedValue))
	case "is_empty":
		return fieldValue == nil || fmt.Sprintf("%v", fieldValue) == ""
	case "is_not_empty":
		return fieldValue != nil && fmt.Sprintf("%v", fieldValue) != ""
	case "matches":
		if regex, err := regexp.Compile(fmt.Sprintf("%v", expectedValue)); err == nil {
			return regex.MatchString(fmt.Sprintf("%v", fieldValue))
		}
		return false
	default:
		return false
	}
}

func (e *ConditionNodeExecutor) compareNumeric(fieldValue any, operator string, expectedValue any) bool {
	fieldStr := fmt.Sprintf("%v", fieldValue)
	expectedStr := fmt.Sprintf("%v", expectedValue)

	fieldNum, err1 := strconv.ParseFloat(fieldStr, 64)
	expectedNum, err2 := strconv.ParseFloat(expectedStr, 64)

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
