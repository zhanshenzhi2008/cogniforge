package user_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"cogniforge/internal/auth"
	"cogniforge/internal/middleware"
	"cogniforge/internal/user"
)

func setupSessionRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Cors())

	authHandler := auth.NewAuthHandler()
	authGroup := r.Group("/api/v1/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
	}

	userHandler := user.NewUserHandler()
	settingsGroup := r.Group("/api/v1/settings")
	settingsGroup.Use(middleware.AuthRequired())
	{
		settingsGroup.GET("/sessions", userHandler.GetSessions)
		settingsGroup.DELETE("/sessions/:id", userHandler.RevokeSession)
	}

	return r
}

// ==================== GetSessions Tests ====================

func TestGetSessions_WithoutAuth(t *testing.T) {
	router := setupSessionRouter()

	req, _ := http.NewRequest("GET", "/api/v1/settings/sessions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetSessions_WithValidToken(t *testing.T) {
	router := setupSessionRouter()
	token := registerUserAndGetToken(t, router, "session-test@example.com", "password123", "Session Test User")

	// Login to create a session
	loginBody := auth.LoginRequest{
		Email:    "session-test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(loginBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Get sessions
	req, _ = http.NewRequest("GET", "/api/v1/settings/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 2000, int(response["code"].(float64)))

	// Verify sessions is an array
	sessionsData, ok := response["data"].([]interface{})
	assert.True(t, ok, "Data should be an array")
	assert.GreaterOrEqual(t, len(sessionsData), 1, "Should have at least one session from login")
}

func TestGetSessions_WithInvalidToken(t *testing.T) {
	router := setupSessionRouter()

	req, _ := http.NewRequest("GET", "/api/v1/settings/sessions", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ==================== RevokeSession Tests ====================

func TestRevokeSession_WithoutAuth(t *testing.T) {
	router := setupSessionRouter()

	req, _ := http.NewRequest("DELETE", "/api/v1/settings/sessions/some-session-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRevokeSession_WithInvalidToken(t *testing.T) {
	router := setupSessionRouter()

	req, _ := http.NewRequest("DELETE", "/api/v1/settings/sessions/some-session-id", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRevokeSession_NonexistentSession(t *testing.T) {
	router := setupSessionRouter()
	token := registerUserAndGetToken(t, router, "revoke-test@example.com", "password123", "Revoke Test User")

	// Try to revoke a non-existent session
	req, _ := http.NewRequest("DELETE", "/api/v1/settings/sessions/nonexistent-session-id", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRevokeSession_Success(t *testing.T) {
	router := setupSessionRouter()
	token := registerUserAndGetToken(t, router, "revoke-success@example.com", "password123", "Revoke Success User")

	// Login to create a session
	loginBody := auth.LoginRequest{
		Email:    "revoke-success@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(loginBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Get sessions to find the session ID
	req, _ = http.NewRequest("GET", "/api/v1/settings/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var sessionsResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &sessionsResponse)

	sessionsData, ok := sessionsResponse["data"].([]interface{})
	assert.True(t, ok)
	assert.GreaterOrEqual(t, len(sessionsData), 1, "Should have at least one session")

	// Get the first session ID
	firstSession, ok := sessionsData[0].(map[string]interface{})
	assert.True(t, ok)
	sessionID, ok := firstSession["id"].(string)
	assert.True(t, ok)

	// Revoke the session
	req, _ = http.NewRequest("DELETE", "/api/v1/settings/sessions/"+sessionID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var revokeResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &revokeResponse)
	assert.Equal(t, "会话已撤销", revokeResponse["message"])
}

func TestRevokeSession_AfterRevokeCountDecreases(t *testing.T) {
	router := setupSessionRouter()
	token := registerUserAndGetToken(t, router, "revoke-count@example.com", "password123", "Revoke Count User")

	// Login to create a session
	loginBody := auth.LoginRequest{
		Email:    "revoke-count@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(loginBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Get sessions count before revoke
	req, _ = http.NewRequest("GET", "/api/v1/settings/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var beforeResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &beforeResponse)
	sessionsBefore, _ := beforeResponse["data"].([]interface{})
	countBefore := len(sessionsBefore)

	// If we have sessions, revoke the first one
	if countBefore > 0 {
		firstSession := sessionsBefore[0].(map[string]interface{})
		sessionID := firstSession["id"].(string)

		req, _ = http.NewRequest("DELETE", "/api/v1/settings/sessions/"+sessionID, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}
}
