package handler

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cogniforge/internal/database"
	"cogniforge/internal/model"
)

func ListWorkflows(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		model.Unauthorized(c, "unauthorized")
		return
	}

	var workflows []model.Workflow
	if err := database.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&workflows).Error; err != nil {
		model.InternalError(c, "查询工作流列表失败")
		return
	}

	model.Success(c, workflows)
}

func CreateWorkflow(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		model.Unauthorized(c, "unauthorized")
		return
	}

	var req struct {
		Name        string         `json:"name" binding:"required"`
		Description string         `json:"description"`
		Definition  map[string]any `json:"definition"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		model.BadRequest(c, "请求参数无效: "+err.Error())
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
		model.InternalError(c, "创建工作流失败")
		return
	}

	model.Created(c, workflow)
}

func GetWorkflow(c *gin.Context) {
	userID := c.GetString("user_id")
	workflowID := c.Param("id")

	var workflow model.Workflow
	if err := database.DB.Where("id = ? AND user_id = ?", workflowID, userID).First(&workflow).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			model.NotFound(c, "工作流不存在")
		} else {
			model.InternalError(c, "查询工作流失败")
		}
		return
	}

	model.Success(c, workflow)
}

func UpdateWorkflow(c *gin.Context) {
	userID := c.GetString("user_id")
	workflowID := c.Param("id")

	var workflow model.Workflow
	if err := database.DB.Where("id = ? AND user_id = ?", workflowID, userID).First(&workflow).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			model.NotFound(c, "工作流不存在")
		} else {
			model.InternalError(c, "查询工作流失败")
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
		model.BadRequest(c, "请求参数无效: "+err.Error())
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
			model.InternalError(c, "更新工作流失败")
			return
		}
	}

	database.DB.Model(&workflow).Update("version", workflow.Version+1)
	database.DB.First(&workflow, "id = ?", workflowID)
	model.Success(c, workflow)
}

func DeleteWorkflow(c *gin.Context) {
	userID := c.GetString("user_id")
	workflowID := c.Param("id")

	result := database.DB.Where("id = ? AND user_id = ?", workflowID, userID).Delete(&model.Workflow{})
	if result.Error != nil {
		model.InternalError(c, "删除工作流失败")
		return
	}
	if result.RowsAffected == 0 {
		model.NotFound(c, "工作流不存在")
		return
	}

	model.SuccessWithMessage(c, nil, "工作流已删除")
}

func ExecuteWorkflow(c *gin.Context) {
	userID := c.GetString("user_id")
	workflowID := c.Param("id")

	var workflow model.Workflow
	if err := database.DB.Where("id = ? AND user_id = ?", workflowID, userID).First(&workflow).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			model.NotFound(c, "工作流不存在")
		} else {
			model.InternalError(c, "查询工作流失败")
		}
		return
	}

	var req struct {
		Input map[string]any `json:"input"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Input = make(map[string]any)
	}

	now := time.Now()
	execution := model.WorkflowExecution{
		ID:         generateID(),
		WorkflowID: workflowID,
		UserID:     userID,
		Status:     "pending",
		Input:      model.JSONBMap(req.Input),
		StartedAt:  &now,
	}

	if err := database.DB.Create(&execution).Error; err != nil {
		model.InternalError(c, "创建执行记录失败")
		return
	}

	go func() {
		time.Sleep(2 * time.Second)
		database.DB.Model(&model.WorkflowExecution{}).
			Where("id = ?", execution.ID).
			Updates(map[string]any{
				"status": "completed",
				"output": `{"result": "workflow completed"}`,
			})
	}()

	model.Accepted(c, gin.H{
		"execution_id": execution.ID,
		"status":       execution.Status,
	})
}
