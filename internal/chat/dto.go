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
