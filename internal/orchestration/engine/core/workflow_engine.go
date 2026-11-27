package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"cogniforge/internal/infra/database"
	"cogniforge/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const EngineVersion = "18.3"

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
	// 兼容前端 Vue Flow：nodes[].data 承载配置，引擎要 nodes[].config
	var raw struct {
		Nodes []struct {
			ID       string          `json:"id"`
			Type     string          `json:"type"`
			Name     string          `json:"name"`
			Config   json.RawMessage `json:"config"`
			Data     map[string]any  `json:"data"`
			Position Position        `json:"position"`
		} `json:"nodes"`
		Edges []EdgeDefinition `json:"edges"`
	}
	if err := json.Unmarshal([]byte(definition), &raw); err != nil {
		return nil, fmt.Errorf("failed to parse workflow definition: %w", err)
	}
	def := &WorkflowDefinition{Edges: raw.Edges, Nodes: make([]NodeDefinition, 0, len(raw.Nodes))}
	for _, n := range raw.Nodes {
		cfg := n.Config
		name := n.Name
		if len(cfg) == 0 || string(cfg) == "null" {
			cfg = vueDataToConfig(n.Type, n.Data)
		}
		if name == "" && n.Data != nil {
			if label, ok := n.Data["label"].(string); ok {
				name = label
			}
		}
		def.Nodes = append(def.Nodes, NodeDefinition{
			ID: n.ID, Type: n.Type, Name: name, Config: cfg, Position: n.Position,
		})
	}
	return def, nil
}

// vueDataToConfig 将画布 data 字段映射为节点 config JSON
func vueDataToConfig(nodeType string, data map[string]any) json.RawMessage {
	if data == nil {
		return json.RawMessage(`{}`)
	}
	out := map[string]any{}
	switch nodeType {
	case "start":
		if v, ok := data["variables"]; ok {
			out["variables"] = v
		}
	case "wait":
		if v, ok := data["waitReason"]; ok {
			out["reason"] = v
		} else if v, ok := data["reason"]; ok {
			out["reason"] = v
		}
		if v, ok := data["waitMessage"]; ok {
			out["message"] = v
		} else if v, ok := data["message"]; ok {
			out["message"] = v
		}
		if v, ok := data["timeoutSec"]; ok {
			out["timeout_sec"] = v
		} else if v, ok := data["timeout_sec"]; ok {
			out["timeout_sec"] = v
		}
	case "llm", "agent":
		if v, ok := data["model"]; ok {
			out["model"] = v
		}
		if v, ok := data["systemPrompt"]; ok {
			out["system_prompt"] = v
		}
		if v, ok := data["prompt"]; ok {
			out["prompt"] = v
		}
		if v, ok := data["temperature"]; ok {
			out["temperature"] = v
		}
		if v, ok := data["maxTokens"]; ok {
			out["max_tokens"] = v
		}
		if v, ok := data["stream"]; ok {
			out["stream"] = v
		} else if v, ok := data["llmStream"]; ok {
			out["stream"] = v
		}
		if v, ok := data["agentId"]; ok {
			out["agent_id"] = v
		}
	case "team":
		if v, ok := data["teamId"]; ok {
			out["team_id"] = v
		} else if v, ok := data["team_id"]; ok {
			out["team_id"] = v
		}
		if v, ok := data["task"]; ok {
			out["task"] = v
		}
		if v, ok := data["prompt"]; ok {
			out["prompt"] = v
			if _, has := out["task"]; !has {
				out["task"] = v
			}
		}
		if v, ok := data["goal"]; ok {
			out["goal"] = v
		}
	case "subflow":
		if v, ok := data["workflowId"]; ok {
			out["workflow_id"] = v
		} else if v, ok := data["workflow_id"]; ok {
			out["workflow_id"] = v
		} else if v, ok := data["subflowWorkflowId"]; ok {
			out["workflow_id"] = v
		}
		if v, ok := data["input"]; ok {
			out["input"] = v
		} else if v, ok := data["subflowInput"]; ok {
			out["input"] = v
		}
	case "search":
		if v, ok := data["knowledgeBaseId"]; ok {
			out["knowledge_base_id"] = v
		} else if v, ok := data["knowledge_base_id"]; ok {
			out["knowledge_base_id"] = v
		} else if v, ok := data["kbId"]; ok {
			out["knowledge_base_id"] = v
		}
		if v, ok := data["query"]; ok {
			out["query"] = v
		}
		if v, ok := data["topK"]; ok {
			out["top_k"] = v
		} else if v, ok := data["top_k"]; ok {
			out["top_k"] = v
		}
		if v, ok := data["minScore"]; ok {
			out["min_score"] = v
		} else if v, ok := data["min_score"]; ok {
			out["min_score"] = v
		}
		if v, ok := data["searchMode"]; ok {
			out["search_mode"] = v
		} else if v, ok := data["search_mode"]; ok {
			out["search_mode"] = v
		}
		if v, ok := data["rerank"]; ok {
			out["rerank"] = v
		}
		if v, ok := data["rerankProviderId"]; ok {
			out["rerank_provider_id"] = v
		} else if v, ok := data["rerank_provider_id"]; ok {
			out["rerank_provider_id"] = v
		}
		if v, ok := data["hybridSemanticWeight"]; ok {
			out["hybrid_semantic_weight"] = v
		} else if v, ok := data["hybrid_semantic_weight"]; ok {
			out["hybrid_semantic_weight"] = v
		}
	case "http":
		if v, ok := data["url"]; ok {
			out["url"] = v
		}
		if v, ok := data["method"]; ok {
			out["method"] = v
		}
		if v, ok := data["body"]; ok {
			out["body"] = v
		}
		if v, ok := data["timeout"]; ok {
			out["timeout"] = v
		} else if v, ok := data["timeoutSec"]; ok {
			out["timeout"] = v
		}
	case "delay":
		if v, ok := data["delayDuration"]; ok {
			out["duration"] = v
		} else if v, ok := data["duration"]; ok {
			out["duration"] = v
		}
		if v, ok := data["delayUnit"]; ok {
			out["unit"] = v
		} else if v, ok := data["unit"]; ok {
			out["unit"] = v
		}
	case "loop":
		if v, ok := data["loopType"]; ok {
			out["type"] = v
		} else if v, ok := data["type"]; ok {
			out["type"] = v
		}
		if v, ok := data["loopCount"]; ok {
			out["count"] = v
		} else if v, ok := data["count"]; ok {
			out["count"] = v
		}
		if v, ok := data["loopMaxIter"]; ok {
			out["max_iter"] = v
		} else if v, ok := data["max_iter"]; ok {
			out["max_iter"] = v
		}
		if v, ok := data["loopCondition"]; ok {
			out["condition"] = v
		} else if v, ok := data["condition"]; ok {
			out["condition"] = v
		}
	case "code":
		if v, ok := data["codeLanguage"]; ok {
			out["language"] = v
		} else if v, ok := data["language"]; ok {
			out["language"] = v
		}
		if v, ok := data["codeSource"]; ok {
			out["code"] = v
		} else if v, ok := data["code"]; ok {
			out["code"] = v
		}
		if v, ok := data["timeout"]; ok {
			out["timeout"] = v
		} else if v, ok := data["codeTimeout"]; ok {
			out["timeout"] = v
		}
	case "condition":
		if v, ok := data["conditions"]; ok {
			out["conditions"] = v
		} else if field, ok := data["conditionField"]; ok {
			op, _ := data["conditionOperator"].(string)
			if op == "" {
				op = "=="
			}
			branch, _ := data["conditionBranch"].(string)
			if branch == "" {
				branch = "true"
			}
			out["conditions"] = []map[string]any{{
				"field":    field,
				"operator": op,
				"value":    data["conditionValue"],
				"branch":   branch,
			}}
		}
	default:
		for k, v := range data {
			if k == "label" {
				continue
			}
			out[k] = v
		}
	}
	b, err := json.Marshal(out)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return b
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

	completed := make([]string, 0)
	for id, st := range ctx.State {
		if st == "skipped" || st == "succeeded" {
			completed = append(completed, id)
		}
	}

	for i, node := range sortedNodes {
		// resume 后已完成节点标为 skipped：只跳过，勿把历史 succeeded 改写成 skipped
		st := ctx.GetNodeState(node.ID)
		if st == "skipped" || st == "succeeded" {
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

		// wait 节点：写成功记录 + checkpoint，挂起为 waiting
		if reason, msg, timeoutSec, isWait := isWaitOutput(output); isWait {
			ctx.Logger.Logf(node.ID, "info", "Entering waiting: %s (%s) timeout=%ds", reason, msg, timeoutSec)
			if output != nil {
				ctx.Output[node.ID] = output
			}
			e.persistNodeRun(ctx, node, "succeeded", 1, nil, output, "", duration)
			completed = append(completed, node.ID)
			ctx.SetNodeState(node.ID, "succeeded")
			nextID := ""
			if i+1 < len(sortedNodes) {
				nextID = sortedNodes[i+1].ID
			}
			e.saveCheckpoint(ctx, nextID, completed, "enter_waiting")
			e.setWaiting(ctx.ExecutionID, reason, msg, timeoutSec)
			result.Status = "waiting"
			result.Output = ctx.Output
			result.Duration = time.Since(ctx.StartTime).Milliseconds()
			if ctx.OnNodeEnd != nil {
				ctx.OnNodeEnd(node.ID, "waiting")
			}
			return result
		}

		ctx.Logger.Logf(node.ID, "info", "Node completed in %dms", duration)
		if output != nil {
			ctx.Output[node.ID] = output
		}
		// end 节点勿把整包 ctx.Output 写入 node_run（易 jsonb 失败导致永久 running）
		persistOut := output
		if node.Type == "end" {
			persistOut = map[string]any{"finished": true}
		}
		e.persistNodeRun(ctx, node, "succeeded", 1, nil, persistOut, "", duration)
		completed = append(completed, node.ID)
		ctx.SetNodeState(node.ID, "succeeded")
		nextID := ""
		if i+1 < len(sortedNodes) {
			nextID = sortedNodes[i+1].ID
		}
		e.saveCheckpoint(ctx, nextID, completed, "node_ok")
		if ctx.OnNodeEnd != nil {
			ctx.OnNodeEnd(node.ID, "completed")
		}
	}

	result.Status = "completed"
	result.Output = ctx.Output
	result.Duration = time.Since(ctx.StartTime).Milliseconds()
	e.setCurrentNode(ctx.ExecutionID, "")
	e.saveCheckpoint(ctx, "", completed, "final")
	e.updateExecutionStatus(ctx.ExecutionID, "completed", result.Output, "")
	return result
}

func isWaitOutput(output any) (reason, message string, timeoutSec int, ok bool) {
	m, ok := output.(map[string]any)
	if !ok {
		return "", "", 0, false
	}
	flag, _ := m["__wait__"].(bool)
	if !flag {
		return "", "", 0, false
	}
	reason, _ = m["reason"].(string)
	message, _ = m["message"].(string)
	if reason == "" {
		reason = "approval"
	}
	switch v := m["timeout_sec"].(type) {
	case float64:
		timeoutSec = int(v)
	case int:
		timeoutSec = v
	}
	return reason, message, timeoutSec, true
}

func (e *WorkflowEngine) setWaiting(executionID, reason, message string, timeoutSec int) {
	if executionID == "" || database.DB == nil {
		return
	}
	updates := map[string]any{
		"status":      "waiting",
		"wait_reason": reason,
		"error":       message, // 复用 error 字段展示等待说明（非失败）
		"wait_until":  nil,
	}
	if timeoutSec > 0 {
		until := time.Now().Add(time.Duration(timeoutSec) * time.Second)
		updates["wait_until"] = &until
	}
	database.DB.Model(&model.WorkflowExecution{}).Where("id = ?", executionID).Updates(updates)
}

func (e *WorkflowEngine) saveCheckpoint(ctx *ExecutionContext, cursorNodeID string, completed []string, reason string) {
	if ctx == nil || ctx.ExecutionID == "" || database.DB == nil {
		return
	}
	seq := e.nextCheckpointSeq(ctx.ExecutionID)
	vars := model.JSONBMap{}
	ctx.mu.RLock()
	for k, v := range ctx.Variables {
		vars[k] = v
	}
	outs := model.JSONBMap{}
	for k, v := range ctx.Output {
		outs[k] = v
	}
	hash := ctx.DefinitionHash
	ctx.mu.RUnlock()

	cp := model.WorkflowCheckpoint{
		ID:               uuid.New().String(),
		ExecutionID:      ctx.ExecutionID,
		Seq:              seq,
		CursorNodeID:     cursorNodeID,
		Variables:        vars,
		Outputs:          outs,
		CompletedNodeIDs: model.JSONBArray(completed),
		Reason:           reason,
		EngineVersion:    EngineVersion,
		DefinitionHash:   hash,
		CreatedAt:        time.Now(),
	}
	if err := database.DB.Create(&cp).Error; err != nil {
		slog.Warn("save checkpoint failed", "error", err, "execution_id", ctx.ExecutionID, "seq", seq)
	}
}

func (e *WorkflowEngine) nextCheckpointSeq(executionID string) int {
	var maxSeq int
	database.DB.Model(&model.WorkflowCheckpoint{}).
		Where("execution_id = ?", executionID).
		Select("COALESCE(MAX(seq), 0)").
		Scan(&maxSeq)
	return maxSeq + 1
}

// HashDefinition 工作流定义内容哈希
func HashDefinition(definition string) string {
	sum := sha256.Sum256([]byte(definition))
	return hex.EncodeToString(sum[:8])
}

// ApplyCheckpoint 将 checkpoint 恢复到执行上下文（标记已完成节点为 skipped）
func ApplyCheckpoint(ctx *ExecutionContext, cp *model.WorkflowCheckpoint) {
	if ctx == nil || cp == nil {
		return
	}
	if cp.Variables != nil {
		for k, v := range cp.Variables {
			ctx.SetVariable(k, v)
		}
	}
	if cp.Outputs != nil {
		if ctx.Output == nil {
			ctx.Output = make(map[string]any)
		}
		for k, v := range cp.Outputs {
			ctx.Output[k] = v
		}
	}
	for _, id := range cp.CompletedNodeIDs {
		ctx.SetNodeState(id, "skipped")
	}
}

// ApplyStartFrom 从指定节点开始调试：祖先节点标 skipped（start 类型仍会执行，便于入参默认值）
// 返回被跳过的节点定义，供调用方写入 node_runs。
func ApplyStartFrom(ctx *ExecutionContext, def *WorkflowDefinition, startFromNodeID string) ([]NodeDefinition, error) {
	if ctx == nil || def == nil {
		return nil, fmt.Errorf("缺少执行上下文或定义")
	}
	if startFromNodeID == "" {
		return nil, nil
	}
	exists := false
	nodeByID := make(map[string]NodeDefinition, len(def.Nodes))
	for _, n := range def.Nodes {
		nodeByID[n.ID] = n
		if n.ID == startFromNodeID {
			exists = true
		}
	}
	if !exists {
		return nil, fmt.Errorf("起点节点不存在: %s", startFromNodeID)
	}

	// 反向边：收集 startFrom 的所有祖先
	preds := make(map[string][]string)
	for _, e := range def.Edges {
		preds[e.Target] = append(preds[e.Target], e.Source)
	}
	ancestors := map[string]bool{}
	stack := []string{startFromNodeID}
	seen := map[string]bool{startFromNodeID: true}
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, p := range preds[cur] {
			if seen[p] {
				continue
			}
			seen[p] = true
			ancestors[p] = true
			stack = append(stack, p)
		}
	}

	var skipped []NodeDefinition
	for id := range ancestors {
		n := nodeByID[id]
		if n.Type == "start" {
			continue // 仍执行开始节点，补默认入参
		}
		ctx.SetNodeState(id, "skipped")
		skipped = append(skipped, n)
	}
	return skipped, nil
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
			ID:          uuid.New().String(),
			ExecutionID: ctx.ExecutionID,
			NodeID:      node.ID,
			NodeType:    node.Type,
			Attempt:     attempt,
			Status:      status,
			Input:       inMap,
			Output:      outMap,
			Error:       errMsg,
			Detail: model.JSONBMap{
				"idempotency_key": fmt.Sprintf("%s:%s:%d", ctx.ExecutionID, node.ID, attempt),
			},
			DurationMS: durationMS,
			CreatedAt:  now,
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
		// jsonb 输出异常时仍要落终态，避免 end 等节点永久 running
		if status == "succeeded" || status == "failed" || status == "skipped" {
			fallback := map[string]any{
				"status":      status,
				"error":       errMsg,
				"duration_ms": durationMS,
				"finished_at": &now,
			}
			if fbErr := database.DB.Model(&existing).Updates(fallback).Error; fbErr != nil {
				slog.Warn("persist node run fallback status failed", "error", fbErr, "execution_id", ctx.ExecutionID, "node_id", node.ID)
			}
		}
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
	if executionID == "" || database.DB == nil {
		return
	}
	updates := map[string]any{"status": status}
	if output != nil {
		if outJSON, err := json.Marshal(output); err == nil {
			updates["output"] = string(outJSON)
		}
	}
	if errorMsg != "" {
		updates["error"] = errorMsg
	}
	if status == "completed" || status == "failed" || status == "cancelled" {
		now := time.Now()
		updates["completed_at"] = &now
		updates["current_node"] = ""
		updates["wait_reason"] = ""
		updates["wait_until"] = nil
	}
	if status == "running" {
		updates["wait_reason"] = ""
		updates["wait_until"] = nil
		updates["error"] = ""
	}
	database.DB.Model(&model.WorkflowExecution{}).Where("id = ?", executionID).Updates(updates)
	if status == "completed" || status == "failed" || status == "cancelled" {
		go fireWorkflowWebhook(executionID, status, errorMsg)
	}
}

// fireWorkflowWebhook 执行终态时 POST 工作流配置的 webhook_url（失败仅打日志）
func fireWorkflowWebhook(executionID, status, errMsg string) {
	if database.DB == nil || executionID == "" {
		return
	}
	var exec model.WorkflowExecution
	if err := database.DB.Where("id = ?", executionID).First(&exec).Error; err != nil {
		return
	}
	var wf model.Workflow
	if err := database.DB.Where("id = ?", exec.WorkflowID).First(&wf).Error; err != nil {
		return
	}
	url := strings.TrimSpace(wf.WebhookURL)
	if url == "" {
		return
	}
	body, _ := json.Marshal(map[string]any{
		"event":        "workflow.execution." + status,
		"execution_id": exec.ID,
		"workflow_id":  exec.WorkflowID,
		"status":       status,
		"error":        errMsg,
		"output":       exec.Output,
		"at":           time.Now().UTC().Format(time.RFC3339),
	})
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		slog.Warn("workflow webhook build request failed", "error", err, "execution_id", executionID)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		slog.Warn("workflow webhook post failed", "error", err, "execution_id", executionID, "url", url)
		return
	}
	_ = resp.Body.Close()
	if resp.StatusCode >= 300 {
		slog.Warn("workflow webhook non-2xx", "status", resp.StatusCode, "execution_id", executionID)
	}
}

// MaxSubflowDepth 子流程最大嵌套层数（根流程深度 0，其子流程为 1）
const MaxSubflowDepth = 3

// RunNested 同步跑嵌套工作流；写入独立 execution 记录，等待完成返回输出
func (e *WorkflowEngine) RunNested(parent *ExecutionContext, workflowID string, input map[string]any) (map[string]any, error) {
	if parent == nil {
		return nil, fmt.Errorf("subflow: missing parent context")
	}
	userID := parent.UserID
	if userID == "" {
		return nil, fmt.Errorf("subflow: missing user_id")
	}
	depth := parent.SubflowDepth + 1
	if depth > MaxSubflowDepth {
		return nil, fmt.Errorf("subflow nesting exceeds max depth %d", MaxSubflowDepth)
	}
	if workflowID == parent.WorkflowID {
		return nil, fmt.Errorf("subflow cannot call the same workflow")
	}

	var workflow model.Workflow
	if err := database.DB.Where("id = ? AND user_id = ?", workflowID, userID).First(&workflow).Error; err != nil {
		return nil, fmt.Errorf("subflow workflow not found: %w", err)
	}

	executionID := generateID()
	now := time.Now()
	execution := model.WorkflowExecution{
		ID:         executionID,
		TraceID:    parent.TraceID,
		WorkflowID: workflowID,
		UserID:     userID,
		Status:     "pending",
		Input:      model.JSONBMap(input),
		StartedAt:  &now,
	}
	if err := database.DB.Create(&execution).Error; err != nil {
		return nil, fmt.Errorf("subflow create execution: %w", err)
	}

	def, err := e.ParseDefinition(workflow.Definition)
	if err != nil {
		database.DB.Model(&model.WorkflowExecution{}).Where("id = ?", executionID).
			Updates(map[string]any{"status": "failed", "error": err.Error()})
		return nil, fmt.Errorf("subflow parse definition: %w", err)
	}

	ctx := NewExecutionContextWithTraceID(workflowID, executionID, input, parent.TraceID).WithUserID(userID)
	ctx.SubflowDepth = depth
	ctx.DefinitionHash = HashDefinition(workflow.Definition)

	result := e.Execute(ctx, def)
	if result.Status == "waiting" {
		return nil, fmt.Errorf("subflow entered waiting; nested wait/approval is not supported yet")
	}
	if result.Status != "completed" {
		errMsg := result.Error
		if errMsg == "" {
			errMsg = result.Status
		}
		return nil, fmt.Errorf("subflow failed: %s", errMsg)
	}
	if result.Output == nil {
		return map[string]any{}, nil
	}
	return result.Output, nil
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
	// 旧实现用 UnixNano*31 伪随机，int64 溢出后取模为负，letters[负下标] 直接 panic，
	// 表现是 POST /workflows/:id/execute 返回空 body 的 500；调试接口走 UUID 所以正常。
	return uuid.New().String()
}
