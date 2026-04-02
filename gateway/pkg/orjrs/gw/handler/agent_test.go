package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orjrs/gateway/pkg/orjrs/gw/database"
	"github.com/orjrs/gateway/pkg/orjrs/gw/middleware"
	"github.com/orjrs/gateway/pkg/orjrs/gw/model"
	"github.com/stretchr/testify/assert"
)

var now = time.Now()

func setupAgentTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Mock auth middleware
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user-123")
		c.Next()
	})

	// Agent routes
	agents := r.Group("/v1/agents")
	{
		agents.GET("/", ListAgents)
		agents.POST("/", CreateAgent)
		agents.GET("/:id", GetAgent)
		agents.PUT("/:id", UpdateAgent)
		agents.DELETE("/:id", DeleteAgent)
	}

	return r
}

func TestListAgents(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// Setup
	router := setupAgentTestRouter()

	// Create a test agent first
	agent := model.Agent{
		ID:        "test-agent-1",
		UserID:    "test-user-123",
		Name:      "Test Agent",
		Model:     "gpt-4o",
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}
	database.DB.Create(&agent)

	// Test
	req, _ := http.NewRequest("GET", "/v1/agents/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response AgentListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.Data)

	// Cleanup
	database.DB.Delete(&agent)
}

func TestCreateAgent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	router := setupAgentTestRouter()

	tests := []struct {
		name       string
		request    CreateAgentRequest
		wantStatus int
	}{
		{
			name: "valid agent",
			request: CreateAgentRequest{
				Name:         "My Test Agent",
				Description:  "A test agent",
				Model:        "gpt-4o",
				SystemPrompt: "You are a helpful assistant.",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "missing name",
			request: CreateAgentRequest{
				Model: "gpt-4o",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing model",
			request: CreateAgentRequest{
				Name: "Test Agent",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest("POST", "/v1/agents/", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			// Cleanup if created
			if tt.wantStatus == http.StatusCreated {
				var response model.Agent
				json.Unmarshal(w.Body.Bytes(), &response)
				database.DB.Delete(&response)
			}
		})
	}
}

func TestGetAgent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	router := setupAgentTestRouter()

	// Create a test agent
	agent := model.Agent{
		ID:        "test-agent-get",
		UserID:    "test-user-123",
		Name:      "Get Test Agent",
		Model:     "gpt-4o",
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}
	database.DB.Create(&agent)

	// Test getting existing agent
	req, _ := http.NewRequest("GET", "/v1/agents/test-agent-get", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.Agent
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Get Test Agent", response.Name)

	// Test getting non-existing agent
	req2, _ := http.NewRequest("GET", "/v1/agents/non-existing", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusNotFound, w2.Code)

	// Cleanup
	database.DB.Delete(&agent)
}

func TestUpdateAgent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	router := setupAgentTestRouter()

	// Create a test agent
	agent := model.Agent{
		ID:        "test-agent-update",
		UserID:    "test-user-123",
		Name:      "Original Name",
		Model:     "gpt-4o",
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}
	database.DB.Create(&agent)

	// Test updating agent
	updateReq := UpdateAgentRequest{
		Name: "Updated Name",
	}
	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/v1/agents/test-agent-update", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.Agent
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", response.Name)

	// Cleanup
	database.DB.Delete(&agent)
}

func TestDeleteAgent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	router := setupAgentTestRouter()

	// Create a test agent
	agent := model.Agent{
		ID:        "test-agent-delete",
		UserID:    "test-user-123",
		Name:      "Delete Test Agent",
		Model:     "gpt-4o",
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}
	database.DB.Create(&agent)

	// Test deleting agent
	req, _ := http.NewRequest("DELETE", "/v1/agents/test-agent-delete", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify agent is deleted
	var deletedAgent model.Agent
	err := database.DB.Where("id = ?", "test-agent-delete").First(&deletedAgent).Error
	assert.Error(t, err) // Should not find the agent
}

func TestCreateAgentWithMemoryConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	router := setupAgentTestRouter()

	req := CreateAgentRequest{
		Name:         "Agent with Memory",
		Model:        "gpt-4o",
		SystemPrompt: "You are a helpful assistant.",
	}
	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", "/v1/agents/", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response model.Agent
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "long_term", response.MemoryType)
	assert.Equal(t, 10, response.MemoryTurns) // Default value

	// Cleanup
	database.DB.Delete(&response)
}

func TestCreateAgentWithTools(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	router := setupAgentTestRouter()

	req := CreateAgentRequest{
		Name:         "Agent with Tools",
		Model:        "gpt-4o",
		SystemPrompt: "You are a helpful assistant.",
		Tools:        []string{"web_search", "calculator", "code_executor"},
	}
	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", "/v1/agents/", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response model.Agent
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Tools, 3)

	// Cleanup
	database.DB.Delete(&response)
}

// Mock middleware for testing
var _ = middleware.AuthRequired
