package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cogniforge/internal/database"
	"cogniforge/internal/model"
)

func ListWorkflows(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var workflows []model.Workflow
	if err := database.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&workflows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch workflows"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": workflows})
}

func CreateWorkflow(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Name        string         `json:"name" binding:"required"`
		Description string         `json:"description"`
		Definition  map[string]any `json:"definition"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create workflow"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": workflow})
}

func GetWorkflow(c *gin.Context) {
	userID := c.GetString("user_id")
	workflowID := c.Param("id")

	var workflow model.Workflow
	if err := database.DB.Where("id = ? AND user_id = ?", workflowID, userID).First(&workflow).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workflow not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch workflow"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": workflow})
}

func UpdateWorkflow(c *gin.Context) {
	userID := c.GetString("user_id")
	workflowID := c.Param("id")

	var workflow model.Workflow
	if err := database.DB.Where("id = ? AND user_id = ?", workflowID, userID).First(&workflow).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workflow not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch workflow"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update workflow"})
			return
		}
	}

	database.DB.Model(&workflow).Update("version", workflow.Version+1)
	database.DB.First(&workflow, "id = ?", workflowID)
	c.JSON(http.StatusOK, gin.H{"data": workflow})
}

func DeleteWorkflow(c *gin.Context) {
	userID := c.GetString("user_id")
	workflowID := c.Param("id")

	result := database.DB.Where("id = ? AND user_id = ?", workflowID, userID).Delete(&model.Workflow{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete workflow"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "workflow not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "workflow deleted successfully"})
}

func ExecuteWorkflow(c *gin.Context) {
	userID := c.GetString("user_id")
	workflowID := c.Param("id")

	var workflow model.Workflow
	if err := database.DB.Where("id = ? AND user_id = ?", workflowID, userID).First(&workflow).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workflow not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch workflow"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create execution"})
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

	c.JSON(http.StatusAccepted, gin.H{
		"execution_id": execution.ID,
		"status":       execution.Status,
	})
}
