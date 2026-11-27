package nodes

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cogniforge/internal/orchestration/engine/core"
)

func TestHTTPNodeExecutor_ResolvesTemplates(t *testing.T) {
	var gotMethod, gotURL, gotAuth, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotURL = r.URL.String()
		gotAuth = r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	ex := &HTTPNodeExecutor{client: srv.Client()}
	ctx := core.NewExecutionContext("wf", "exec", map[string]any{"token": "abc"}).WithUserID("u1")
	ctx.SetVariable("path", "/v1/items")
	ctx.SetVariable("q", "hello")

	cfg, _ := json.Marshal(map[string]any{
		"url":    srv.URL + "{{path}}",
		"method": "POST",
		"headers": map[string]string{
			"Authorization": "Bearer {{token}}",
		},
		"body": map[string]any{
			"query": "{{q}}",
		},
		"timeout": 5,
	})
	out, err := ex.Execute(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != "POST" || gotURL != "/v1/items" {
		t.Fatalf("req method=%s url=%s", gotMethod, gotURL)
	}
	if gotAuth != "Bearer abc" {
		t.Fatalf("auth=%q", gotAuth)
	}
	if gotBody != `{"query":"hello"}` {
		t.Fatalf("body=%q", gotBody)
	}
	m, ok := out.(map[string]any)
	if !ok || m["status_code"] != 201 {
		t.Fatalf("out=%v", out)
	}
}

func TestHTTPNodeExecutor_FailOnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(502)
		_, _ = w.Write([]byte(`upstream down`))
	}))
	defer srv.Close()

	ex := &HTTPNodeExecutor{client: srv.Client()}
	ctx := core.NewExecutionContext("wf", "exec", nil)

	cfgOK, _ := json.Marshal(map[string]any{
		"url":    srv.URL,
		"method": "GET",
	})
	out, err := ex.Execute(ctx, cfgOK)
	if err != nil {
		t.Fatalf("default should not fail on 502: %v", err)
	}
	if m, _ := out.(map[string]any); m["status_code"] != 502 {
		t.Fatalf("out=%v", out)
	}

	cfgFail, _ := json.Marshal(map[string]any{
		"url":           srv.URL,
		"method":        "GET",
		"fail_on_error": true,
	})
	_, err = ex.Execute(ctx, cfgFail)
	if err == nil {
		t.Fatal("expected error when fail_on_error")
	}
	if !strings.Contains(err.Error(), "502") {
		t.Fatalf("err=%v", err)
	}
}

func TestResolveAnyTemplates(t *testing.T) {
	ctx := core.NewExecutionContext("wf", "exec", map[string]any{"a": 1})
	ctx.SetVariable("name", "Ada")
	got := resolveAnyTemplates(map[string]any{
		"hi": "Hello {{name}}",
		"n":  []any{"{{a}}", 2},
	}, ctx).(map[string]any)
	if got["hi"] != "Hello Ada" {
		t.Fatalf("hi=%v", got["hi"])
	}
	arr := got["n"].([]any)
	if arr[0] != "1" {
		t.Fatalf("arr=%v", arr)
	}
}
