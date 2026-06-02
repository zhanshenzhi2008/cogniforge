package engine

import (
	"cogniforge/internal/engine/core"
	"cogniforge/internal/engine/debug"
	"cogniforge/internal/engine/nodes"
)

type (
	// 节点执行器
	ConditionNodeExecutor = nodes.ConditionNodeExecutor
	LoopNodeExecutor      = nodes.LoopNodeExecutor
	StartNodeExecutor     = nodes.StartNodeExecutor
	EndNodeExecutor       = nodes.EndNodeExecutor
	LLMNodeExecutor       = nodes.LLMNodeExecutor
	AgentNodeExecutor     = nodes.AgentNodeExecutor
	HTTPNodeExecutor      = nodes.HTTPNodeExecutor
	CodeNodeExecutor      = nodes.CodeNodeExecutor
	DelayNodeExecutor     = nodes.DelayNodeExecutor

	// 核心类型
	WorkflowDefinition = core.WorkflowDefinition
	NodeDefinition     = core.NodeDefinition
	EdgeDefinition     = core.EdgeDefinition
	ExecutionContext   = core.ExecutionContext
	ExecutionLogger    = core.ExecutionLogger
	ExecutionResult    = core.ExecutionResult
	LogEntry           = core.LogEntry
	Position           = core.Position

	// 调试器
	Debugger     = debug.Debugger
	DebugMode    = debug.DebugMode
	DebugState   = debug.DebugState
	Breakpoint   = debug.Breakpoint
	DebugStatus  = debug.DebugStatus
	DebugCommand = debug.DebugCommand
)

// 核心包函数
var (
	NewExecutionContext            = core.NewExecutionContext
	NewExecutionContextWithTraceID = core.NewExecutionContextWithTraceID
	NewExecutionLogger             = core.NewExecutionLogger
)

// 调试器构造函数
var NewDebugger = debug.NewDebugger

// WorkflowEngine 引擎实例
type WorkflowEngine = core.WorkflowEngine

// NewEngine 创建并初始化引擎（注册默认节点执行器）
func NewEngine() *WorkflowEngine {
	engine := core.NewEngine()

	// 注册默认节点执行器
	engine.RegisterExecutor("start", &StartNodeExecutor{})
	engine.RegisterExecutor("end", &EndNodeExecutor{})
	engine.RegisterExecutor("llm", &LLMNodeExecutor{})
	engine.RegisterExecutor("agent", &AgentNodeExecutor{})
	engine.RegisterExecutor("condition", &ConditionNodeExecutor{})
	engine.RegisterExecutor("loop", &LoopNodeExecutor{})
	engine.RegisterExecutor("http", nodes.NewHTTPNodeExecutor())
	engine.RegisterExecutor("code", &CodeNodeExecutor{})
	engine.RegisterExecutor("delay", &DelayNodeExecutor{})

	return engine
}

// 调试器常量
const (
	DebugModeStep   = debug.DebugModeStep
	DebugModeResume = debug.DebugModeResume
	DebugModePause  = debug.DebugModePause
	DebugModeStop   = debug.DebugModeStop
	DebugModeAuto   = debug.DebugModeAuto

	DebugStateRunning = debug.DebugStateRunning
	DebugStatePaused  = debug.DebugStatePaused
	DebugStateStopped = debug.DebugStateStopped
	DebugStateDone    = debug.DebugStateDone
)
