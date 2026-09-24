package chat

// ============ 请求结构 ============

// ChatMessage OpenAI 兼容。Content 可为 string，或
// [{type:text|image_url,...}] 多模态数组（透传上游）。
type ChatMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages" binding:"required"`
	Stream   bool          `json:"stream"`

	// 阶段十五 15.2：MCP 工具列表（由 MCP Registry 注入）
	Tools []ToolCallSpec `json:"tools,omitempty"`

	// 滑动窗口参数（阶段十四 14.1）
	// MemoryTurns: 窗口轮数，默认 10，范围 1–40。0=禁用服务端裁剪。
	// MaxInputTokens: Token 预算上限，0=不限。
	MemoryTurns    int `json:"memory_turns,omitempty"`
	MaxInputTokens int `json:"max_input_tokens,omitempty"`

	// 阶段十四 14.3：对话 ID，用于加载历史和保存消息
	ConversationID string `json:"conversation_id,omitempty"`

	// 阶段十四 14.4：是否使用长期记忆（前端可关）
	UseLongTerm bool   `json:"use_long_term,omitempty"`
	// AgentID：用于长期记忆 scope 过滤（agent专属记忆 vs 全局记忆）
	AgentID     string `json:"agent_id,omitempty"`

	Temperature      *float64 `json:"temperature,omitempty"`
	MaxTokens        *int     `json:"max_tokens,omitempty"`
	TopP             *float64 `json:"top_p,omitempty"`
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64 `json:"presence_penalty,omitempty"`
	Stop             []string `json:"stop,omitempty"`
}

// ============ 响应结构 ============

type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
	// 阶段十五 15.2：OpenAI tool_calls 字段（不在 Content 里）
	ToolCalls []map[string]any `json:"tool_calls,omitempty"`
}

type ChatUsage struct {
	PromptTokens     int  `json:"prompt_tokens"`
	CompletionTokens int  `json:"completion_tokens"`
	TotalTokens      int  `json:"total_tokens"`
	Estimated        bool `json:"-"`
}

type ChatResponse struct {
	ID                string       `json:"id"`
	Object            string       `json:"object"`
	Created           int64        `json:"created"`
	Model             string       `json:"model"`
	Choices           []ChatChoice `json:"choices"`
	Usage             ChatUsage    `json:"usage"`
	SystemFingerprint string       `json:"system_fingerprint,omitempty"`
}

type SSEEvent struct {
	ID                string      `json:"id"`
	Object            string      `json:"object"`
	Created           int64       `json:"created"`
	Model             string      `json:"model"`
	Choices           []SSEChoice `json:"choices"`
	SystemFingerprint string      `json:"system_fingerprint,omitempty"`
}

type SSEChoice struct {
	Index        int            `json:"index"`
	Delta        map[string]any `json:"delta"`
	FinishReason string         `json:"finish_reason,omitempty"`
}

type ModelInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ListModelsResponse struct {
	Models []ModelInfo `json:"models"`
}

// ToolCallSpec 工具调用规格（阶段十五 15.2）
type ToolCallSpec struct {
	Type        string                 `json:"type"`                 // 固定 "function"
	Function    ToolFunctionSpec       `json:"function"`
}

type ToolFunctionSpec struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"` // JSON Schema
}

// ToolCall 工具调用请求（上游 LLM 返回）
type ToolCall struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Function ToolCallFunction       `json:"function"`
}

type ToolCallFunction struct {
	Name      string                 `json:"name"`
	Arguments string                 `json:"arguments"` // JSON string
}

// ToolResult 工具执行结果
type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Role       string `json:"role"` // "tool"
	Content    string `json:"content"`
	IsError    bool   `json:"is_error,omitempty"`
}

// ToolExecutor 工具执行器接口（阶段十五 15.2）
type ToolExecutor interface {
	Execute(call ToolCall) (result string, err error)
}
type ToolCallResponse struct {
	ID       string `json:"id"`
	Object   string `json:"object"`
	Created  int64  `json:"created"`
	Model    string `json:"model"`
	Choices  []struct {
		Index        int            `json:"index"`
		Message      ChatMessage    `json:"message"`
		FinishReason string         `json:"finish_reason"`
	} `json:"choices"`
	Usage ChatUsage `json:"usage"`
}

// EmbeddingsRequest OpenAI 兼容：input 可以是字符串或字符串数组
type EmbeddingsRequest struct {
	Model string `json:"model"`
	Input any    `json:"input" binding:"required"`
}

type EmbeddingItem struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

type EmbeddingsUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

type EmbeddingsResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingItem `json:"data"`
	Model  string          `json:"model"`
	Usage  EmbeddingsUsage `json:"usage"`
}
