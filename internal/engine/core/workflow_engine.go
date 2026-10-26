package core

import (
	"encoding/json"
	"fmt"
	"log/slog"
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
	return engine
}

func (e *WorkflowEngine) RegisterExecutor(nodeType string, executor NodeExecutor) {
	e.nodeExecutors[nodeType] = executor
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
		TraceID:     ctx.TraceID,
		Status:      "running",
		Output:      make(map[string]any),
		Logs:        ctx.Logger.Logs,
	}

	e.updateExecutionStatus(ctx.ExecutionID, "running", "", "")

	sortedNodes, err := e.TopologicalSort(def)
	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		e.updateExecutionStatus(ctx.ExecutionID, "failed", "", result.Error)
		return result
	}

	for _, node := range sortedNodes {
		if ctx.GetNodeState(node.ID) == "skipped" {
			e.persistNodeRun(ctx, node, "skipped", 1, nil, nil, "", 0)
			continue
		}

		ctx.NodeID = node.ID
		e.setCurrentNode(ctx.ExecutionID, node.ID)

		ctx.Logger.Logf(node.ID, "info", "Starting node: %s (type: %s)", node.Name, node.Type)
		if ctx.OnNodeStart != nil {
			ctx.OnNodeStart(node.ID)
		}

		executor, ok := e.nodeExecutors[node.Type]
		if !ok {
			err := fmt.Errorf("unknown node type: %s", node.Type)
			ctx.Logger.Logf(node.ID, "error", "Node execution failed: %v", err)
			e.persistNodeRun(ctx, node, "failed", 1, nil, nil, err.Error(), 0)
			result.Status = "failed"
			result.Error = err.Error()
			e.updateExecutionStatus(ctx.ExecutionID, "failed", "", result.Error)
			return result
		}

		startTime := time.Now()
		e.persistNodeRun(ctx, node, "running", 1, nil, nil, "", 0)
		output, err := executor.Execute(ctx, node.Config)
		duration := time.Since(startTime).Milliseconds()

		if err != nil {
			ctx.Logger.Logf(node.ID, "error", "Node execution failed after %dms: %v", duration, err)
			e.persistNodeRun(ctx, node, "failed", 1, nil, nil, err.Error(), duration)
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
		e.persistNodeRun(ctx, node, "succeeded", 1, nil, output, "", duration)
		if ctx.OnNodeEnd != nil {
			ctx.OnNodeEnd(node.ID, "completed")
		}
	}

	result.Status = "completed"
	result.Output = ctx.Output
	result.Duration = time.Since(ctx.StartTime).Milliseconds()
	e.setCurrentNode(ctx.ExecutionID, "")
	e.updateExecutionStatus(ctx.ExecutionID, "completed", result.Output, "")
	return result
}

func (e *WorkflowEngine) setCurrentNode(executionID, nodeID string) {
	if executionID == "" || database.DB == nil {
		return
	}
	database.DB.Model(&model.WorkflowExecution{}).Where("id = ?", executionID).
		Update("current_node", nodeID)
}

// persistNodeRun 写入/更新节点尝试记录（同一 execution+node+attempt 幂等更新）
func (e *WorkflowEngine) persistNodeRun(
	ctx *ExecutionContext,
	node NodeDefinition,
	status string,
	attempt int,
	input any,
	output any,
	errMsg string,
	durationMS int64,
) {
	if ctx == nil || ctx.ExecutionID == "" || database.DB == nil {
		return
	}
	now := time.Now()
	var existing model.WorkflowNodeRun
	q := database.DB.Where("execution_id = ? AND node_id = ? AND attempt = ?", ctx.ExecutionID, node.ID, attempt)
	err := q.First(&existing).Error
	outMap := toJSONBMap(output)
	inMap := toJSONBMap(input)

	if err == gorm.ErrRecordNotFound {
		run := model.WorkflowNodeRun{
			ID:          fmt.Sprintf("wnr_%d_%s", time.Now().UnixNano(), randomString(6)),
			ExecutionID: ctx.ExecutionID,
			NodeID:      node.ID,
			NodeType:    node.Type,
			Attempt:     attempt,
			Status:      status,
			Input:       inMap,
			Output:      outMap,
			Error:       errMsg,
			DurationMS:  durationMS,
			CreatedAt:   now,
		}
		if status == "running" {
			run.StartedAt = &now
		}
		if status == "succeeded" || status == "failed" || status == "skipped" {
			run.FinishedAt = &now
			if run.StartedAt == nil {
				run.StartedAt = &now
			}
		}
		if createErr := database.DB.Create(&run).Error; createErr != nil {
			slog.Warn("persist node run create failed", "error", createErr, "execution_id", ctx.ExecutionID, "node_id", node.ID)
		}
		return
	}
	if err != nil {
		slog.Warn("persist node run lookup failed", "error", err, "execution_id", ctx.ExecutionID, "node_id", node.ID)
		return
	}

	updates := map[string]any{
		"status":      status,
		"error":       errMsg,
		"duration_ms": durationMS,
		"node_type":   node.Type,
	}
	if inMap != nil {
		updates["input"] = inMap
	}
	if outMap != nil {
		updates["output"] = outMap
	}
	if status == "running" && existing.StartedAt == nil {
		updates["started_at"] = &now
	}
	if status == "succeeded" || status == "failed" || status == "skipped" {
		updates["finished_at"] = &now
	}
	if updErr := database.DB.Model(&existing).Updates(updates).Error; updErr != nil {
		slog.Warn("persist node run update failed", "error", updErr, "execution_id", ctx.ExecutionID, "node_id", node.ID)
	}
}

func toJSONBMap(v any) model.JSONBMap {
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]any); ok {
		return model.JSONBMap(m)
	}
	if m, ok := v.(model.JSONBMap); ok {
		return m
	}
	b, err := json.Marshal(v)
	if err != nil {
		return model.JSONBMap{"value": fmt.Sprint(v)}
	}
	var out model.JSONBMap
	if err := json.Unmarshal(b, &out); err != nil {
		return model.JSONBMap{"value": string(b)}
	}
	return out
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
		updates["current_node"] = ""
	}
	database.DB.Model(&model.WorkflowExecution{}).Where("id = ?", executionID).Updates(updates)
}

func (e *WorkflowEngine) ExecuteAsync(workflowID, userID string, input map[string]any) (string, error) {
	return e.ExecuteAsyncWithTraceID(workflowID, userID, input, "")
}

func (e *WorkflowEngine) ExecuteAsyncWithTraceID(workflowID, userID string, input map[string]any, traceID string) (string, error) {
	var workflow model.Workflow
	if err := database.DB.Where("id = ? AND user_id = ?", workflowID, userID).First(&workflow).Error; err != nil {
		return "", fmt.Errorf("workflow not found: %w", err)
	}

	executionID := generateID()
	now := time.Now()
	execution := model.WorkflowExecution{
		ID:         executionID,
		TraceID:    traceID,
		WorkflowID: workflowID,
		UserID:     userID,
		Status:     "pending",
		Input:      model.JSONBMap(input),
		StartedAt:  &now,
	}

	if err := database.DB.Create(&execution).Error; err != nil {
		return "", fmt.Errorf("failed to create execution: %w", err)
	}

	// 捕获 traceID 到 goroutine 中
	currentTraceID := traceID
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("workflow execution panicked", "workflow_id", workflowID, "execution_id", executionID, "error", r)
				database.DB.Model(&model.WorkflowExecution{}).
					Where("id = ?", executionID).
					Updates(map[string]any{
						"status": "failed",
						"error":  fmt.Sprintf("panic: %v", r),
					})
			}
		}()

		ctx := NewExecutionContextWithTraceID(workflowID, executionID, input, currentTraceID).WithUserID(userID)
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
		slog.Info("workflow execution completed",
			"workflow_id", workflowID,
			"execution_id", executionID,
			"trace_id", currentTraceID,
			"status", result.Status,
			"duration_ms", result.Duration)
	}()

	return executionID, nil
}

func generateID() string {
	return fmt.Sprintf("exec_%s%s", time.Now().Format("20060102150405"), randomString(8))
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	r := time.Now().UnixNano()
	for i := range b {
		r = r*31 + int64(letters[r%int64(len(letters))])
		b[i] = letters[r%int64(len(letters))]
	}
	return string(b)
}
