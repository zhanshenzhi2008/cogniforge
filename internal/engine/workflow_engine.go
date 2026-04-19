package engine

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"cogniforge/internal/database"
	"cogniforge/internal/model"
	"gorm.io/gorm"
)

type WorkflowEngine struct {
	db            *gorm.DB
	nodeExecutors map[string]NodeExecutor
}

func NewEngine() *WorkflowEngine {
	engine := &WorkflowEngine{
		db:            database.DB,
		nodeExecutors: make(map[string]NodeExecutor),
	}
	engine.registerDefaultExecutors()
	return engine
}

func (e *WorkflowEngine) registerDefaultExecutors() {
	e.nodeExecutors["start"] = &StartNodeExecutor{}
	e.nodeExecutors["end"] = &EndNodeExecutor{}
	e.nodeExecutors["llm"] = &LLMNodeExecutor{}
	e.nodeExecutors["agent"] = &AgentNodeExecutor{}
	e.nodeExecutors["condition"] = &ConditionNodeExecutor{}
	e.nodeExecutors["loop"] = &LoopNodeExecutor{}
	e.nodeExecutors["http"] = &HTTPNodeExecutor{}
	e.nodeExecutors["code"] = &CodeNodeExecutor{}
	e.nodeExecutors["delay"] = &DelayNodeExecutor{}
}

func (e *WorkflowEngine) RegisterExecutor(nodeType string, executor NodeExecutor) {
	e.nodeExecutors[nodeType] = executor
}

type NodeExecutor interface {
	Execute(ctx *ExecutionContext, config json.RawMessage) (any, error)
}

type ExecutionContext struct {
	WorkflowID  string
	ExecutionID string
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

func NewExecutionContext(workflowID, executionID string, input map[string]any) *ExecutionContext {
	return &ExecutionContext{
		WorkflowID:  workflowID,
		ExecutionID: executionID,
		Variables:   make(map[string]any),
		Input:       input,
		Output:      make(map[string]any),
		State:       make(map[string]string),
		StartTime:   time.Now(),
		Logger:      NewExecutionLogger(executionID),
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

func NewExecutionLogger(executionID string) *ExecutionLogger {
	return &ExecutionLogger{
		ExecutionID: executionID,
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
	ExecutionID string
	Status      string
	Output      map[string]any
	Error       string
	Logs        []LogEntry
	Duration    int64
}

func (e *WorkflowEngine) ParseDefinition(definition string) (*WorkflowDefinition, error) {
	var def WorkflowDefinition
	if err := json.Unmarshal([]byte(definition), &def); err != nil {
		return nil, fmt.Errorf("failed to parse workflow definition: %w", err)
	}
	return &def, nil
}

func (e *WorkflowEngine) TopologicalSort(def *WorkflowDefinition) ([]NodeDefinition, error) {
	nodeMap := make(map[string]*NodeDefinition)
	for i := range def.Nodes {
		nodeMap[def.Nodes[i].ID] = &def.Nodes[i]
	}

	inDegree := make(map[string]int)
	for _, node := range def.Nodes {
		inDegree[node.ID] = 0
	}
	for _, edge := range def.Edges {
		inDegree[edge.Target]++
	}

	var queue []string
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	var result []NodeDefinition
	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]
		if node, ok := nodeMap[nodeID]; ok {
			result = append(result, *node)
		}
		for _, edge := range def.Edges {
			if edge.Source == nodeID {
				inDegree[edge.Target]--
				if inDegree[edge.Target] == 0 {
					queue = append(queue, edge.Target)
				}
			}
		}
	}

	if len(result) != len(def.Nodes) {
		return nil, fmt.Errorf("cycle detected in workflow graph")
	}

	return result, nil
}

func (e *WorkflowEngine) Execute(ctx *ExecutionContext, def *WorkflowDefinition) *ExecutionResult {
	result := &ExecutionResult{
		ExecutionID: ctx.ExecutionID,
		Status:      "running",
		Output:      make(map[string]any),
		Logs:        ctx.Logger.Logs,
	}

	e.updateExecutionStatus(ctx.ExecutionID, "running", "", "")

	nodes, err := e.TopologicalSort(def)
	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		e.updateExecutionStatus(ctx.ExecutionID, "failed", "", result.Error)
		return result
	}

	for _, node := range nodes {
		if ctx.GetNodeState(node.ID) == "skipped" {
			continue
		}

		ctx.NodeID = node.ID

		ctx.Logger.Logf(node.ID, "info", "Starting node: %s (type: %s)", node.Name, node.Type)
		if ctx.OnNodeStart != nil {
			ctx.OnNodeStart(node.ID)
		}

		executor, ok := e.nodeExecutors[node.Type]
		if !ok {
			err := fmt.Errorf("unknown node type: %s", node.Type)
			ctx.Logger.Logf(node.ID, "error", "Node execution failed: %v", err)
			result.Status = "failed"
			result.Error = err.Error()
			e.updateExecutionStatus(ctx.ExecutionID, "failed", "", result.Error)
			return result
		}

		startTime := time.Now()
		output, err := executor.Execute(ctx, node.Config)
		duration := time.Since(startTime).Milliseconds()

		if err != nil {
			ctx.Logger.Logf(node.ID, "error", "Node execution failed after %dms: %v", duration, err)
			if ctx.OnNodeEnd != nil {
				ctx.OnNodeEnd(node.ID, "failed")
			}
			result.Status = "failed"
			result.Error = fmt.Sprintf("Node %s failed: %v", node.Name, err)
			e.updateExecutionStatus(ctx.ExecutionID, "failed", "", result.Error)
			return result
		}

		ctx.Logger.Logf(node.ID, "info", "Node completed in %dms", duration)
		if output != nil {
			ctx.Output[node.ID] = output
		}
		if ctx.OnNodeEnd != nil {
			ctx.OnNodeEnd(node.ID, "completed")
		}
	}

	result.Status = "completed"
	result.Output = ctx.Output
	result.Duration = time.Since(ctx.StartTime).Milliseconds()
	e.updateExecutionStatus(ctx.ExecutionID, "completed", result.Output, "")
	return result
}

func (e *WorkflowEngine) updateExecutionStatus(executionID, status string, output any, errorMsg string) {
	updates := map[string]any{"status": status}
	if output != nil {
		if outJSON, err := json.Marshal(output); err == nil {
			updates["output"] = string(outJSON)
		}
	}
	if errorMsg != "" {
		updates["error"] = errorMsg
	}
	if status == "completed" || status == "failed" {
		now := time.Now()
		updates["completed_at"] = &now
	}
	database.DB.Model(&model.WorkflowExecution{}).Where("id = ?", executionID).Updates(updates)
}

func (e *WorkflowEngine) ExecuteAsync(workflowID, userID string, input map[string]any) (string, error) {
	var workflow model.Workflow
	if err := database.DB.Where("id = ? AND user_id = ?", workflowID, userID).First(&workflow).Error; err != nil {
		return "", fmt.Errorf("workflow not found: %w", err)
	}

	executionID := generateID()
	now := time.Now()
	execution := model.WorkflowExecution{
		ID:         executionID,
		WorkflowID: workflowID,
		UserID:     userID,
		Status:     "pending",
		Input:      model.JSONBMap(input),
		StartedAt:  &now,
	}

	if err := database.DB.Create(&execution).Error; err != nil {
		return "", fmt.Errorf("failed to create execution: %w", err)
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Workflow execution panicked: %v", r)
				database.DB.Model(&model.WorkflowExecution{}).
					Where("id = ?", executionID).
					Updates(map[string]any{
						"status": "failed",
						"error":  fmt.Sprintf("panic: %v", r),
					})
			}
		}()

		ctx := NewExecutionContext(workflowID, executionID, input)
		def, err := e.ParseDefinition(workflow.Definition)
		if err != nil {
			database.DB.Model(&model.WorkflowExecution{}).
				Where("id = ?", executionID).
				Updates(map[string]any{
					"status": "failed",
					"error":  err.Error(),
				})
			return
		}

		result := e.Execute(ctx, def)
		log.Printf("Workflow %s execution %s: %s (duration: %dms)", workflowID, executionID, result.Status, result.Duration)
	}()

	return executionID, nil
}

func generateID() string {
	return fmt.Sprintf("exec_%s", fmt.Sprintf("%d%s", time.Now().UnixMilli(), randomString(8)))
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
