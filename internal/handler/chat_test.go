package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"cogniforge/internal/handler"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupChatRouter() *gin.Engine {
	r := gin.New()
	r.POST("/api/v1/chat/stream", handler.ChatStream)
	r.POST("/api/v1/models/chat", handler.Chat)
	return r
}

// ==================== Chat Non-Streaming Tests ====================

func TestChat_Success(t *testing.T) {
	router := setupChatRouter()

	reqBody := handler.ChatRequest{
		Model: "gpt-4o",
		Messages: []handler.ChatMessage{
			{Role: "user", Content: "Hello"},
		},
		Stream: false,
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/models/chat", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp handler.ChatResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "chat.completion", resp.Object)
	assert.Equal(t, "gpt-4o", resp.Model)
	assert.Len(t, resp.Choices, 1)
	assert.Equal(t, "assistant", resp.Choices[0].Message.Role)
	assert.NotEmpty(t, resp.Choices[0].Message.Content)
	assert.Equal(t, "stop", resp.Choices[0].FinishReason)
	assert.Greater(t, resp.Usage.TotalTokens, 0)
}

func TestChat_DefaultModel(t *testing.T) {
	router := setupChatRouter()

	reqBody := map[string]any{
		"messages": []map[string]string{{"role": "user", "content": "Hi"}},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/models/chat", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp handler.ChatResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "gpt-3.5-turbo", resp.Model)
}

func TestChat_MissingMessages(t *testing.T) {
	router := setupChatRouter()

	testCases := []struct {
		name string
		body map[string]any
	}{
		{"empty body", map[string]any{"model": "gpt-4o"}},
		{"empty messages array", map[string]any{"model": "gpt-4o", "messages": []any{}}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.body)
			req, _ := http.NewRequest("POST", "/api/v1/models/chat", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			var errResp map[string]string
			json.Unmarshal(w.Body.Bytes(), &errResp)
			assert.Contains(t, errResp["error"], "messages")
		})
	}
}

func TestChat_InvalidJSON(t *testing.T) {
	router := setupChatRouter()

	req, _ := http.NewRequest("POST", "/api/v1/models/chat", bytes.NewBuffer([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==================== ChatStream SSE Tests ====================

func TestChatStream_Success(t *testing.T) {
	router := setupChatRouter()

	reqBody := handler.ChatRequest{
		Model: "gpt-4o",
		Messages: []handler.ChatMessage{
			{Role: "user", Content: "Hello"},
		},
		Stream: true,
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/chat/stream", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Cache-Control"), "no-cache")

	bodyBytes := w.Body.Bytes()
	assert.True(t, strings.Contains(string(bodyBytes), "data: "))
	assert.True(t, strings.Contains(string(bodyBytes), `"model":"gpt-4o"`))
	assert.True(t, strings.Contains(string(bodyBytes), "data: [DONE]"))
}

func TestChatStream_DefaultModel(t *testing.T) {
	router := setupChatRouter()

	reqBody := map[string]any{
		"messages": []map[string]string{{"role": "user", "content": "Hi"}},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/chat/stream", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, strings.Contains(w.Body.String(), `"model":"gpt-3.5-turbo"`))
}

func TestChatStream_MissingMessages(t *testing.T) {
	router := setupChatRouter()

	reqBody := map[string]any{"model": "gpt-4o"}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/chat/stream", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChatStream_InvalidJSON(t *testing.T) {
	router := setupChatRouter()

	req, _ := http.NewRequest("POST", "/api/v1/chat/stream", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChatStream_SSEFormat(t *testing.T) {
	router := setupChatRouter()

	reqBody := handler.ChatRequest{
		Model: "gpt-4o",
		Messages: []handler.ChatMessage{
			{Role: "user", Content: "Hello"},
		},
		Stream: true,
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/chat/stream", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	bodyStr := w.Body.String()
	lines := strings.Split(bodyStr, "\n")

	dataLines := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "data: ") {
			dataLines++
			dataContent := strings.TrimPrefix(line, "data: ")
			if dataContent == "[DONE]" {
				continue
			}
			var event map[string]any
			err := json.Unmarshal([]byte(dataContent), &event)
			assert.NoError(t, err, "SSE data should be valid JSON: "+dataContent)
			assert.Contains(t, event, "choices")
		}
	}
	assert.Greater(t, dataLines, 0, "Should have at least one data line plus [DONE]")
}
