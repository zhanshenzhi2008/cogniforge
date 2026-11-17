package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStreamableHTTPClient_ListAndCall(t *testing.T) {
	var gotSession string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.Header.Get("Accept"), "application/json")

		var req jsonRPCRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))

		switch req.Method {
		case "initialize":
			w.Header().Set("Mcp-Session-Id", "sess-test-1")
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      req.ID,
				"result": map[string]interface{}{
					"protocolVersion": "2025-06-18",
					"capabilities":    map[string]interface{}{"tools": map[string]interface{}{}},
					"serverInfo":      map[string]string{"name": "mock", "version": "0.1"},
				},
			})
		case "notifications/initialized":
			w.WriteHeader(http.StatusAccepted)
		case "tools/list":
			gotSession = r.Header.Get("Mcp-Session-Id")
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      req.ID,
				"result": map[string]interface{}{
					"tools": []map[string]interface{}{
						{
							"name":        "feishu_docs_search",
							"description": "搜索飞书文档",
							"inputSchema": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"query": map[string]interface{}{"type": "string"},
								},
								"required": []string{"query"},
							},
						},
					},
				},
			})
		case "tools/call":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      req.ID,
				"result": map[string]interface{}{
					"content": []map[string]interface{}{
						{"type": "text", "text": "found: hello"},
					},
					"isError": false,
				},
			})
		default:
			t.Fatalf("unexpected method %s", req.Method)
		}
	}))
	defer srv.Close()

	client := NewStreamableHTTPClient(srv.URL, "sk-test-token")
	ctx := context.Background()

	tools, err := client.ListTools(ctx)
	require.NoError(t, err)
	require.Len(t, tools, 1)
	assert.Equal(t, "feishu_docs_search", tools[0].Name)
	assert.Equal(t, "sess-test-1", gotSession)
	assert.Equal(t, "Bearer sk-test-token", client.authHeader)

	out, err := client.CallTool(ctx, "feishu_docs_search", map[string]interface{}{"query": "hello"})
	require.NoError(t, err)
	assert.Equal(t, "found: hello", out["text"])
	assert.Equal(t, "ok", out["status"])
}

func TestStreamableHTTPClient_SSEResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req jsonRPCRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		if req.Method == "initialize" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0", "id": req.ID,
				"result": map[string]interface{}{"protocolVersion": "2025-06-18"},
			})
			return
		}
		if req.Method == "notifications/initialized" {
			w.WriteHeader(http.StatusAccepted)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		payload, _ := json.Marshal(map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      req.ID,
			"result": map[string]interface{}{
				"tools": []map[string]interface{}{
					{"name": "ping", "description": "pong", "inputSchema": map[string]interface{}{"type": "object"}},
				},
			},
		})
		_, _ = w.Write([]byte("event: message\ndata: " + string(payload) + "\n\n"))
	}))
	defer srv.Close()

	client := NewStreamableHTTPClient(srv.URL, "")
	tools, err := client.ListTools(context.Background())
	require.NoError(t, err)
	require.Len(t, tools, 1)
	assert.Equal(t, "ping", tools[0].Name)
}

func TestCallHTTPClient_FallsBackToLegacyGateway(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.URL.Path == "/tools/my_tool/call" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok", "via": "legacy"})
			return
		}
		// Streamable initialize 故意失败
		http.Error(w, "not mcp", http.StatusNotFound)
	}))
	defer srv.Close()

	ex := NewBuiltInExecutor()
	out, err := ex.CallHTTPClient(srv.URL, "", "my_tool", map[string]interface{}{"x": 1})
	require.NoError(t, err)
	assert.Equal(t, "legacy", out["via"])
	assert.GreaterOrEqual(t, hits.Load(), int32(1))
}

func TestNormalizeAuthHeader(t *testing.T) {
	assert.Equal(t, "Bearer abc", normalizeAuthHeader("abc"))
	assert.Equal(t, "Bearer xyz", normalizeAuthHeader("Bearer xyz"))
	assert.Equal(t, "Basic dXNlcjpwYXNz", normalizeAuthHeader("Basic dXNlcjpwYXNz"))
}
