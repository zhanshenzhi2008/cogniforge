package quota

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cogniforge/internal/response"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHandler_Me_NewUserLimited(t *testing.T) {
	db := testDB(t)
	seedUser(t, db, "u-http", "user")
	svc := New(db, NewMemoryStore())
	svc.EnsureDefaultPolicy()
	h := NewHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "u-http")
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/quota/me", nil)
	h.Me(c)

	require.Equal(t, http.StatusOK, w.Code)
	var body response.ApiResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, response.CodeSuccess, body.Code)

	raw, err := json.Marshal(body.Data)
	require.NoError(t, err)
	var snap Snapshot
	require.NoError(t, json.Unmarshal(raw, &snap))
	assert.False(t, snap.Unlimited)
	assert.Equal(t, DefaultDailyRequests, snap.Day.RequestsLimit)
}

func TestHandler_Me_AdminUnlimited(t *testing.T) {
	db := testDB(t)
	seedUser(t, db, "admin-http", "admin")
	svc := New(db, NewMemoryStore())
	svc.EnsureDefaultPolicy()
	h := NewHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "admin-http")
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/quota/me", nil)
	h.Me(c)

	require.Equal(t, http.StatusOK, w.Code)
	var body response.ApiResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	raw, err := json.Marshal(body.Data)
	require.NoError(t, err)
	var snap Snapshot
	require.NoError(t, json.Unmarshal(raw, &snap))
	assert.True(t, snap.Unlimited)
}
