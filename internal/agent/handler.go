package agent

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cogniforge/internal/database"
	"cogniforge/internal/response"
)

type AgentHandler struct {
	service      *AgentService
	defaultModel string
}

func NewAgentHandler(defaultModel string) *AgentHandler {
	return &AgentHandler{
		service:      NewAgentService(),
		defaultModel: defaultModel,
	}
}

// ListAgents 获取 Agent 列表
func (h *AgentHandler) ListAgents(c *gin.Context) {
	userID := c.GetString("user_id")
	agents, err := h.service.ListAgents(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, agents)
}

// CreateAgent 创建 Agent
func (h *AgentHandler) CreateAgent(c *gin.Context) {
	userID := c.GetString("user_id")
	var req CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	agent, err := h.service.CreateAgent(userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, agent)
}

// GetAgent 获取 Agent 详情
func (h *AgentHandler) GetAgent(c *gin.Context) {
	userID := c.GetString("user_id")
	agentID := c.Param("id")

	agent, err := h.service.GetAgent(userID, agentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err.Error() == "Agent 不存在" {
			response.NotFound(c, err.Error())
		} else {
			response.InternalError(c, err.Error())
		}
		return
	}
	response.Success(c, agent)
}

// UpdateAgent 更新 Agent
func (h *AgentHandler) UpdateAgent(c *gin.Context) {
	userID := c.GetString("user_id")
	agentID := c.Param("id")

	var req UpdateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	agent, err := h.service.UpdateAgent(userID, agentID, &req)
	if err != nil {
		if err.Error() == "Agent 不存在" {
			response.NotFound(c, err.Error())
		} else {
			response.InternalError(c, err.Error())
		}
		return
	}
	response.Success(c, agent)
}

// DeleteAgent 删除 Agent
func (h *AgentHandler) DeleteAgent(c *gin.Context) {
	userID := c.GetString("user_id")
	agentID := c.Param("id")

	err := h.service.DeleteAgent(userID, agentID)
	if err != nil {
		if err.Error() == "Agent 不存在" {
			response.NotFound(c, err.Error())
		} else {
			response.InternalError(c, err.Error())
		}
		return
	}
	response.SuccessWithMessage(c, nil, "Agent 已删除")
}

// AgentChat Agent 对话
func (h *AgentHandler) AgentChat(c *gin.Context) {
	agentID := c.Param("id")
	userID := c.GetString("user_id")

	var agent struct {
		ID           string `json:"id"`
		Model        string `json:"model"`
		SystemPrompt string `json:"system_prompt"`
	}

	// 从数据库获取 Agent 信息
	var dbAgent struct {
		ID           string
		UserID       string
		Model        string
		SystemPrompt string
	}
	if err := database.DB.Where("id = ? AND user_id = ?", agentID, userID).First(&dbAgent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "Agent 不存在")
		} else {
			response.InternalError(c, "查询 Agent 失败")
		}
		return
	}

	agent.ID = dbAgent.ID
	agent.Model = dbAgent.Model
	agent.SystemPrompt = dbAgent.SystemPrompt

	var req struct {
		Model       string        `json:"model"`
		Messages    []ChatMessage `json:"messages" binding:"required"`
		Stream      bool          `json:"stream"`
		Temperature *float64      `json:"temperature,omitempty"`
		MaxTokens   *int          `json:"max_tokens,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if len(req.Messages) == 0 {
		response.BadRequest(c, "messages 不能为空")
		return
	}

	model := req.Model
	if model == "" {
		model = agent.Model
	}
	if model == "" {
		model = h.defaultModel
	}

	systemPrompt := agent.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = "You are a helpful AI assistant."
	}

	messages := append([]ChatMessage{{Role: "system", Content: systemPrompt}}, req.Messages...)

	// 构建请求
	chatReq := &ChatRequest{
		Model:       model,
		Messages:    messages,
		Stream:      req.Stream,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}

	if req.Stream {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")

		if err := h.streamChat(c, chatReq); err != nil {
			slog.Error("AgentChat stream failed",
				"error", err,
				"agent_id", agentID,
				"model", model,
			)
			fmt.Fprintf(c.Writer, "data: {\"error\": \"AI provider error: %s\"}\n\n", err.Error())
			c.Writer.Flush()
		}
	} else {
		resp, err := h.callChat(chatReq)
		if err != nil {
			response.Fail(c, http.StatusBadGateway, "AI provider error: "+err.Error())
			return
		}
		response.Success(c, resp)
	}
}

func (h *AgentHandler) callChat(req *ChatRequest) (*ChatResponse, error) {
	// 使用 chat 模块的服务
	// 这里暂时内联实现，后续可以注入 chat.Service
	return mockChatResponse(req)
}

func (h *AgentHandler) streamChat(c *gin.Context, req *ChatRequest) error {
	return mockStreamResponse(c, req)
}

// ChatMessage 和 ChatRequest 从 chat 模块复用
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Stream      bool          `json:"stream"`
	Temperature *float64      `json:"temperature,omitempty"`
	MaxTokens   *int          `json:"max_tokens,omitempty"`
}

type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int         `json:"index"`
		Message      ChatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
}

func mockChatResponse(req *ChatRequest) (*ChatResponse, error) {
	lastUserMsg := "hello"
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			lastUserMsg = req.Messages[i].Content
			break
		}
	}

	content := fmt.Sprintf("Mock response to: %s (model: %s)", lastUserMsg, req.Model)
	if len(content) > 500 {
		content = content[:500]
	}

	return &ChatResponse{
		ID:      fmt.Sprintf("chatcmpl-%d", 0),
		Object:  "chat.completion",
		Created: 0,
		Model:   req.Model,
		Choices: []struct {
			Index        int         `json:"index"`
			Message      ChatMessage `json:"message"`
			FinishReason string      `json:"finish_reason"`
		}{{Index: 0, Message: ChatMessage{Role: "assistant", Content: content}, FinishReason: "stop"}},
	}, nil
}

func mockStreamResponse(c *gin.Context, req *ChatRequest) error {
	lastUserMsg := "hello"
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			lastUserMsg = req.Messages[i].Content
			break
		}
	}

	fullText := fmt.Sprintf("Mock stream response to: %s (model: %s)", lastUserMsg, req.Model)
	const chunkRunes = 12
	runes := []rune(fullText)
	words := make([]string, 0, (len(runes)+chunkRunes-1)/chunkRunes)
	for i := 0; i < len(runes); i += chunkRunes {
		end := i + chunkRunes
		if end > len(runes) {
			end = len(runes)
		}
		words = append(words, string(runes[i:end]))
	}

	eventID := fmt.Sprintf("chatcmpl-%d", 0)
	for i, word := range words {
		finish := ""
		if i == len(words)-1 {
			finish = "stop"
		}
		fmt.Fprintf(c.Writer, "data: {\"id\":\"%s\",\"object\":\"chat.completion.chunk\",\"created\":0,\"model\":\"%s\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"%s\"},\"finish_reason\":\"%s\"}]}\n\n",
			eventID, req.Model, word, finish)
		c.Writer.Flush()
	}
	fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
	c.Writer.Flush()
	return nil
}

// RegisterRoutes 注册路由
func (h *AgentHandler) RegisterRoutes(rg *gin.RouterGroup) {
	agents := rg.Group("/agents")
	{
		agents.GET("", h.ListAgents)
		agents.POST("", h.CreateAgent)
		agents.GET("/:id", h.GetAgent)
		agents.PUT("/:id", h.UpdateAgent)
		agents.DELETE("/:id", h.DeleteAgent)
		agents.POST("/:id/chat", h.AgentChat)
	}
}
