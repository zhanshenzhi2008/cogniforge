package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/orjrs/gateway/pkg/orjrs/gw/middleware"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ==================== CORS Middleware Tests ====================

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

// ==================== CORS Preflight Tests ====================

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

	// Handler should NOT be called for OPTIONS preflight
	assert.False(t, handlerCalled, "Handler should not be called for preflight request")
}

// ==================== CORS with Different Origins ====================

func TestCors_OriginWildcard(t *testing.T) {
	testCases := []struct {
		name  string
		origin string
	}{
		{"localhost", "http://localhost:3000"},
		{"production", "https://cogniforge.com"},
		{"empty origin", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.Use(middleware.Cors())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{})
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		})
	}
}

// ==================== CORS Integration with API Routes ====================

func TestCors_WorksWithAuthMiddleware(t *testing.T) {
	router := gin.New()
	router.Use(middleware.Cors())

	api := router.Group("/api/v1")
	{
		api.GET("/public", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"public": true})
		})
	}

	// Test preflight for API route
	req, _ := http.NewRequest("OPTIONS", "/api/v1/public", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
}
