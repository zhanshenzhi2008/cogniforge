package handler

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cogniforge/internal/database"
	"cogniforge/internal/engine"
	"cogniforge/internal/model"
	"cogniforge/internal/response"
)

var workflowEngine = engine.NewEngine()

func ListWorkflows(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var workflows []model.Workflow
	if err := database.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&workflows).Error; err != nil {
		response.InternalError(c, "查询工作流列表失败")
		return
	}

	response.Success(c, workflows)
}

func CreateWorkflow(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var req struct {
		Name        string         `json:"name" binding:"required"`
		Description string         `json:"description"`
		Definition  map[string]any `json:"definition"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

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

	if err := database.DB.Create(&workflow).Error; err != nil {
		response.InternalError(c, "创建工作流失败")
		return
	}

	response.Created(c, workflow)
}

func GetWorkflow(c *gin.Context) {
	userID := c.GetString("user_id")
	workflowID := c.Param("id")

	var workflow model.Workflow
	if err := database.DB.Where("id = ? AND user_id = ?", workflowID, userID).First(&workflow).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "工作流不存在")
		} else {
			response.InternalError(c, "查询工作流失败")
		}
		return
	}

	response.Success(c, workflow)
}

func UpdateWorkflow(c *gin.Context) {
	userID := c.GetString("user_id")
	workflowID := c.Param("id")

	var workflow model.Workflow
	if err := database.DB.Where("id = ? AND user_id = ?", workflowID, userID).First(&workflow).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "工作流不存在")
		} else {
			response.InternalError(c, "查询工作流失败")
		}
		return
	}

	var req struct {
		Name        string         `json:"name"`
		Description string         `json:"description"`
		Status      string         `json:"status"`
		Definition  map[string]any `json:"definition"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效: "+err.Error())
		return
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
		if err := database.DB.Model(&workflow).Updates(updates).Error; err != nil {
			response.InternalError(c, "更新工作流失败")
			return
		}
	}

	database.DB.Model(&workflow).Update("version", workflow.Version+1)
	database.DB.First(&workflow, "id = ?", workflowID)
	response.Success(c, workflow)
}

func DeleteWorkflow(c *gin.Context) {
	userID := c.GetString("user_id")
	workflowID := c.Param("id")

	result := database.DB.Where("id = ? AND user_id = ?", workflowID, userID).Delete(&model.Workflow{})
	if result.Error != nil {
		response.InternalError(c, "删除工作流失败")
		return
	}
	if result.RowsAffected == 0 {
		response.NotFound(c, "工作流不存在")
		return
	}

	response.SuccessWithMessage(c, nil, "工作流已删除")
}

func ExecuteWorkflow(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}

	workflowID := c.Param("id")

	var workflow model.Workflow
	if err := database.DB.Where("id = ? AND user_id = ?", workflowID, userID).First(&workflow).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "工作流不存在")
		} else {
			response.InternalError(c, "查询工作流失败")
		}
		return
	}

	var req struct {
		Input map[string]any `json:"input"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Input = make(map[string]any)
	}

	executionID, err := workflowEngine.ExecuteAsync(workflowID, userID, req.Input)
	if err != nil {
		response.InternalError(c, "执行工作流失败: "+err.Error())
		return
	}

	response.Accepted(c, gin.H{
		"execution_id": executionID,
		"status":       "pending",
	})
}

func ListWorkflowExecutions(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}

	workflowID := c.Param("id")

	var executions []model.WorkflowExecution
	if err := database.DB.Where("workflow_id = ? AND user_id = ?", workflowID, userID).
		Order("created_at DESC").
		Limit(50).
		Find(&executions).Error; err != nil {
		response.InternalError(c, "查询执行记录失败")
		return
	}

	response.Success(c, executions)
}

func GetWorkflowExecution(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}

	executionID := c.Param("executionId")

	var execution model.WorkflowExecution
	if err := database.DB.Where("id = ? AND user_id = ?", executionID, userID).First(&execution).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "执行记录不存在")
		} else {
			response.InternalError(c, "查询执行记录失败")
		}
		return
	}

	response.Success(c, execution)
}

func CancelWorkflowExecution(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}

	executionID := c.Param("executionId")

	var execution model.WorkflowExecution
	if err := database.DB.Where("id = ? AND user_id = ?", executionID, userID).First(&execution).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "执行记录不存在")
		} else {
			response.InternalError(c, "查询执行记录失败")
		}
		return
	}

	if execution.Status != "pending" && execution.Status != "running" {
		response.BadRequest(c, "当前状态无法取消")
		return
	}

	now := time.Now()
	database.DB.Model(&execution).Updates(map[string]any{
		"status":       "cancelled",
		"completed_at": &now,
		"error":        "Cancelled by user",
	})

	response.SuccessWithMessage(c, nil, "执行已取消")
}

func DebugWorkflow(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Unauthorized(c, "unauthorized")
		return
	}

	workflowID := c.Param("id")

	var workflow model.Workflow
	if err := database.DB.Where("id = ? AND user_id = ?", workflowID, userID).First(&workflow).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "工作流不存在")
		} else {
			response.InternalError(c, "查询工作流失败")
		}
		return
	}

	var req struct {
		Input           map[string]any `json:"input"`
		NodeBreakpoints []string       `json:"node_breakpoints"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Input = make(map[string]any)
	}

	executionID := generateID()
	now := time.Now()
	execution := model.WorkflowExecution{
		ID:         executionID,
		WorkflowID: workflowID,
		UserID:     userID,
		Status:     "debugging",
		Input:      model.JSONBMap(req.Input),
		StartedAt:  &now,
	}

	if err := database.DB.Create(&execution).Error; err != nil {
		response.InternalError(c, "创建调试会话失败")
		return
	}

	go func() {
		ctx := engine.NewExecutionContext(workflowID, executionID, req.Input)
		ctx.OnNodeStart = func(nodeID string) {
			updates := map[string]any{
				"current_node": nodeID,
				"status":       "running",
			}
			database.DB.Model(&model.WorkflowExecution{}).Where("id = ?", executionID).Updates(updates)
		}
		ctx.OnNodeEnd = func(nodeID string, status string) {
			updates := map[string]any{
				"current_node": nodeID,
				"status":       status,
			}
			database.DB.Model(&model.WorkflowExecution{}).Where("id = ?", executionID).Updates(updates)
		}

		def, err := workflowEngine.ParseDefinition(workflow.Definition)
		if err != nil {
			database.DB.Model(&model.WorkflowExecution{}).
				Where("id = ?", executionID).
				Updates(map[string]any{"status": "failed", "error": err.Error()})
			return
		}

		result := workflowEngine.Execute(ctx, def)
		_ = result
	}()

	response.Accepted(c, gin.H{
		"execution_id": executionID,
		"status":       "debugging",
	})
}
