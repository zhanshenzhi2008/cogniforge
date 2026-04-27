package user_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"cogniforge/internal/auth"
)

// registerUserAndGetToken 注册用户并返回 token
func registerUserAndGetToken(t *testing.T, router *gin.Engine, email, password, name string) string {
	reqBody := auth.RegisterRequest{
		Email:    email,
		Password: password,
		Name:     name,
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to register user: status %d, body: %s", w.Code, w.Body.String())
	}
	return extractToken(w.Body.Bytes())
}

// extractToken 从响应中提取 token
func extractToken(body []byte) string {
	var resp map[string]interface{}
	json.Unmarshal(body, &resp)
	if data, ok := resp["data"].(map[string]interface{}); ok {
		if token, ok := data["token"].(string); ok {
			return token
		}
	}
	return ""
}
