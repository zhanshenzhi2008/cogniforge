package workflow

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"cogniforge/internal/database"
	"cogniforge/internal/engine"
	"cogniforge/internal/model"
)

type WorkflowService struct {
	db             *gorm.DB
	workflowEngine *engine.WorkflowEngine
	locker         *ExecLocker
}

func NewWorkflowService() *WorkflowService {
	return NewWorkflowServiceWithEngine(engine.NewEngine())
}

// NewWorkflowServiceWithEngine 注入已配置 AI 依赖的引擎（阶段十八 W0）
func NewWorkflowServiceWithEngine(eng *engine.WorkflowEngine) *WorkflowService {
	return NewWorkflowServiceWithDeps(eng, nil)
}

// NewWorkflowServiceWithDeps 注入引擎 + Redis（W3 resume 锁）
func NewWorkflowServiceWithDeps(eng *engine.WorkflowEngine, rdb *redis.Client) *WorkflowService {
	if eng == nil {
		eng = engine.NewEngine()
	}
	return &WorkflowService{
		db:             database.DB,
		workflowEngine: eng,
		locker:         NewExecLocker(rdb),
	}
}

// ListWorkflows 获取工作流列表
func (s *WorkflowService) ListWorkflows(userID string) ([]model.Workflow, error) {
	var workflows []model.Workflow
	if err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&workflows).Error; err != nil {
		return nil, fmt.Errorf("查询工作流列表失败")
	}
	return workflows, nil
}

// CreateWorkflow 创建工作流
func (s *WorkflowService) CreateWorkflow(userID string, req *CreateWorkflowRequest) (*model.Workflow, error) {
	workflow := model.Workflow{
		ID:          generateID(),
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Status:      "draft",
		Version:     1,
	}

	if req.Definition != nil {
		defJSON, _ := json.Marshal(req.Definition)
		workflow.Definition = string(defJSON)
	}

	if err := s.db.Create(&workflow).Error; err != nil {
		return nil, fmt.Errorf("创建工作流失败")
	}

	return &workflow, nil
}

// GetWorkflow 获取工作流详情
func (s *WorkflowService) GetWorkflow(userID, workflowID string) (*model.Workflow, error) {
	var workflow model.Workflow
	if err := s.db.Where("id = ? AND user_id = ?", workflowID, userID).First(&workflow).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("工作流不存在")
		}
		return nil, fmt.Errorf("查询工作流失败")
	}
	return &workflow, nil
}

// UpdateWorkflow 更新工作流
func (s *WorkflowService) UpdateWorkflow(userID, workflowID string, req *UpdateWorkflowRequest) (*model.Workflow, error) {
	var workflow model.Workflow
	if err := s.db.Where("id = ? AND user_id = ?", workflowID, userID).First(&workflow).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("工作流不存在")
		}
		return nil, fmt.Errorf("查询工作流失败")
	}

	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.Definition != nil {
		defJSON, _ := json.Marshal(req.Definition)
		updates["definition"] = string(defJSON)
	}

	if len(updates) > 0 {
		if err := s.db.Model(&workflow).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("更新工作流失败")
		}
	}

	s.db.Model(&workflow).Update("version", workflow.Version+1)
	s.db.First(&workflow, "id = ?", workflowID)
	return &workflow, nil
}

// DeleteWorkflow 删除工作流
func (s *WorkflowService) DeleteWorkflow(userID, workflowID string) error {
	result := s.db.Where("id = ? AND user_id = ?", workflowID, userID).Delete(&model.Workflow{})
	if result.Error != nil {
		return fmt.Errorf("删除工作流失败")
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("工作流不存在")
	}
	return nil
}

// ExecuteWorkflow 异步执行工作流
func (s *WorkflowService) ExecuteWorkflow(userID, workflowID string, req *ExecuteWorkflowRequest) (*ExecuteResponse, error) {
	return s.ExecuteWorkflowWithTraceID(userID, workflowID, req, "")
}

// ExecuteWorkflowWithTraceID 异步执行工作流（带 traceId）
func (s *WorkflowService) ExecuteWorkflowWithTraceID(userID, workflowID string, req *ExecuteWorkflowRequest, traceID string) (*ExecuteResponse, error) {
	var workflow model.Workflow
	if err := s.db.Where("id = ? AND user_id = ?", workflowID, userID).First(&workflow).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("工作流不存在")
		}
		return nil, fmt.Errorf("查询工作流失败")
	}

	input := req.Input
	if input == nil {
		input = make(map[string]any)
	}

	executionID, err := s.workflowEngine.ExecuteAsyncWithTraceID(workflowID, userID, input, traceID)
	if err != nil {
		return nil, fmt.Errorf("执行工作流失败: %w", err)
	}

	return &ExecuteResponse{
		ExecutionID: executionID,
		TraceID:     traceID,
		Status:      "pending",
	}, nil
}

// ListWorkflowExecutions 获取执行记录列表
func (s *WorkflowService) ListWorkflowExecutions(userID, workflowID string) ([]model.WorkflowExecution, error) {
	var executions []model.WorkflowExecution
	if err := s.db.Where("workflow_id = ? AND user_id = ?", workflowID, userID).
		Order("created_at DESC").
		Limit(50).
		Find(&executions).Error; err != nil {
		return nil, fmt.Errorf("查询执行记录失败")
	}
	return executions, nil
}

// GetWorkflowExecution 获取执行记录详情
func (s *WorkflowService) GetWorkflowExecution(userID, executionID string) (*model.WorkflowExecution, error) {
	var execution model.WorkflowExecution
	if err := s.db.Where("id = ? AND user_id = ?", executionID, userID).First(&execution).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("执行记录不存在")
		}
		return nil, fmt.Errorf("查询执行记录失败")
	}
	return &execution, nil
}

// GetWorkflowExecutionDetail 详情 + node_runs（阶段十八 W1）
func (s *WorkflowService) GetWorkflowExecutionDetail(userID, executionID string) (*ExecutionDetail, error) {
	exec, err := s.GetWorkflowExecution(userID, executionID)
	if err != nil {
		return nil, err
	}
	var runs []model.WorkflowNodeRun
	_ = s.db.Where("execution_id = ?", executionID).Order("created_at ASC").Find(&runs).Error
	var maxSeq int
	_ = s.db.Model(&model.WorkflowCheckpoint{}).
		Where("execution_id = ?", executionID).
		Select("COALESCE(MAX(seq), 0)").Scan(&maxSeq).Error
	return &ExecutionDetail{
		ID:                  exec.ID,
		WorkflowID:          exec.WorkflowID,
		UserID:              exec.UserID,
		Status:              exec.Status,
		WaitReason:          exec.WaitReason,
		WaitUntil:           exec.WaitUntil,
		Input:               exec.Input,
		Output:              exec.Output,
		Error:               exec.Error,
		CurrentNode:         exec.CurrentNode,
		TraceID:             exec.TraceID,
		StartedAt:           exec.StartedAt,
		CompletedAt:         exec.CompletedAt,
		CreatedAt:           exec.CreatedAt,
		UpdatedAt:           exec.UpdatedAt,
		NodeRuns:            runs,
		LatestCheckpointSeq: maxSeq,
	}, nil
}

// ListCheckpoints 检查点列表
func (s *WorkflowService) ListCheckpoints(userID, executionID string) ([]model.WorkflowCheckpoint, error) {
	if _, err := s.GetWorkflowExecution(userID, executionID); err != nil {
		return nil, err
	}
	var list []model.WorkflowCheckpoint
	if err := s.db.Where("execution_id = ?", executionID).Order("seq ASC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("查询检查点失败")
	}
	return list, nil
}

// ResumeExecution 从 checkpoint 继续（仅 waiting）
func (s *WorkflowService) ResumeExecution(userID, executionID string, req *ResumeExecutionRequest) (*ExecuteResponse, error) {
	return s.resumeWithExtras(userID, executionID, req, nil)
}

// ApproveExecution 审批通过 → resume（W3）
func (s *WorkflowService) ApproveExecution(userID, executionID string, req *ApproveExecutionRequest) (*ExecuteResponse, error) {
	exec, err := s.GetWorkflowExecution(userID, executionID)
	if err != nil {
		return nil, err
	}
	if exec.Status != "waiting" {
		return nil, fmt.Errorf("仅 waiting 可审批，当前: %s", exec.Status)
	}
	extras := map[string]any{
		"approval_decision": "approved",
		"approved_at":       time.Now().UTC().Format(time.RFC3339),
	}
	if req != nil && req.Comment != "" {
		extras["approval_comment"] = req.Comment
	}
	var resumeReq *ResumeExecutionRequest
	if req != nil && req.Seq != nil {
		resumeReq = &ResumeExecutionRequest{Seq: req.Seq}
	}
	return s.resumeWithExtras(userID, executionID, resumeReq, extras)
}

// RejectExecution 审批拒绝 → failed（W3）
func (s *WorkflowService) RejectExecution(userID, executionID string, req *RejectExecutionRequest) error {
	exec, err := s.GetWorkflowExecution(userID, executionID)
	if err != nil {
		return err
	}
	if exec.Status != "waiting" {
		return fmt.Errorf("仅 waiting 可拒绝，当前: %s", exec.Status)
	}
	reason := "Rejected by user"
	if req != nil && req.Reason != "" {
		reason = req.Reason
	}
	now := time.Now()
	return s.db.Model(&model.WorkflowExecution{}).Where("id = ?", executionID).Updates(map[string]any{
		"status":       "failed",
		"error":        reason,
		"wait_reason":  "",
		"wait_until":   nil,
		"current_node": "",
		"completed_at": &now,
	}).Error
}

// CallbackExecution 外部回调 approve/reject（W3；需登录，与审批同权）
func (s *WorkflowService) CallbackExecution(userID, executionID string, req *CallbackExecutionRequest) (*ExecuteResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("缺少回调参数")
	}
	switch req.Action {
	case "approve":
		return s.ApproveExecution(userID, executionID, &ApproveExecutionRequest{
			Comment: req.Comment,
			Seq:     req.Seq,
		})
	case "reject":
		reason := req.Reason
		if reason == "" {
			reason = "Rejected by callback"
		}
		if err := s.RejectExecution(userID, executionID, &RejectExecutionRequest{Reason: reason}); err != nil {
			return nil, err
		}
		return &ExecuteResponse{ExecutionID: executionID, Status: "failed"}, nil
	default:
		return nil, fmt.Errorf("action 仅支持 approve / reject")
	}
}

func (s *WorkflowService) resumeWithExtras(userID, executionID string, req *ResumeExecutionRequest, extras map[string]any) (*ExecuteResponse, error) {
	exec, err := s.GetWorkflowExecution(userID, executionID)
	if err != nil {
		return nil, err
	}
	if exec.Status != "waiting" {
		return nil, fmt.Errorf("仅 waiting 状态可 resume，当前: %s", exec.Status)
	}

	unlock, ok := s.locker.TryLock(executionID)
	if !ok {
		return nil, fmt.Errorf("执行正在恢复中，请勿重复操作")
	}

	var wf model.Workflow
	if err := s.db.Where("id = ? AND user_id = ?", exec.WorkflowID, userID).First(&wf).Error; err != nil {
		unlock()
		return nil, fmt.Errorf("工作流不存在")
	}

	var cp model.WorkflowCheckpoint
	q := s.db.Where("execution_id = ?", executionID)
	if req != nil && req.Seq != nil {
		q = q.Where("seq = ?", *req.Seq)
	} else {
		q = q.Order("seq DESC")
	}
	if err := q.First(&cp).Error; err != nil {
		unlock()
		return nil, fmt.Errorf("没有可用检查点")
	}

	input := map[string]any(exec.Input)
	if input == nil {
		input = map[string]any{}
	}

	s.db.Model(&model.WorkflowExecution{}).Where("id = ?", executionID).Updates(map[string]any{
		"status":       "running",
		"wait_reason":  "",
		"wait_until":   nil,
		"error":        "",
		"completed_at": nil,
	})

	go func() {
		defer unlock()
		defer func() {
			if r := recover(); r != nil {
				s.db.Model(&model.WorkflowExecution{}).Where("id = ?", executionID).
					Updates(map[string]any{"status": "failed", "error": fmt.Sprintf("panic: %v", r)})
			}
		}()
		ctx := engine.NewExecutionContextWithTraceID(exec.WorkflowID, executionID, input, exec.TraceID).WithUserID(userID)
		ctx.DefinitionHash = engine.HashDefinition(wf.Definition)
		engine.ApplyCheckpoint(ctx, &cp)
		for k, v := range extras {
			ctx.SetVariable(k, v)
		}
		def, err := s.workflowEngine.ParseDefinition(wf.Definition)
		if err != nil {
			s.db.Model(&model.WorkflowExecution{}).Where("id = ?", executionID).
				Updates(map[string]any{"status": "failed", "error": err.Error()})
			return
		}
		_ = s.workflowEngine.Execute(ctx, def)
	}()

	return &ExecuteResponse{ExecutionID: executionID, TraceID: exec.TraceID, Status: "running"}, nil
}

// ListNodeRuns 某次执行的节点尝试列表
func (s *WorkflowService) ListNodeRuns(userID, executionID string) ([]model.WorkflowNodeRun, error) {
	if _, err := s.GetWorkflowExecution(userID, executionID); err != nil {
		return nil, err
	}
	var runs []model.WorkflowNodeRun
	if err := s.db.Where("execution_id = ?", executionID).Order("created_at ASC").Find(&runs).Error; err != nil {
		return nil, fmt.Errorf("查询节点运行失败")
	}
	return runs, nil
}

// CancelWorkflowExecution 取消执行
func (s *WorkflowService) CancelWorkflowExecution(userID, executionID string) error {
	var execution model.WorkflowExecution
	if err := s.db.Where("id = ? AND user_id = ?", executionID, userID).First(&execution).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("执行记录不存在")
		}
		return fmt.Errorf("查询执行记录失败")
	}

	if execution.Status != "pending" && execution.Status != "running" && execution.Status != "waiting" && execution.Status != "debugging" {
		return fmt.Errorf("当前状态无法取消")
	}

	now := time.Now()
	s.db.Model(&execution).Updates(map[string]any{
		"status":       "cancelled",
		"completed_at": &now,
		"wait_reason":  "",
		"wait_until":   nil,
		"error":        "Cancelled by user",
		"current_node": "",
	})

	return nil
}

// DebugWorkflow 调试工作流
func (s *WorkflowService) DebugWorkflow(userID, workflowID string, req *DebugWorkflowRequest) (*ExecuteResponse, error) {
	return s.DebugWorkflowWithTraceID(userID, workflowID, req, "")
}

// DebugWorkflowWithTraceID 调试工作流（带 traceId）
func (s *WorkflowService) DebugWorkflowWithTraceID(userID, workflowID string, req *DebugWorkflowRequest, traceID string) (*ExecuteResponse, error) {
	var workflow model.Workflow
	if err := s.db.Where("id = ? AND user_id = ?", workflowID, userID).First(&workflow).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("工作流不存在")
		}
		return nil, fmt.Errorf("查询工作流失败")
	}

	input := req.Input
	if input == nil {
		input = make(map[string]any)
	}

	executionID := generateID()
	now := time.Now()
	execution := model.WorkflowExecution{
		ID:         executionID,
		WorkflowID: workflowID,
		UserID:     userID,
		Status:     "debugging",
		Input:      model.JSONBMap(input),
		StartedAt:  &now,
	}

	if err := s.db.Create(&execution).Error; err != nil {
		return nil, fmt.Errorf("创建调试会话失败")
	}

	go func() {
		ctx := engine.NewExecutionContextWithTraceID(workflowID, executionID, input, traceID).WithUserID(userID)
		ctx.DefinitionHash = engine.HashDefinition(workflow.Definition)
		// 节点进度由引擎写 current_node / node_runs；勿用节点 status 覆盖整次执行 status
		ctx.OnNodeStart = func(nodeID string) {
			s.db.Model(&model.WorkflowExecution{}).Where("id = ?", executionID).
				Updates(map[string]any{"current_node": nodeID, "status": "running"})
		}

		def, err := s.workflowEngine.ParseDefinition(workflow.Definition)
		if err != nil {
			s.db.Model(&model.WorkflowExecution{}).
				Where("id = ?", executionID).
				Updates(map[string]any{"status": "failed", "error": err.Error()})
			return
		}

		_ = s.workflowEngine.Execute(ctx, def)
	}()

	return &ExecuteResponse{
		ExecutionID: executionID,
		TraceID:     traceID,
		Status:      "debugging",
	}, nil
}

// generateID 生成唯一ID
func generateID() string {
	return newID()
}
