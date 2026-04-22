package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"cogniforge/internal/handler"
	"cogniforge/internal/middleware"
	"cogniforge/internal/model"
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

// ==================== Register Tests ====================

func TestRegister_Success(t *testing.T) {
	router := setupTestRouter()

	reqBody := handler.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
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
	assert.Equal(t, "test@example.com", response.User.Email)
	assert.Equal(t, "Test User", response.User.Name)
}

func TestRegister_InvalidEmail(t *testing.T) {
	router := setupTestRouter()

	reqBody := map[string]string{
		"email":    "invalid-email",
		"password": "password123",
		"name":     "Test User",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_ShortPassword(t *testing.T) {
	router := setupTestRouter()

	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "12345",
		"name":     "Test User",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_MissingFields(t *testing.T) {
	router := setupTestRouter()

	testCases := []struct {
		name string
		body map[string]string
	}{
		{"missing email", map[string]string{"password": "password123", "name": "Test"}},
		{"missing password", map[string]string{"email": "test@example.com", "name": "Test"}},
		{"missing name", map[string]string{"email": "test@example.com", "password": "password123"}},
		{"empty body", map[string]string{}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.body)
			req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestRegister_DuplicateUser(t *testing.T) {
	router := setupTestRouter()

	reqBody := handler.RegisterRequest{
		Email:    "duplicate@example.com",
		Password: "password123",
		Name:     "Test User",
	}
	body, _ := json.Marshal(reqBody)

	req1, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	req2, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusConflict, w2.Code)
}

// ==================== Login Tests ====================

func TestLogin_Success(t *testing.T) {
	router := setupTestRouter()

	registerBody := handler.RegisterRequest{
		Email:    "login@example.com",
		Password: "password123",
		Name:     "Login User",
	}
	body, _ := json.Marshal(registerBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	loginBody := handler.LoginRequest{
		Email:    "login@example.com",
		Password: "password123",
	}
	body, _ = json.Marshal(loginBody)
	req, _ = http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response handler.AuthResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response.Token)
}

func TestLogin_InvalidEmail(t *testing.T) {
	router := setupTestRouter()

	loginBody := map[string]string{
		"email":    "nonexistent@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(loginBody)

	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogin_WrongPassword(t *testing.T) {
	router := setupTestRouter()

	registerBody := handler.RegisterRequest{
		Email:    "wrongpass@example.com",
		Password: "correctpassword",
		Name:     "Test User",
	}
	body, _ := json.Marshal(registerBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

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
}

func TestLogin_MissingFields(t *testing.T) {
	router := setupTestRouter()

	testCases := []struct {
		name string
		body map[string]string
	}{
		{"missing email", map[string]string{"password": "password123"}},
		{"missing password", map[string]string{"email": "test@example.com"}},
		{"empty body", map[string]string{}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.body)
			req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

// ==================== Logout & GetCurrentUser Tests ====================

func TestLogout_Success(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("POST", "/api/v1/auth/logout", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetCurrentUser_WithValidToken(t *testing.T) {
	router := setupTestRouter()

	registerBody := handler.RegisterRequest{
		Email:    "currentuser@example.com",
		Password: "password123",
		Name:     "Current User",
	}
	body, _ := json.Marshal(registerBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var registerResponse handler.AuthResponse
	json.Unmarshal(w.Body.Bytes(), &registerResponse)

	req, _ = http.NewRequest("GET", "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+registerResponse.Token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var user model.User
	json.Unmarshal(w.Body.Bytes(), &user)
	assert.Equal(t, "currentuser@example.com", user.Email)
}

func TestGetCurrentUser_WithoutToken(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetCurrentUser_InvalidToken(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetCurrentUser_MalformedAuthHeader(t *testing.T) {
	router := setupTestRouter()

	testCases := []struct {
		name   string
		header string
	}{
		{"no bearer prefix", "some-token"},
		{"wrong prefix", "Basic some-token"},
		{"empty bearer", "Bearer "},
		{"bearer only", "Bearer"},
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
}

// ==================== JWT Token Tests ====================

func TestJWTTokenGeneration(t *testing.T) {
	router := setupTestRouter()

	registerBody := handler.RegisterRequest{
		Email:    "jwt@example.com",
		Password: "password123",
		Name:     "JWT User",
	}
	body, _ := json.Marshal(registerBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response handler.AuthResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	tokenParts := bytes.Split([]byte(response.Token), []byte("."))
	assert.Len(t, tokenParts, 3, "Token should have 3 parts")

	req, _ = http.NewRequest("GET", "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+response.Token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// ==================== Middleware Tests ====================

func TestCorsMiddleware(t *testing.T) {
	router := gin.New()
	router.Use(middleware.Cors())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
}

func TestAuthRequired_MissingHeader(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthRequired_InvalidFormat(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "InvalidFormat token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
