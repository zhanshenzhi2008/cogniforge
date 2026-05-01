package engine

import (
	"cogniforge/internal/engine/core"
	"cogniforge/internal/engine/nodes"
)

type (
	ConditionNodeExecutor = nodes.ConditionNodeExecutor
	LoopNodeExecutor = nodes.LoopNodeExecutor
	StartNodeExecutor = nodes.StartNodeExecutor
	EndNodeExecutor = nodes.EndNodeExecutor
	LLMNodeExecutor = nodes.LLMNodeExecutor
	AgentNodeExecutor = nodes.AgentNodeExecutor
	HTTPNodeExecutor = nodes.HTTPNodeExecutor
	CodeNodeExecutor = nodes.CodeNodeExecutor
	DelayNodeExecutor = nodes.DelayNodeExecutor

	WorkflowDefinition = core.WorkflowDefinition
	NodeDefinition = core.NodeDefinition
	EdgeDefinition = core.EdgeDefinition
	ExecutionContext = core.ExecutionContext
	ExecutionLogger = core.ExecutionLogger
	ExecutionResult = core.ExecutionResult
	LogEntry = core.LogEntry
	Position = core.Position
)

var NewExecutionContext = core.NewExecutionContext
var NewExecutionLogger = core.NewExecutionLogger
