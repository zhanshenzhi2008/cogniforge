package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"cogniforge/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestCors_AllowsAllOrigins(t *testing.T) {
	router := gin.New()
	router.Use(middleware.Cors())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCors_AllowsAllMethods(t *testing.T) {
	router := gin.New()
	router.Use(middleware.Cors())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	allowedMethods := w.Header().Get("Access-Control-Allow-Methods")
	assert.Contains(t, allowedMethods, "GET")
	assert.Contains(t, allowedMethods, "POST")
	assert.Contains(t, allowedMethods, "PUT")
	assert.Contains(t, allowedMethods, "DELETE")
	assert.Contains(t, allowedMethods, "OPTIONS")
}

func TestCors_AllowsRequiredHeaders(t *testing.T) {
	router := gin.New()
	router.Use(middleware.Cors())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	allowedHeaders := w.Header().Get("Access-Control-Allow-Headers")
	assert.Contains(t, allowedHeaders, "Origin")
	assert.Contains(t, allowedHeaders, "Content-Type")
	assert.Contains(t, allowedHeaders, "Accept")
	assert.Contains(t, allowedHeaders, "Authorization")
}

func TestCors_ExposesHeaders(t *testing.T) {
	router := gin.New()
	router.Use(middleware.Cors())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, "Content-Length", w.Header().Get("Access-Control-Expose-Headers"))
}

func TestCors_AllowsCredentials(t *testing.T) {
	router := gin.New()
	router.Use(middleware.Cors())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCors_PreflightReturns204(t *testing.T) {
	router := gin.New()
	router.Use(middleware.Cors())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestCors_PreflightIncludesCorsHeaders(t *testing.T) {
	router := gin.New()
	router.Use(middleware.Cors())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Authorization")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
}

func TestCors_PreflightAbortsHandler(t *testing.T) {
	handlerCalled := false

	router := gin.New()
	router.Use(middleware.Cors())
	router.GET("/test", func(c *gin.Context) {
		handlerCalled = true
		c.JSON(http.StatusOK, gin.H{})
	})

	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.False(t, handlerCalled, "Handler should not be called for preflight request")
}

func TestCors_NormalGetRequestIncludesCorsHeaders(t *testing.T) {
	router := gin.New()
	router.Use(middleware.Cors())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "PUT")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "DELETE")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
}

func TestCors_WorksWithAuthMiddleware(t *testing.T) {
	router := gin.New()
	router.Use(middleware.Cors())

	api := router.Group("/api/v1")
	{
		api.GET("/public", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"public": true})
		})
	}

	req, _ := http.NewRequest("OPTIONS", "/api/v1/public", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
}
