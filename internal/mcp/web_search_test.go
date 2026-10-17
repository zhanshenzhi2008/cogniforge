package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchKeyless_Non200ReturnsErrorNotNilNil(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"no key"}`))
	}))
	t.Cleanup(srv.Close)

	client := &TavilyClient{keyless: true, baseURL: srv.URL}
	res, err := client.Search(context.Background(), "北京天气", 3)
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "401")
}

func TestWebSearch_NilResultDoesNotPanic(t *testing.T) {
	e := &BuiltInExecutor{tavily: &TavilyClient{keyless: true, baseURL: "http://127.0.0.1:1"}}
	out, err := e.webSearch(map[string]interface{}{"query": "北京天气"})
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, "error", out["status"])
}
