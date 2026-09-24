package user_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cogniforge/internal/auth"
	"cogniforge/internal/middleware"
	"cogniforge/internal/user"
)

func setupPasswordRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	authHandler := auth.NewAuthHandler()
	r.POST("/api/v1/auth/register", authHandler.Register)
	r.POST("/api/v1/auth/login", authHandler.Login)

	h := user.NewUserHandler()
	g := r.Group("/api/v1/settings")
	g.Use(middleware.AuthRequired())
	g.POST("/password", h.ChangePassword)
	return r
}

func postPassword(router *gin.Engine, token, oldPwd, newPwd string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(map[string]string{
		"old_password": oldPwd,
		"new_password": newPwd,
	})
	req, _ := http.NewRequest("POST", "/api/v1/settings/password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestChangePassword_WithoutAuth(t *testing.T) {
	w := postPassword(setupPasswordRouter(), "", "password123", "NewPass1!")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestChangePassword_SuccessThenLoginWithNew(t *testing.T) {
	router := setupPasswordRouter()
	token := registerUserAndGetToken(t, router, "pwd-ok@example.com", "password123", "Pwd Ok")

	w := postPassword(router, token, "password123", "NewPass1!")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 2000, int(resp["code"].(float64)))

	loginBody, _ := json.Marshal(auth.LoginRequest{Email: "pwd-ok@example.com", Password: "NewPass1!"})
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, req)
	assert.Equal(t, http.StatusOK, loginW.Code, loginW.Body.String())
}

func TestChangePassword_WrongOldPassword(t *testing.T) {
	router := setupPasswordRouter()
	token := registerUserAndGetToken(t, router, "pwd-wrong@example.com", "password123", "Pwd Wrong")

	w := postPassword(router, token, "not-the-old", "NewPass1!")
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 5011, int(resp["code"].(float64)))
}

func TestChangePassword_SamePassword(t *testing.T) {
	router := setupPasswordRouter()
	token := registerUserAndGetToken(t, router, "pwd-same@example.com", "password123", "Pwd Same")

	w := postPassword(router, token, "password123", "password123")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChangePassword_WeakNewPassword(t *testing.T) {
	router := setupPasswordRouter()
	token := registerUserAndGetToken(t, router, "pwd-weak@example.com", "password123", "Pwd Weak")

	w := postPassword(router, token, "password123", "short1")
	assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
}

func TestCheckPasswordStrength(t *testing.T) {
	ok, _ := user.CheckPasswordStrength("NewPass1!")
	assert.True(t, ok)

	ok, msg := user.CheckPasswordStrength("password123")
	assert.False(t, ok)
	assert.NotEmpty(t, msg)
}
