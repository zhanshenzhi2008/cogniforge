package nodes

import (
	"encoding/json"

	"cogniforge/internal/engine/core"
)

type StartNodeExecutor struct{}

func (e *StartNodeExecutor) Execute(ctx *core.ExecutionContext, config json.RawMessage) (any, error) {
	return map[string]any{
		"started": true,
	}, nil
}

type EndNodeExecutor struct{}

func (e *EndNodeExecutor) Execute(ctx *core.ExecutionContext, config json.RawMessage) (any, error) {
	return map[string]any{
		"finished": true,
		"output":   ctx.Output,
	}, nil
}
