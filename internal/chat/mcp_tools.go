package chat

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"

	"cogniforge/internal/mcp"
)

// SetMCPRegistry 注入 MCP Registry（Playground / 共享工具执行）
func (h *ChatHandler) SetMCPRegistry(reg *mcp.Registry) {
	h.mcpRegistry = reg
}

// InjectMCPTools 按 server ID 列表把工具 schema 写入 req.Tools（已有 Tools 时追加）
func InjectMCPTools(req *ChatRequest, reg *mcp.Registry, serverIDs []string) {
	if req == nil || reg == nil || len(serverIDs) == 0 {
		return
	}
	for _, toolDef := range reg.GetToolsForAgent(serverIDs) {
		req.Tools = append(req.Tools, ToolCallSpec{
			Type: "function",
			Function: ToolFunctionSpec{
				Name:        toolDef.Name,
				Description: toolDef.Description,
				Parameters:  toolDef.InputSchema,
			},
		})
	}
}

// NewMCPToolExecutor 供 ChatWithTools 使用
func NewMCPToolExecutor(reg *mcp.Registry) ToolExecutor {
	return &mcpToolExecutor{registry: reg, executor: mcp.NewBuiltInExecutor()}
}

type mcpToolExecutor struct {
	registry *mcp.Registry
	executor *mcp.BuiltInExecutor
}

func (e *mcpToolExecutor) Execute(call ToolCall) (string, error) {
	if e.registry == nil {
		return "", fmt.Errorf("mcp registry not configured")
	}
	toolDef := e.registry.GetBuiltInTool(call.Function.Name)
	if toolDef == nil {
		return "", fmt.Errorf("unknown tool: %s", call.Function.Name)
	}
	var args map[string]interface{}
	if call.Function.Arguments != "" {
		_ = json.Unmarshal([]byte(call.Function.Arguments), &args)
	}
	if toolDef.ServerType == "built-in" {
		result, err := e.executor.Execute(toolDef.Name, args)
		if err != nil {
			return "", err
		}
		b, _ := json.Marshal(result)
		return string(b), nil
	}
	srv := e.registry.GetServer(toolDef.ServerID)
	if srv == nil || srv.URL == "" {
		return "", fmt.Errorf("MCP server not available: %s", toolDef.ServerID)
	}
	result, err := e.executor.CallHTTPClient(srv.URL, srv.AuthHeader, toolDef.Name, args)
	if err != nil {
		return "", err
	}
	b, _ := json.Marshal(result)
	return string(b), nil
}

// WriteResponseSSE 把非流式最终回复写成 OpenAI 兼容 SSE
func WriteResponseSSE(c *gin.Context, resp *ChatResponse) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	content := ""
	if len(resp.Choices) > 0 {
		if s, ok := resp.Choices[0].Message.Content.(string); ok {
			content = s
		}
	}
	evt := map[string]any{
		"id":      resp.ID,
		"object":  "chat.completion.chunk",
		"created": resp.Created,
		"model":   resp.Model,
		"choices": []map[string]any{
			{
				"index": 0,
				"delta": map[string]any{
					"role":    "assistant",
					"content": content,
				},
				"finish_reason": "stop",
			},
		},
	}
	b, _ := json.Marshal(evt)
	_, _ = c.Writer.Write([]byte("data: " + string(b) + "\n\n"))
	_, _ = c.Writer.Write([]byte("data: [DONE]\n\n"))
	c.Writer.Flush()
}
