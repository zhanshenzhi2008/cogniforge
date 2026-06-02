package workflow

// ============ 请求结构 ============

type CreateWorkflowRequest struct {
	Name        string         `json:"name" binding:"required"`
	Description string         `json:"description"`
	Definition  map[string]any `json:"definition"`
}

type UpdateWorkflowRequest struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Status      string         `json:"status"`
	Definition  map[string]any `json:"definition"`
}

type ExecuteWorkflowRequest struct {
	Input map[string]any `json:"input"`
}

type DebugWorkflowRequest struct {
	Input           map[string]any `json:"input"`
	NodeBreakpoints []string       `json:"node_breakpoints"`
}

// ============ 响应结构 ============

type ExecuteResponse struct {
	ExecutionID string `json:"execution_id"`
	TraceID     string `json:"trace_id,omitempty"` // 可选的 traceId，用于追踪
	Status      string `json:"status"`
}

type WorkflowExecutionResponse struct {
	ID         string `json:"id"`
	WorkflowID string `json:"workflow_id"`
	Status     string `json:"status"`
}
