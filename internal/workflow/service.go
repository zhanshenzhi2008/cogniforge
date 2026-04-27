package workflow

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"cogniforge/internal/database"
	"cogniforge/internal/engine"
	"cogniforge/internal/model"
)

type WorkflowService struct {
	db             *gorm.DB
	workflowEngine *engine.WorkflowEngine
}

func NewWorkflowService() *WorkflowService {
	return &WorkflowService{
		db:             database.DB,
		workflowEngine: engine.NewEngine(),
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

	executionID, err := s.workflowEngine.ExecuteAsync(workflowID, userID, input)
	if err != nil {
		return nil, fmt.Errorf("执行工作流失败: %w", err)
	}

	return &ExecuteResponse{
		ExecutionID: executionID,
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

// CancelWorkflowExecution 取消执行
func (s *WorkflowService) CancelWorkflowExecution(userID, executionID string) error {
	var execution model.WorkflowExecution
	if err := s.db.Where("id = ? AND user_id = ?", executionID, userID).First(&execution).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("执行记录不存在")
		}
		return fmt.Errorf("查询执行记录失败")
	}

	if execution.Status != "pending" && execution.Status != "running" {
		return fmt.Errorf("当前状态无法取消")
	}

	now := time.Now()
	s.db.Model(&execution).Updates(map[string]any{
		"status":       "cancelled",
		"completed_at": &now,
		"error":        "Cancelled by user",
	})

	return nil
}

// DebugWorkflow 调试工作流
func (s *WorkflowService) DebugWorkflow(userID, workflowID string, req *DebugWorkflowRequest) (*ExecuteResponse, error) {
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
		ctx := engine.NewExecutionContext(workflowID, executionID, input)
		ctx.OnNodeStart = func(nodeID string) {
			updates := map[string]any{
				"current_node": nodeID,
				"status":       "running",
			}
			s.db.Model(&model.WorkflowExecution{}).Where("id = ?", executionID).Updates(updates)
		}
		ctx.OnNodeEnd = func(nodeID string, status string) {
			updates := map[string]any{
				"current_node": nodeID,
				"status":       status,
			}
			s.db.Model(&model.WorkflowExecution{}).Where("id = ?", executionID).Updates(updates)
		}

		def, err := s.workflowEngine.ParseDefinition(workflow.Definition)
		if err != nil {
			s.db.Model(&model.WorkflowExecution{}).
				Where("id = ?", executionID).
				Updates(map[string]any{"status": "failed", "error": err.Error()})
			return
		}

		result := s.workflowEngine.Execute(ctx, def)
		_ = result
	}()

	return &ExecuteResponse{
		ExecutionID: executionID,
		Status:      "debugging",
	}, nil
}

// generateID 生成唯一ID
func generateID() string {
	return newID()
}
