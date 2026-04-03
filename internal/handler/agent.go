package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"cogniforge/internal/database"
	"cogniforge/internal/model"
)

type CreateAgentRequest struct {
	Name         string   `json:"name" binding:"required"`
	Description  string   `json:"description"`
	Model        string   `json:"model" binding:"required"`
	SystemPrompt string   `json:"system_prompt"`
	Tools        []string `json:"tools"`
}

type UpdateAgentRequest struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Model        string   `json:"model"`
	SystemPrompt string   `json:"system_prompt"`
	Tools        []string `json:"tools"`
	Status       string   `json:"status"`
}

type AgentListResponse struct {
	Data []model.Agent `json:"data"`
}

func ListAgents(c *gin.Context) {
	userID := c.GetString("user_id")

	var agents []model.Agent
	if err := database.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&agents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agents"})
		return
	}

	c.JSON(http.StatusOK, AgentListResponse{Data: agents})
}

func CreateAgent(c *gin.Context) {
	userID := c.GetString("user_id")

	var req CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "名称不能为空"})
		return
	}
	if req.Model == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择模型"})
		return
	}

	tools := model.JSONBArray{}
	if req.Tools != nil {
		tools = req.Tools
	}

	agent := model.Agent{
		ID:           generateID(),
		UserID:       userID,
		Name:         req.Name,
		Description:  req.Description,
		Model:        req.Model,
		SystemPrompt: req.SystemPrompt,
		Tools:        tools,
		MemoryType:   "short_term",
		MemoryTurns:  10,
		InputFilter:  true,
		OutputFilter: true,
		Status:       "active",
		Metadata:     model.JSONBMap{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := database.DB.Create(&agent).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create agent", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, agent)
}

func GetAgent(c *gin.Context) {
	userID := c.GetString("user_id")
	agentID := c.Param("id")

	var agent model.Agent
	if err := database.DB.Where("id = ? AND user_id = ?", agentID, userID).First(&agent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agent"})
		}
		return
	}

	c.JSON(http.StatusOK, agent)
}

func UpdateAgent(c *gin.Context) {
	userID := c.GetString("user_id")
	agentID := c.Param("id")

	var agent model.Agent
	if err := database.DB.Where("id = ? AND user_id = ?", agentID, userID).First(&agent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agent"})
		}
		return
	}

	var req UpdateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		agent.Name = req.Name
	}
	if req.Description != "" {
		agent.Description = req.Description
	}
	if req.Model != "" {
		agent.Model = req.Model
	}
	if req.SystemPrompt != "" {
		agent.SystemPrompt = req.SystemPrompt
	}
	if req.Tools != nil {
		agent.Tools = req.Tools
	}
	if req.Status != "" {
		agent.Status = req.Status
	}
	agent.UpdatedAt = time.Now()

	if err := database.DB.Save(&agent).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update agent"})
		return
	}

	c.JSON(http.StatusOK, agent)
}

func DeleteAgent(c *gin.Context) {
	userID := c.GetString("user_id")
	agentID := c.Param("id")

	var agent model.Agent
	if err := database.DB.Where("id = ? AND user_id = ?", agentID, userID).First(&agent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agent"})
		}
		return
	}

	if err := database.DB.Delete(&agent).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete agent"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Agent deleted successfully"})
}

func AgentChat(c *gin.Context) {
	agentID := c.Param("id")
	userID := c.GetString("user_id")

	var agent model.Agent
	if err := database.DB.Where("id = ? AND user_id = ?", agentID, userID).First(&agent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agent"})
		}
		return
	}

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "messages is required and cannot be empty"})
		return
	}

	if req.Model == "" {
		req.Model = agent.Model
	}
	if req.Model == "" {
		req.Model = defaultModel()
	}

	systemPrompt := agent.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = "You are a helpful AI assistant."
	}

	messages := append([]ChatMessage{{Role: "system", Content: systemPrompt}}, req.Messages...)
	req.Messages = messages

	if req.Stream {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")

		if err := streamAIProvider(c, req); err != nil {
			slog.Error("streamAIProvider failed for agent chat",
				"error", err,
				"agent_id", agentID,
				"model", req.Model,
			)
			fmt.Fprintf(c.Writer, "data: {\"error\": \"AI provider error: %s\"}\n\n", err.Error())
			c.Writer.Flush()
		}
	} else {
		resp, err := callAIProvider(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "AI provider error: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}
