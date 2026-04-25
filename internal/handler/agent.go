package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cogniforge/internal/database"
	"cogniforge/internal/model"
	"cogniforge/internal/response"
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

func ListAgents(c *gin.Context) {
	userID := c.GetString("user_id")

	var agents []model.Agent
	if err := database.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&agents).Error; err != nil {
		response.InternalError(c, "查询 Agent 列表失败")
		return
	}

	response.Success(c, agents)
}

func CreateAgent(c *gin.Context) {
	userID := c.GetString("user_id")

	var req CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Name == "" {
		response.BadRequest(c, "名称不能为空")
		return
	}
	if req.Model == "" {
		response.BadRequest(c, "请选择模型")
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
		response.InternalError(c, "创建 Agent 失败")
		return
	}

	response.Created(c, agent)
}

func GetAgent(c *gin.Context) {
	userID := c.GetString("user_id")
	agentID := c.Param("id")

	var agent model.Agent
	if err := database.DB.Where("id = ? AND user_id = ?", agentID, userID).First(&agent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "Agent 不存在")
		} else {
			response.InternalError(c, "查询 Agent 失败")
		}
		return
	}

	response.Success(c, agent)
}

func UpdateAgent(c *gin.Context) {
	userID := c.GetString("user_id")
	agentID := c.Param("id")

	var agent model.Agent
	if err := database.DB.Where("id = ? AND user_id = ?", agentID, userID).First(&agent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "Agent 不存在")
		} else {
			response.InternalError(c, "查询 Agent 失败")
		}
		return
	}

	var req UpdateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
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
		response.InternalError(c, "更新 Agent 失败")
		return
	}

	response.Success(c, agent)
}

func DeleteAgent(c *gin.Context) {
	userID := c.GetString("user_id")
	agentID := c.Param("id")

	var agent model.Agent
	if err := database.DB.Where("id = ? AND user_id = ?", agentID, userID).First(&agent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "Agent 不存在")
		} else {
			response.InternalError(c, "查询 Agent 失败")
		}
		return
	}

	if err := database.DB.Delete(&agent).Error; err != nil {
		response.InternalError(c, "删除 Agent 失败")
		return
	}

	response.SuccessWithMessage(c, nil, "Agent 已删除")
}

func AgentChat(c *gin.Context) {
	agentID := c.Param("id")
	userID := c.GetString("user_id")

	var agent model.Agent
	if err := database.DB.Where("id = ? AND user_id = ?", agentID, userID).First(&agent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "Agent 不存在")
		} else {
			response.InternalError(c, "查询 Agent 失败")
		}
		return
	}

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if len(req.Messages) == 0 {
		response.BadRequest(c, "messages 不能为空")
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
			response.Fail(c, http.StatusBadGateway, "AI provider error: "+err.Error())
			return
		}
		response.Success(c, resp)
	}
}
