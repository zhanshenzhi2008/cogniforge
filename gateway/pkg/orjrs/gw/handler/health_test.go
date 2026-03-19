package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/orjrs/gateway/pkg/orjrs/gw/handler"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ==================== Health Check Tests ====================

func TestHealth_Success(t *testing.T) {
	router := gin.New()
	router.GET("/health", handler.Health)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response handler.HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response.Status)
	assert.NotEmpty(t, response.Timestamp)
	assert.Equal(t, handler.AppVersion, response.Version)
}

func TestHealth_ReturnsJSON(t *testing.T) {
	router := gin.New()
	router.GET("/health", handler.Health)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	assert.True(t, strings.Contains(contentType, "application/json"))
}

func TestHealth_ResponseFormat(t *testing.T) {
	router := gin.New()
	router.GET("/health", handler.Health)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify all expected fields exist
	assert.Contains(t, response, "status")
	assert.Contains(t, response, "timestamp")
	assert.Contains(t, response, "version")
}

// ==================== Ready Check Tests ====================

func TestReady_Success(t *testing.T) {
	router := gin.New()
	router.GET("/ready", handler.Ready)

	req, _ := http.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ready", response["status"])
}

// ==================== Live Check Tests ====================

func TestLive_Success(t *testing.T) {
	router := gin.New()
	router.GET("/live", handler.Live)

	req, _ := http.NewRequest("GET", "/live", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "alive", response["status"])
}

// ==================== AppVersion Test ====================

func TestAppVersion_IsDefined(t *testing.T) {
	assert.NotEmpty(t, handler.AppVersion, "AppVersion should be defined")
	assert.Equal(t, "1.0.0", handler.AppVersion, "AppVersion should be 1.0.0")
}
