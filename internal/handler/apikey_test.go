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
	"cogniforge/internal/middleware"
	"cogniforge/internal/model"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupApikeyRouter() *gin.Engine {
	r := gin.New()
	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
	}
	apikeys := r.Group("/api/v1/apikeys")
	apikeys.Use(middleware.AuthRequired())
	{
		apikeys.POST("", handler.CreateApiKey)
		apikeys.GET("", handler.ListApiKeys)
		apikeys.DELETE("/:id", handler.DeleteApiKey)
	}
	return r
}

func registerAndGetToken(t *testing.T, router *gin.Engine, email, password, name string) string {
	reqBody := handler.RegisterRequest{
		Email:    email,
		Password: password,
		Name:     name,
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp handler.AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	return resp.Token
}

func TestCreateApiKey_Success(t *testing.T) {
	db := database.GetTestDB()
	if db == nil {
		t.Skip("Test DB not available")
	}
	router := setupApikeyRouter()
	token := registerAndGetToken(t, router, "createkey@example.com", "password123", "Create Key User")

	reqBody := handler.ApiKeyRequest{Name: "My First Key"}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/apikeys", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp["key"])
	assert.Contains(t, resp["key"], "sk-")
	assert.NotEmpty(t, resp["id"])
	assert.Equal(t, "My First Key", resp["name"])
}

func TestCreateApiKey_MissingName(t *testing.T) {
	db := database.GetTestDB()
	if db == nil {
		t.Skip("Test DB not available")
	}
	router := setupApikeyRouter()
	token := registerAndGetToken(t, router, "missingname@example.com", "password123", "Missing Name")

	reqBody := map[string]string{}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/apikeys", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateApiKey_WithoutToken(t *testing.T) {
	router := setupApikeyRouter()

	reqBody := handler.ApiKeyRequest{Name: "No Auth Key"}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/apikeys", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestListApiKeys_Empty(t *testing.T) {
	db := database.GetTestDB()
	if db == nil {
		t.Skip("Test DB not available")
	}
	router := setupApikeyRouter()
	token := registerAndGetToken(t, router, "listempty@example.com", "password123", "List Empty")

	req, _ := http.NewRequest("GET", "/api/v1/apikeys", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotNil(t, resp["keys"])
}

func TestListApiKeys_OnlyOwnKeys(t *testing.T) {
	db := database.GetTestDB()
	if db == nil {
		t.Skip("Test DB not available")
	}
	router := setupApikeyRouter()
	token1 := registerAndGetToken(t, router, "keyowner1@example.com", "password123", "Key Owner 1")
	token2 := registerAndGetToken(t, router, "keyowner2@example.com", "password123", "Key Owner 2")

	reqBody := handler.ApiKeyRequest{Name: "Owner 1 Key"}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/apikeys", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token1)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	req, _ = http.NewRequest("GET", "/api/v1/apikeys", nil)
	req.Header.Set("Authorization", "Bearer "+token2)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	keys := resp["keys"].([]interface{})
	assert.Len(t, keys, 0, "Owner 2 should not see Owner 1's keys")
}

func TestDeleteApiKey_Success(t *testing.T) {
	db := database.GetTestDB()
	if db == nil {
		t.Skip("Test DB not available")
	}
	router := setupApikeyRouter()
	token := registerAndGetToken(t, router, "deletesuccess@example.com", "password123", "Delete Success")

	reqBody := handler.ApiKeyRequest{Name: "To Delete"}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/apikeys", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var createResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResp)
	keyID := createResp["id"].(string)

	req, _ = http.NewRequest("DELETE", "/api/v1/apikeys/"+keyID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var deleteResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &deleteResp)
	assert.Equal(t, "API key revoked successfully", deleteResp["message"])
}

func TestDeleteApiKey_NotFound(t *testing.T) {
	db := database.GetTestDB()
	if db == nil {
		t.Skip("Test DB not available")
	}
	router := setupApikeyRouter()
	token := registerAndGetToken(t, router, "deletenotfound@example.com", "password123", "Delete Not Found")

	req, _ := http.NewRequest("DELETE", "/api/v1/apikeys/nonexistent-id", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteApiKey_CannotDeleteOthersKey(t *testing.T) {
	db := database.GetTestDB()
	if db == nil {
		t.Skip("Test DB not available")
	}
	router := setupApikeyRouter()
	token1 := registerAndGetToken(t, router, "keyowner1b@example.com", "password123", "Key Owner 1B")
	token2 := registerAndGetToken(t, router, "keyowner2b@example.com", "password123", "Key Owner 2B")

	reqBody := handler.ApiKeyRequest{Name: "Owner 1 Key"}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/apikeys", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token1)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var createResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResp)
	keyID := createResp["id"].(string)

	req, _ = http.NewRequest("DELETE", "/api/v1/apikeys/"+keyID, nil)
	req.Header.Set("Authorization", "Bearer "+token2)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestApiKey_TableName(t *testing.T) {
	key := model.ApiKey{}
	assert.Equal(t, "api_keys", key.TableName())
}
