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

type ResumeExecutionRequest struct {
	Seq *int `json:"seq"` // 可选：指定 checkpoint seq；默认最新
}

type ApproveExecutionRequest struct {
	Comment string `json:"comment,omitempty"`
	Seq     *int   `json:"seq,omitempty"`
}

type RejectExecutionRequest struct {
	Reason string `json:"reason,omitempty"`
}

type CallbackExecutionRequest struct {
	Action  string         `json:"action" binding:"required"` // approve | reject
	Comment string         `json:"comment,omitempty"`
	Reason  string         `json:"reason,omitempty"`
	Payload map[string]any `json:"payload,omitempty"`
	Seq     *int           `json:"seq,omitempty"`
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

// ExecutionDetail 执行详情（含节点运行记录，阶段十八 W1）
type ExecutionDetail struct {
	ID                 string      `json:"id"`
	WorkflowID         string      `json:"workflow_id"`
	UserID             string      `json:"user_id"`
	Status             string      `json:"status"`
	Input              any         `json:"input,omitempty"`
	Output             string      `json:"output,omitempty"`
	Error              string      `json:"error,omitempty"`
	CurrentNode        string      `json:"current_node,omitempty"`
	TraceID            string      `json:"trace_id,omitempty"`
	StartedAt          any         `json:"started_at,omitempty"`
	CompletedAt        any         `json:"completed_at,omitempty"`
	CreatedAt          any         `json:"created_at,omitempty"`
	UpdatedAt          any         `json:"updated_at,omitempty"`
	WaitReason          string `json:"wait_reason,omitempty"`
	WaitUntil           any    `json:"wait_until,omitempty"`
	NodeRuns            any    `json:"node_runs"`
	LatestCheckpointSeq int    `json:"latest_checkpoint_seq,omitempty"`
}
