package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"cogniforge/internal/database"
	"cogniforge/internal/handler"
	"cogniforge/internal/model"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupWorkflowTestRouter() *gin.Engine {
	r := gin.New()

	db := database.GetTestDB()
	if db == nil {
		return r
	}

	db.AutoMigrate(&model.Workflow{}, &model.WorkflowExecution{})

	user := model.User{
		ID:       "test-user-workflow",
		Email:    "workflow-test@example.com",
		Name:     "Workflow Test",
		Password: "password123",
	}
	db.FirstOrCreate(&user)

	auth := r.Group("/api/v1/workflows")
	auth.Use(func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Next()
	})
	auth.GET("/", handler.ListWorkflows)
	auth.POST("/", handler.CreateWorkflow)
	auth.GET("/:id", handler.GetWorkflow)
	auth.PUT("/:id", handler.UpdateWorkflow)
	auth.DELETE("/:id", handler.DeleteWorkflow)
	auth.POST("/:id/execute", handler.ExecuteWorkflow)

	return r
}

func TestListWorkflows(t *testing.T) {
	db := database.GetTestDB()
	if db == nil {
		t.Skip("Test DB not available")
	}

	db.Exec("DELETE FROM workflows WHERE user_id = ?", "test-user-workflow")

	router := setupWorkflowTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/workflows/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string][]model.Workflow
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Empty(t, resp["data"])
}

func TestCreateWorkflow(t *testing.T) {
	db := database.GetTestDB()
	if db == nil {
		t.Skip("Test DB not available")
	}

	db.Exec("DELETE FROM workflows WHERE user_id = ?", "test-user-workflow")

	router := setupWorkflowTestRouter()

	body := map[string]any{
		"name":        "Test Workflow",
		"description": "A test workflow",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/v1/workflows/", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]model.Workflow
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Test Workflow", resp["data"].Name)
	assert.Equal(t, "draft", resp["data"].Status)
}

func TestCreateWorkflowValidation(t *testing.T) {
	db := database.GetTestDB()
	if db == nil {
		t.Skip("Test DB not available")
	}

	db.Exec("DELETE FROM workflows WHERE user_id = ?", "test-user-workflow")

	router := setupWorkflowTestRouter()

	body := map[string]any{
		"description": "Missing name",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/v1/workflows/", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetWorkflow(t *testing.T) {
	db := database.GetTestDB()
	if db == nil {
		t.Skip("Test DB not available")
	}

	db.Exec("DELETE FROM workflows WHERE user_id = ?", "test-user-workflow")

	router := setupWorkflowTestRouter()

	workflow := model.Workflow{
		ID:          "test-workflow-get",
		UserID:      "test-user-workflow",
		Name:        "Get Test Workflow",
		Description: "Testing GET",
		Status:      "draft",
	}
	db.Create(&workflow)

	req, _ := http.NewRequest("GET", "/api/v1/workflows/test-workflow-get", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]model.Workflow
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Get Test Workflow", resp["data"].Name)
}

func TestGetWorkflowNotFound(t *testing.T) {
	db := database.GetTestDB()
	if db == nil {
		t.Skip("Test DB not available")
	}

	db.Exec("DELETE FROM workflows WHERE user_id = ?", "test-user-workflow")

	router := setupWorkflowTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/workflows/non-existent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateWorkflow(t *testing.T) {
	db := database.GetTestDB()
	if db == nil {
		t.Skip("Test DB not available")
	}

	db.Exec("DELETE FROM workflows WHERE user_id = ?", "test-user-workflow")

	router := setupWorkflowTestRouter()

	workflow := model.Workflow{
		ID:          "test-workflow-update",
		UserID:      "test-user-workflow",
		Name:        "Original Name",
		Description: "Original Description",
		Status:      "draft",
	}
	db.Create(&workflow)

	body := map[string]any{
		"name":        "Updated Name",
		"description": "Updated Description",
		"status":      "published",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("PUT", "/api/v1/workflows/test-workflow-update", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]model.Workflow
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", resp["data"].Name)
	assert.Equal(t, "published", resp["data"].Status)
	assert.Equal(t, 2, resp["data"].Version)
}

func TestDeleteWorkflow(t *testing.T) {
	db := database.GetTestDB()
	if db == nil {
		t.Skip("Test DB not available")
	}

	db.Exec("DELETE FROM workflows WHERE user_id = ?", "test-user-workflow")

	router := setupWorkflowTestRouter()

	workflow := model.Workflow{
		ID:     "test-workflow-delete",
		UserID: "test-user-workflow",
		Name:   "Delete Test Workflow",
		Status: "draft",
	}
	db.Create(&workflow)

	req, _ := http.NewRequest("DELETE", "/api/v1/workflows/test-workflow-delete", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestExecuteWorkflow(t *testing.T) {
	db := database.GetTestDB()
	if db == nil {
		t.Skip("Test DB not available")
	}

	db.Exec("DELETE FROM workflows WHERE user_id = ?", "test-user-workflow")
	db.Exec("DELETE FROM workflow_executions WHERE user_id = ?", "test-user-workflow")

	router := setupWorkflowTestRouter()

	workflow := model.Workflow{
		ID:     "test-workflow-execute",
		UserID: "test-user-workflow",
		Name:   "Execute Test Workflow",
		Status: "published",
	}
	db.Create(&workflow)

	body := map[string]any{
		"input": map[string]any{"key": "value"},
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/v1/workflows/test-workflow-execute/execute", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp["execution_id"])
	assert.Equal(t, "pending", resp["status"])
}
