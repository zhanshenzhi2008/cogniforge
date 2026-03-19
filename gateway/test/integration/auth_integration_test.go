package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/orjrs/gateway/pkg/orjrs/gw/handler"
	"github.com/orjrs/gateway/pkg/orjrs/gw/middleware"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestRouter() *gin.Engine {
	r := gin.New()
	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
		auth.POST("/logout", handler.Logout)
		auth.GET("/me", middleware.AuthRequired(), handler.GetCurrentUser)
	}
	return r
}

// ==================== Auth API Integration Tests ====================

func TestAuthAPI_CompleteFlow(t *testing.T) {
	router := setupTestRouter()

	t.Run("1. Register new user", func(t *testing.T) {
		reqBody := handler.RegisterRequest{
			Email:    "integration-test@example.com",
			Password: "password123",
			Name:     "Integration Test User",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response handler.AuthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.Token)
		assert.Equal(t, "integration-test@example.com", response.User.Email)
	})

	t.Run("2. Login with registered user", func(t *testing.T) {
		loginBody := handler.LoginRequest{
			Email:    "integration-test@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(loginBody)

		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response handler.AuthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.Token)
	})

	t.Run("3. Get current user with valid token", func(t *testing.T) {
		// First register to get token
		reqBody := handler.RegisterRequest{
			Email:    "current-user@example.com",
			Password: "password123",
			Name:     "Current User",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var registerResponse handler.AuthResponse
		json.Unmarshal(w.Body.Bytes(), &registerResponse)

		// Access protected endpoint
		req, _ = http.NewRequest("GET", "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+registerResponse.Token)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var user handler.User
		err := json.Unmarshal(w.Body.Bytes(), &user)
		assert.NoError(t, err)
		assert.Equal(t, "current-user@example.com", user.Email)
	})

	t.Run("4. Logout", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/auth/logout", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestAuthAPI_ErrorCases(t *testing.T) {
	router := setupTestRouter()

	t.Run("Register with invalid email", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    "invalid-email",
			"password": "password123",
			"name":     "Test",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Register with short password", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    "test@example.com",
			"password": "12345",
			"name":     "Test",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Register duplicate user", func(t *testing.T) {
		reqBody := handler.RegisterRequest{
			Email:    "duplicate@example.com",
			Password: "password123",
			Name:     "Test",
		}
		body, _ := json.Marshal(reqBody)

		// First registration
		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		// Duplicate registration
		req, _ = http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("Login with wrong password", func(t *testing.T) {
		// Register first
		reqBody := handler.RegisterRequest{
			Email:    "wrongpass@example.com",
			Password: "correctpassword",
			Name:     "Test",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Login with wrong password
		loginBody := handler.LoginRequest{
			Email:    "wrongpass@example.com",
			Password: "wrongpassword",
		}
		body, _ = json.Marshal(loginBody)
		req, _ = http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Login with non-existent user", func(t *testing.T) {
		loginBody := handler.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(loginBody)

		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Access protected endpoint without token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Access protected endpoint with invalid token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Access protected endpoint with malformed auth header", func(t *testing.T) {
		testCases := []struct {
			name   string
			header string
		}{
			{"no bearer prefix", "some-token"},
			{"wrong prefix", "Basic some-token"},
			{"empty bearer", "Bearer "},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
				req.Header.Set("Authorization", tc.header)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusUnauthorized, w.Code)
			})
		}
	})
}

func TestCorsMiddleware(t *testing.T) {
	router := gin.New()
	router.Use(middleware.Cors())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	t.Run("CORS headers on normal request", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
	})

	t.Run("CORS headers on OPTIONS preflight", func(t *testing.T) {
		req, _ := http.NewRequest("OPTIONS", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	})
}
