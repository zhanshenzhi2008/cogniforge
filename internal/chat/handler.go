package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cogniforge/internal/memory"
	"cogniforge/internal/model"
	"cogniforge/internal/provider"
	"cogniforge/internal/quota"
	"cogniforge/internal/response"
)

// pythonServiceURL is set from config.RAG.PythonServiceURL.
var pythonServiceURL string

// SetPythonURL sets the Python service URL (called from router setup).
func SetPythonURL(url string) { pythonServiceURL = strings.TrimRight(url, "/") }
type ChatHandler struct {
	service *ChatService
	conv    *ConversationService
	quota   *quota.Service
}

func NewChatHandler(providerSvc *provider.Service, db *gorm.DB, quotaSvc *quota.Service) *ChatHandler {
	return &ChatHandler{
		service: NewChatService(providerSvc),
		conv:    NewConversationService(db),
		quota:   quotaSvc,
	}
}

func (h *ChatHandler) Service() *ChatService {
	return h.service
}

func (h *ChatHandler) ListModels(c *gin.Context) {
	result := h.service.ListModels()
	response.Success(c, result)
}

func (h *ChatHandler) GetModel(c *gin.Context) {
	response.Success(c, gin.H{"message": "Get model"})
}

func (h *ChatHandler) gate(c *gin.Context, source string) bool {
	if h.quota == nil {
		return true
	}
	userID := c.GetString("user_id")
	if userID == "" {
		return true
	}
	if err := h.quota.Allow(c.Request.Context(), userID, source); err != nil {
		quota.WriteError(c, err)
		return false
	}
	return true
}

func (h *ChatHandler) commit(c *gin.Context, source, model string, usage *ChatUsage, status string, started time.Time) {
	if h.quota == nil {
		return
	}
	userID := c.GetString("user_id")
	if userID == "" {
		return
	}
	in := quota.CommitInput{
		UserID:    userID,
		Source:    source,
		Model:     model,
		Status:    status,
		TraceID:   quota.TraceID(c),
		LatencyMS: time.Since(started).Milliseconds(),
	}
	if usage != nil {
		in.PromptTokens = usage.PromptTokens
		in.CompletionTokens = usage.CompletionTokens
		in.TotalTokens = usage.TotalTokens
		in.Estimated = usage.Estimated
	}
	h.quota.Commit(c.Request.Context(), in)
}

func (h *ChatHandler) refund(c *gin.Context) {
	if h.quota == nil {
		return
	}
	h.quota.RefundRequest(c.Request.Context(), c.GetString("user_id"))
}

func (h *ChatHandler) Chat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if len(req.Messages) == 0 {
		response.BadRequest(c, "messages 不能为空")
		return
	}
	// 服务端滑动窗口裁剪（阶段十四 14.1）
	windowMsgs := toWindowMessages(req.Messages)
	windowMsgs = memory.ApplyWindow(windowMsgs, req.MemoryTurns, req.MaxInputTokens)
	req.Messages = fromWindowMessages(windowMsgs)
	if !h.gate(c, "playground") {
		return
	}
	started := time.Now()
	resp, err := h.service.Chat(&req)
	if err != nil {
		h.refund(c)
		if WriteNoActiveProvider(c, err) {
			return
		}
		h.commit(c, "playground", req.Model, nil, "error", started)
		response.Fail(c, http.StatusBadGateway, "AI provider error: "+err.Error())
		return
	}
	usage := resp.Usage
	h.commit(c, "playground", req.Model, &usage, "ok", started)
	response.Success(c, resp)
}

func (h *ChatHandler) ChatStream(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if len(req.Messages) == 0 {
		response.BadRequest(c, "messages 不能为空")
		return
	}

	userID := c.GetString("user_id")

	// --- 阶段十四 14.3：加载历史 + 摘要注入 ---
	// 从 conversation 加载历史
	var existingMsgs []model.ConversationMessage
	var existingSummary string
	if req.ConversationID != "" && userID != "" {
		if conv, err := h.conv.Get(userID, req.ConversationID); err == nil {
			existingMsgs = conv.Messages
			existingSummary = conv.Summary
		}
	}

	// --- 阶段十四 14.4：长期记忆检索 ---
	longTermBlocks := []string{}
	if req.UseLongTerm && userID != "" {
		// 取最后一条 user 消息内容作为检索 query
		var lastUserContent string
		for i := len(req.Messages) - 1; i >= 0; i-- {
			if req.Messages[i].Role == "user" {
				lastUserContent = toStringContent(req.Messages[i].Content)
				break
			}
		}
		if lastUserContent != "" {
			longTermBlocks = h.searchLongTermMemory(c.Request.Context(), userID, lastUserContent, req.AgentID)
		}
	}

	// 组装完整 messages：[摘要?] + [长期记忆] + [历史] + [本轮]
	fullMsgs := buildFullMessages(existingSummary, longTermBlocks, existingMsgs, req.Messages)

	// --- 阶段十四 14.1：服务端滑动窗口裁剪 ---
	turns := req.MemoryTurns
	if turns == 0 {
		turns = 10
	}
	trimmed := memory.ApplyWindow(toWindowMessages(fullMsgs), turns, req.MaxInputTokens)
	trimmedMsgs := fromWindowMessages(trimmed)
	req.Messages = trimmedMsgs

	// --- 配额闸门 ---
	if !h.gate(c, "playground") {
		return
	}
	started := time.Now()

	// --- 调用 LLM ---
	usage, err := h.service.ChatStream(c, &req)
	if err != nil {
		h.refund(c)
		if WriteNoActiveProvider(c, err) {
			return
		}
		h.commit(c, "playground", req.Model, nil, "error", started)
		slog.Error("ChatStream failed",
			"error", err, "model", req.Model, "messages_count", len(req.Messages),
		)
		fmt.Fprintf(c.Writer, "data: {\"error\": \"AI provider error: %s\"}\n\n", err.Error())
		c.Writer.Flush()
		return
	}
	h.commit(c, "playground", req.Model, usage, "ok", started)

	// --- 阶段十四 14.3：事后保存 + 摘要 ---
	go h.saveConversationAsync(userID, req.ConversationID, req.Model,
		existingMsgs, existingSummary, trimmedMsgs, fullMsgs)
}

func WriteNoActiveProvider(c *gin.Context, err error) bool {
	if err == nil || !errors.Is(err, ErrNoActiveProvider) || c.Writer.Written() {
		return false
	}
	response.FailWithHTTPStatus(c, http.StatusServiceUnavailable, response.CodeNoActiveProvider, err.Error())
	return true
}

func (h *ChatHandler) Embeddings(c *gin.Context) {
	var req EmbeddingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if req.Input == nil {
		response.BadRequest(c, "input 不能为空")
		return
	}

	resp, err := h.service.Embeddings(&req)
	if err != nil {
		if WriteNoActiveProvider(c, err) {
			return
		}
		slog.Error("Embeddings failed", "error", err, "model", req.Model)
		response.FailWithHTTPStatus(c, http.StatusBadGateway, response.CodeAIProviderError, "AI provider error: "+err.Error())
		return
	}
	response.Success(c, resp)
}

func (h *ChatHandler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	rg.GET("/models", h.ListModels)
	rg.GET("/models/:id", h.GetModel)
	rg.POST("/embeddings", h.Embeddings)
}

func (h *ChatHandler) RegisterRoutes(rg *gin.RouterGroup) {
	h.RegisterPublicRoutes(rg)
}

// ============ memory.WindowMessage 与 ChatMessage 互转 ============

func toWindowMessages(in []ChatMessage) []memory.WindowMessage {
	out := make([]memory.WindowMessage, len(in))
	for i, m := range in {
		out[i] = memory.WindowMessage{Role: m.Role, Content: m.Content}
	}
	return out
}

func fromWindowMessages(in []memory.WindowMessage) []ChatMessage {
	out := make([]ChatMessage, len(in))
	for i, m := range in {
		out[i] = ChatMessage{Role: m.Role, Content: m.Content}
	}
	return out
}

// buildFullMessages 组装完整 messages：[摘要?] + [长期记忆] + [历史] + [本轮]
func buildFullMessages(summary string, longTermBlocks []string, history []model.ConversationMessage, current []ChatMessage) []ChatMessage {
	var msgs []ChatMessage
	// [0] 摘要 system 消息
	if summary != "" {
		msgs = append(msgs, ChatMessage{
			Role:    "system",
			Content: "【对话摘要】" + summary,
		})
	}
	// [1] 长期记忆块
	if len(longTermBlocks) > 0 {
		var b strings.Builder
		b.WriteString("【你的笔记本】以下是你之前记录的重要信息，回答时可参考（但不要复述）：\n")
		for i, block := range longTermBlocks {
			b.WriteString(fmt.Sprintf("%d. %s\n", i+1, block))
		}
		msgs = append(msgs, ChatMessage{Role: "system", Content: b.String()})
	}
	// [2] 历史消息
	for _, m := range history {
		msgs = append(msgs, ChatMessage{Role: m.Role, Content: m.Content})
	}
	// [3] 本轮消息
	for _, m := range current {
		msgs = append(msgs, m)
	}
	return msgs
}

// saveConversationAsync 异步保存消息到 conversation，更新摘要（阶段十四 14.3）
func (h *ChatHandler) saveConversationAsync(userID, conversationID, model string, existingMsgs []model.ConversationMessage, existingSummary string, trimmedMsgs, fullMsgs []ChatMessage) {
	if conversationID == "" || userID == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 追加本轮 user 消息到历史
	newMsgs := append(existingMsgs, toConversationMessages(trimmedMsgs)...)

	// 若 fullMsgs 比 trimmedMsgs 长，说明发生了裁剪，发给 Python 摘要
	var newSummary string
	if len(fullMsgs) > len(trimmedMsgs) {
		newSummary = h.summarizeAsync(ctx, existingSummary, trimmedMsgs, fullMsgs)
	} else {
		newSummary = existingSummary
	}

	// 更新 conversation
	h.conv.UpdateSummary(ctx, userID, conversationID, newMsgs, newSummary)
}

// summarizeAsync 调用 Python /api/memory/summarize 生成摘要（降级时不阻塞）
func (h *ChatHandler) summarizeAsync(ctx context.Context, existingSummary string, trimmedMsgs, fullMsgs []ChatMessage) string {
	if pythonServiceURL == "" {
		return existingSummary
	}

	// 被裁掉的消息（窗口外的旧消息）
	cutCount := len(fullMsgs) - len(trimmedMsgs)
	if cutCount <= 0 {
		return existingSummary
	}
	cutMsgs := make([]map[string]any, cutCount)
	for i, m := range fullMsgs[:cutCount] {
		cutMsgs[i] = map[string]any{"role": m.Role, "content": m.Content}
	}

	payload := map[string]any{
		"messages":         cutMsgs,
		"existing_summary": existingSummary,
	}
	body, _ := json.Marshal(payload)
	url := pythonServiceURL + "/api/memory/summarize"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		slog.Warn("summarize request build failed", "error", err)
		return existingSummary
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Warn("summarize call failed", "error", err)
		return existingSummary
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		slog.Warn("summarize non-200", "status", resp.StatusCode)
		return existingSummary
	}

	var result struct {
		Summary string `json:"summary"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		slog.Warn("summarize decode failed", "error", err)
		return existingSummary
	}
	if result.Summary == "" {
		return existingSummary
	}
	return result.Summary
}

func toConversationMessages(msgs []ChatMessage) []model.ConversationMessage {
	out := make([]model.ConversationMessage, len(msgs))
	for i, m := range msgs {
		out[i] = model.ConversationMessage{
			Role:    m.Role,
			Content: toStringContent(m.Content),
		}
	}
	return out
}

func toStringContent(content any) string {
	switch v := content.(type) {
	case string:
		return v
	case []any:
		var b strings.Builder
		for _, part := range v {
			if m, ok := part.(map[string]any); ok {
				if t, ok := m["text"].(string); ok {
					b.WriteString(t)
				}
			}
		}
		return b.String()
	default:
		return ""
	}
}

// searchLongTermMemory 调用 Python /api/memory/search 检索长期记忆（阶段十四 14.4）
func (h *ChatHandler) searchLongTermMemory(ctx context.Context, userID, query, agentID string) []string {
	if pythonServiceURL == "" {
		return nil
	}
	payload := map[string]any{
		"user_id": userID,
		"query":   query,
		"top_k":   5,
		"min_score": 0.4,
	}
	if agentID != "" {
		payload["agent_id"] = agentID
	}
	body, _ := json.Marshal(payload)
	url := pythonServiceURL + "/api/memory/search"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil
	}
	var result struct {
		Results []struct {
			Content string `json:"content"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}
	var blocks []string
	for _, r := range result.Results {
		if r.Content != "" {
			blocks = append(blocks, r.Content)
		}
	}
	return blocks
}
