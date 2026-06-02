package core

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"cogniforge/internal/trace"
)

type ExecutionContext struct {
	WorkflowID  string
	ExecutionID string
	TraceID     string          // 新增：请求追踪 ID
	Context     context.Context // 新增：用于传递 traceId
	NodeID      string
	Variables   map[string]any
	Input       map[string]any
	Output      map[string]any
	State       map[string]string
	StartTime   time.Time
	mu          sync.RWMutex
	Logger      *ExecutionLogger
	OnNodeStart func(nodeID string)
	OnNodeEnd   func(nodeID string, status string)
	OnProgress  func(nodeID string, message string)
}

// NewExecutionContext 创建执行上下文，可选传入 traceId
func NewExecutionContext(workflowID, executionID string, input map[string]any) *ExecutionContext {
	return NewExecutionContextWithTraceID(workflowID, executionID, input, "")
}

// NewExecutionContextWithTraceID 创建带 traceId 的执行上下文
func NewExecutionContextWithTraceID(workflowID, executionID string, input map[string]any, traceID string) *ExecutionContext {
	var ctx context.Context
	if traceID != "" {
		ctx = context.WithValue(context.Background(), trace.TraceIDKey, traceID)
	} else {
		ctx = context.Background()
	}
	return &ExecutionContext{
		WorkflowID:  workflowID,
		ExecutionID: executionID,
		TraceID:     traceID,
		Context:     ctx,
		Variables:   make(map[string]any),
		Input:       input,
		Output:      make(map[string]any),
		State:       make(map[string]string),
		StartTime:   time.Now(),
		Logger:      NewExecutionLogger(executionID, traceID),
	}
}

func (ctx *ExecutionContext) SetVariable(key string, value any) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.Variables[key] = value
}

func (ctx *ExecutionContext) GetVariable(key string) (any, bool) {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	v, ok := ctx.Variables[key]
	return v, ok
}

func (ctx *ExecutionContext) SetNodeState(nodeID, status string) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.State[nodeID] = status
}

func (ctx *ExecutionContext) GetNodeState(nodeID string) string {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.State[nodeID]
}

type ExecutionLogger struct {
	ExecutionID string
	TraceID     string // 新增：请求追踪 ID
	Logs        []LogEntry
	mu          sync.Mutex
}

type LogEntry struct {
	Time     time.Time `json:"time"`
	NodeID   string    `json:"node_id"`
	Level    string    `json:"level"`
	Message  string    `json:"message"`
	Duration int64     `json:"duration_ms"`
}

func NewExecutionLogger(executionID string, traceID string) *ExecutionLogger {
	return &ExecutionLogger{
		ExecutionID: executionID,
		TraceID:     traceID,
		Logs:        make([]LogEntry, 0),
	}
}

func (l *ExecutionLogger) Log(nodeID, level, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Logs = append(l.Logs, LogEntry{
		Time:    time.Now(),
		NodeID:  nodeID,
		Level:   level,
		Message: message,
	})
}

func (l *ExecutionLogger) Logf(nodeID, level, format string, args ...any) {
	l.Log(nodeID, level, fmt.Sprintf(format, args...))
}

type NodeExecutor interface {
	Execute(ctx *ExecutionContext, config json.RawMessage) (any, error)
}

type WorkflowDefinition struct {
	Nodes []NodeDefinition `json:"nodes"`
	Edges []EdgeDefinition `json:"edges"`
}

type NodeDefinition struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Name     string          `json:"name"`
	Config   json.RawMessage `json:"config"`
	Position Position        `json:"position"`
}

type EdgeDefinition struct {
	ID           string `json:"id"`
	Source       string `json:"source"`
	Target       string `json:"target"`
	SourceHandle string `json:"source_handle,omitempty"`
	TargetHandle string `json:"target_handle,omitempty"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type ExecutionResult struct {
	ExecutionID string         `json:"execution_id"`
	TraceID     string         `json:"trace_id"` // 新增：请求追踪 ID
	Status      string         `json:"status"`
	Output      map[string]any `json:"output"`
	Error       string         `json:"error,omitempty"`
	Logs        []LogEntry     `json:"logs"`
	Duration    int64          `json:"duration_ms"`
	StartedAt   *time.Time     `json:"started_at,omitempty"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
}
