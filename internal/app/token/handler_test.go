package token

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	infratoken "cogniforge/internal/infra/token"
)

func setupRouter(svc *infratoken.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	auth := r.Group("/api/v1")
	auth.Use(func(c *gin.Context) {
		c.Set("user_id", "user-test")
		c.Next()
	})
	NewHandler(svc).RegisterRoutes(auth)
	return r
}

func TestHandler_Issue_LLM(t *testing.T) {
	svc := infratoken.NewService(nil)
	r := setupRouter(svc)

	req := httptest.NewRequest("GET", "/api/v1/token/llm", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"token"`)
	assert.Contains(t, w.Body.String(), `"expires_in":300`)
	assert.Contains(t, w.Body.String(), `"scope":"llm_call"`)
}

func TestHandler_Issue_NoAuth(t *testing.T) {
	svc := infratoken.NewService(nil)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	NewHandler(svc).RegisterRoutes(r.Group("/api/v1"))

	req := httptest.NewRequest("GET", "/api/v1/token/llm", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_Validate_OK(t *testing.T) {
	svc := infratoken.NewService(nil)
	r := setupRouter(svc)

	issueReq := httptest.NewRequest("GET", "/api/v1/token/llm", nil)
	issueW := httptest.NewRecorder()
	r.ServeHTTP(issueW, issueReq)
	tok := extractToken(t, issueW.Body.String())

	body := `{"token":"` + tok + `"}`
	validateReq := httptest.NewRequest("POST", "/api/v1/token/validate", strings.NewReader(body))
	validateReq.Header.Set("Content-Type", "application/json")
	validateW := httptest.NewRecorder()
	r.ServeHTTP(validateW, validateReq)

	assert.Equal(t, http.StatusOK, validateW.Code)
	assert.Contains(t, validateW.Body.String(), `"valid":true`)
	assert.Contains(t, validateW.Body.String(), `"user_id":"user-test"`)
}

func TestHandler_Validate_Invalid(t *testing.T) {
	svc := infratoken.NewService(nil)
	r := setupRouter(svc)

	body := `{"token":"invalid-token"}`
	req := httptest.NewRequest("POST", "/api/v1/token/validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"invalid token"`)
}

func TestHandler_Validate_MissingToken(t *testing.T) {
	svc := infratoken.NewService(nil)
	r := setupRouter(svc)

	req := httptest.NewRequest("POST", "/api/v1/token/validate", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func extractToken(t *testing.T, body string) string {
	t.Helper()
	start := strings.Index(body, `"token":"`) + len(`"token":"`)
	end := strings.Index(body[start:], `"`)
	return body[start : start+end]
}
