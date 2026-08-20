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
	"cogniforge/internal/user"
)

func setupResetPasswordRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	authHandler := auth.NewAuthHandler()
	r.POST("/api/v1/auth/register", authHandler.Register)
	r.POST("/api/v1/auth/login", authHandler.Login)

	h := user.NewUserHandler()
	r.POST("/api/v1/admin/users/:id/reset-password", h.AdminResetPassword)
	return r
}

func TestAdminResetPassword_UserNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/users/no-such-user/reset-password", nil)
	setupResetPasswordRouter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAdminResetPassword_SuccessThenLogin(t *testing.T) {
	router := setupResetPasswordRouter()
	email := "admin-reset@example.com"
	token := registerUserAndGetToken(t, router, email, "password123", "Admin Reset")
	require.NotEmpty(t, token)

	userID := extractUserIDFromRegister(t, router, email, "password123")
	require.NotEmpty(t, userID)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/users/"+userID+"/reset-password", nil)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	tempPwd, ok := data["temporary_password"].(string)
	require.True(t, ok)
	require.NotEmpty(t, tempPwd)

	okStrength, _ := user.CheckPasswordStrength(tempPwd)
	require.True(t, okStrength)

	loginBody, _ := json.Marshal(auth.LoginRequest{Email: email, Password: tempPwd})
	loginReq, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)
	assert.Equal(t, http.StatusOK, loginW.Code, loginW.Body.String())
}

func extractUserIDFromRegister(t *testing.T, router *gin.Engine, email, password string) string {
	t.Helper()
	// 再注册同邮箱会失败；改为登录后解析 JWT 不方便。直接查库。
	svc := user.NewUserService()
	list, err := svc.ListUsers(&user.ListUsersRequest{Page: 1, PageSize: 100, Search: email})
	require.NoError(t, err)
	for _, u := range list.Users {
		if u.Email == email {
			return u.ID
		}
	}
	t.Fatalf("user %s not found", email)
	return ""
}
