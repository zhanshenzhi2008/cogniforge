package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"cogniforge/internal/handler"
	"cogniforge/internal/middleware"
	"cogniforge/internal/model"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupSessionTestRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Cors())

	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
	}

	settings := r.Group("/api/v1/settings")
	settings.Use(middleware.AuthRequired())
	{
		settings.GET("/sessions", handler.GetSessions)
		settings.DELETE("/sessions/:id", handler.RevokeSession)
	}

	return r
}

// ==================== GetSessions Tests ====================

func TestGetSessions_WithoutAuth(t *testing.T) {
	router := setupSessionTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/settings/sessions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetSessions_WithValidToken(t *testing.T) {
	router := setupSessionTestRouter()
	token := registerAndGetToken(t, router, "session-test@example.com", "password123", "Session Test User")

	// Login to create a session
	loginBody := handler.LoginRequest{
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

	var response model.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 2000, response.Code)

	// Verify sessions is an array
	sessionsData, ok := response.Data.([]interface{})
	assert.True(t, ok, "Data should be an array")
	assert.GreaterOrEqual(t, len(sessionsData), 1, "Should have at least one session from login")
}

func TestGetSessions_WithInvalidToken(t *testing.T) {
	router := setupSessionTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/settings/sessions", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetSessions_MultipleSessions(t *testing.T) {
	router := setupSessionTestRouter()

	// Register and login user 1
	registerAndGetToken(t, router, "multi1@example.com", "password123", "User 1")
	loginBody := handler.LoginRequest{Email: "multi1@example.com", Password: "password123"}
	body, _ := json.Marshal(loginBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	token1 := extractTokenFromResponse(w.Body.Bytes())

	// Register and login user 2
	registerAndGetToken(t, router, "multi2@example.com", "password123", "User 2")
	loginBody = handler.LoginRequest{Email: "multi2@example.com", Password: "password123"}
	body, _ = json.Marshal(loginBody)
	req, _ = http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	token2 := extractTokenFromResponse(w.Body.Bytes())

	// Register and login user 3
	registerAndGetToken(t, router, "multi3@example.com", "password123", "User 3")
	loginBody = handler.LoginRequest{Email: "multi3@example.com", Password: "password123"}
	body, _ = json.Marshal(loginBody)
	req, _ = http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	token3 := extractTokenFromResponse(w.Body.Bytes())

	// Get sessions for each user
	req, _ = http.NewRequest("GET", "/api/v1/settings/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+token1)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var response1 model.ApiResponse
	json.Unmarshal(w.Body.Bytes(), &response1)
	sessionsData1, ok := response1.Data.([]interface{})
	assert.True(t, ok)
	assert.GreaterOrEqual(t, len(sessionsData1), 1, "User 1 should have at least 1 session")

	req.Header.Set("Authorization", "Bearer "+token2)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var response2 model.ApiResponse
	json.Unmarshal(w.Body.Bytes(), &response2)
	sessionsData2, ok := response2.Data.([]interface{})
	assert.True(t, ok)
	assert.GreaterOrEqual(t, len(sessionsData2), 1, "User 2 should have at least 1 session")

	req.Header.Set("Authorization", "Bearer "+token3)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var response3 model.ApiResponse
	json.Unmarshal(w.Body.Bytes(), &response3)
	sessionsData3, ok := response3.Data.([]interface{})
	assert.True(t, ok)
	assert.GreaterOrEqual(t, len(sessionsData3), 1, "User 3 should have at least 1 session")

	// Verify total sessions across all users is at least 3
	totalSessions := len(sessionsData1) + len(sessionsData2) + len(sessionsData3)
	assert.GreaterOrEqual(t, totalSessions, 3, "Total sessions should be at least 3")
}

// ==================== RevokeSession Tests ====================

func TestRevokeSession_WithoutAuth(t *testing.T) {
	router := setupSessionTestRouter()

	req, _ := http.NewRequest("DELETE", "/api/v1/settings/sessions/some-session-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRevokeSession_WithInvalidToken(t *testing.T) {
	router := setupSessionTestRouter()

	req, _ := http.NewRequest("DELETE", "/api/v1/settings/sessions/some-session-id", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRevokeSession_NonexistentSession(t *testing.T) {
	router := setupSessionTestRouter()
	token := registerAndGetToken(t, router, "revoke-test@example.com", "password123", "Revoke Test User")

	// Try to revoke a non-existent session
	req, _ := http.NewRequest("DELETE", "/api/v1/settings/sessions/nonexistent-session-id", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRevokeSession_Success(t *testing.T) {
	router := setupSessionTestRouter()
	token := registerAndGetToken(t, router, "revoke-success@example.com", "password123", "Revoke Success User")

	// Login to create a session
	loginBody := handler.LoginRequest{
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

	var sessionsResponse model.ApiResponse
	json.Unmarshal(w.Body.Bytes(), &sessionsResponse)

	sessionsData, ok := sessionsResponse.Data.([]interface{})
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

	var revokeResponse model.ApiResponse
	json.Unmarshal(w.Body.Bytes(), &revokeResponse)
	assert.Equal(t, 2000, revokeResponse.Code)

	// Verify session is revoked (is_active should be false)
	req, _ = http.NewRequest("GET", "/api/v1/settings/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var updatedSessionsResponse model.ApiResponse
	json.Unmarshal(w.Body.Bytes(), &updatedSessionsResponse)

	// After revoking, the session should either be removed from the list
	// or marked as inactive
	updatedSessionsData, ok := updatedSessionsResponse.Data.([]interface{})
	assert.True(t, ok)

	// Find the revoked session and verify it's inactive
	for _, session := range updatedSessionsData {
		sessionMap, ok := session.(map[string]interface{})
		if ok {
			if id, ok := sessionMap["id"].(string); ok && id == sessionID {
				isActive, ok := sessionMap["is_active"].(bool)
				assert.True(t, ok)
				assert.False(t, isActive, "Revoked session should be inactive")
			}
		}
	}
}

func TestRevokeSession_EmptySessionId(t *testing.T) {
	router := setupSessionTestRouter()
	token := registerAndGetToken(t, router, "empty-id@example.com", "password123", "Empty ID User")

	// Try to revoke with empty session ID
	req, _ := http.NewRequest("DELETE", "/api/v1/settings/sessions/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 404 (route not matched) or 400 (bad request)
	assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusBadRequest)
}

func TestRevokeSession_CannotRevokeOtherUserSession(t *testing.T) {
	router := setupSessionTestRouter()

	// Register and login as user 1
	token1 := registerAndGetToken(t, router, "user1@example.com", "password123", "User One")

	// Get user 1's sessions to find a session ID
	req, _ := http.NewRequest("GET", "/api/v1/settings/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+token1)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var sessionsResponse model.ApiResponse
	json.Unmarshal(w.Body.Bytes(), &sessionsResponse)

	sessionsData, ok := sessionsResponse.Data.([]interface{})
	assert.True(t, ok)

	var user1SessionID string
	for _, session := range sessionsData {
		sessionMap, ok := session.(map[string]interface{})
		if ok {
			user1SessionID, _ = sessionMap["id"].(string)
			break
		}
	}

	// Register user 2
	token2 := registerAndGetToken(t, router, "user2@example.com", "password123", "User Two")

	// User 2 tries to revoke user 1's session
	req, _ = http.NewRequest("DELETE", "/api/v1/settings/sessions/"+user1SessionID, nil)
	req.Header.Set("Authorization", "Bearer "+token2)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 404 because user 2 doesn't own this session
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ==================== Session Data Structure Tests ====================

func TestSessionDataStructure(t *testing.T) {
	router := setupSessionTestRouter()
	token := registerAndGetToken(t, router, "structure-test@example.com", "password123", "Structure Test User")

	// Login
	loginBody := handler.LoginRequest{
		Email:    "structure-test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(loginBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Get sessions
	req, _ = http.NewRequest("GET", "/api/v1/settings/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response model.ApiResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	sessionsData, ok := response.Data.([]interface{})
	assert.True(t, ok)
	assert.GreaterOrEqual(t, len(sessionsData), 1)

	// Verify session structure
	session, ok := sessionsData[0].(map[string]interface{})
	assert.True(t, ok)

	// Required fields
	requiredFields := []string{"id", "user_id", "token_id", "device", "location", "ip_address", "expires_at", "last_used", "is_active", "created_at", "updated_at"}
	for _, field := range requiredFields {
		_, exists := session[field]
		assert.True(t, exists, "Session should have field: %s", field)
	}

	// Verify data types
	assert.IsType(t, "", session["id"])
	assert.IsType(t, "", session["user_id"])
	assert.IsType(t, "", session["device"])
	assert.IsType(t, true, session["is_active"])

	// Verify expires_at is a valid timestamp
	expiresAt, ok := session["expires_at"].(string)
	assert.True(t, ok)
	_, err := time.Parse(time.RFC3339, expiresAt)
	assert.NoError(t, err, "expires_at should be a valid RFC3339 timestamp")

	// Verify last_used is a valid timestamp
	lastUsed, ok := session["last_used"].(string)
	assert.True(t, ok)
	_, err = time.Parse(time.RFC3339, lastUsed)
	assert.NoError(t, err, "last_used should be a valid RFC3339 timestamp")
}
