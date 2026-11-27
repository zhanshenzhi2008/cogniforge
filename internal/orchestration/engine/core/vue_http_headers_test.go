package core

import (
	"encoding/json"
	"testing"
)

func TestParseDefinition_HTTPHeadersFromData(t *testing.T) {
	e := &WorkflowEngine{}
	defJSON := `{
		"nodes":[{"id":"h1","type":"http","data":{"url":"https://example.com","method":"POST","headers":{"X-Trace":"1"}}}],
		"edges":[]
	}`
	def, err := e.ParseDefinition(defJSON)
	if err != nil {
		t.Fatal(err)
	}
	if len(def.Nodes) != 1 {
		t.Fatalf("nodes=%d", len(def.Nodes))
	}
	var cfg map[string]any
	if err := json.Unmarshal(def.Nodes[0].Config, &cfg); err != nil {
		t.Fatal(err)
	}
	headers, ok := cfg["headers"].(map[string]any)
	if !ok || headers["X-Trace"] != "1" {
		t.Fatalf("cfg=%v", cfg)
	}
}
