package engine

import (
	"encoding/json"
)

type StartNodeExecutor struct{}

func (e *StartNodeExecutor) Execute(ctx *ExecutionContext, config json.RawMessage) (any, error) {
	return map[string]any{
		"started": true,
	}, nil
}

type EndNodeExecutor struct{}

func (e *EndNodeExecutor) Execute(ctx *ExecutionContext, config json.RawMessage) (any, error) {
	return map[string]any{
		"finished": true,
		"output":   ctx.Output,
	}, nil
}
