package memory

// WindowMessage 窗口裁剪用的消息结构（与 WindowMessage 字段对齐，避免循环依赖）。
type WindowMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

// Config 滑动窗口配置。
type Config struct {
	// MemoryTurns 窗口轮数（1 user + 1 assistant = 1 轮）。范围 1–40。
	// 默认 10。
	MemoryTurns int
	// MaxInputTokens Token 上限（0 = 不限）。
	// 各模型默认值：4k=3800、16k=15000、32k=30000、128k=120000。
	MaxInputTokens int
}

// DefaultConfig 默认 Playground 配置。
func DefaultConfig() Config {
	return Config{
		MemoryTurns:    10,
		MaxInputTokens: 0,
	}
}

// ApplyWindow 对 incoming messages 应用滑动窗口裁剪。
// turns=0 使用默认值 10；turns<0 禁用裁剪。
func ApplyWindow(msgs []WindowMessage, turns, maxTokens int) []WindowMessage {
	if turns < 0 {
		return msgs
	}
	cfg := Config{
		MemoryTurns:    turns,
		MaxInputTokens: maxTokens,
	}
	if cfg.MemoryTurns == 0 {
		cfg = DefaultConfig()
		cfg.MaxInputTokens = maxTokens
	}
	return cfg.Apply(msgs)
}

// Apply 服务端裁剪：输入前端传来的 messages + 本会话已落库历史，返回裁剪后的 messages。
//
// 规则（L1+L2）：
//  1. 去掉空内容消息
//  2. 从尾部取最多 memoryTurns*2 条（优先成对；落单的 user 保留）
//  3. 估算 Token；若超 MaxInputTokens，继续从最旧一条删，直到进入预算
//  4. 本轮 user 消息永不删除（最后一对里的 user）
//  5. 返回裁剪后的 messages
func (c Config) Apply(incoming []WindowMessage) []WindowMessage {
	if c.MemoryTurns <= 0 {
		c.MemoryTurns = 10
	}
	// 1. 过滤空内容
	filtered := filterEmpty(incoming)
	if len(filtered) == 0 {
		return filtered
	}

	// 2. 尾部取 memoryTurns 轮（成对优先）
	window := trimToTurns(filtered, c.MemoryTurns)
	if len(window) == 0 {
		return window
	}

	// 3. Token 预算裁剪（不删最后一对里的 user）
	trimmed := trimByToken(window, c.MaxInputTokens)
	return trimmed
}

// filterEmpty 去掉空内容消息
func filterEmpty(msgs []WindowMessage) []WindowMessage {
	out := make([]WindowMessage, 0, len(msgs))
	for _, m := range msgs {
		if isEmptyContent(m.Content) {
			continue
		}
		out = append(out, m)
	}
	return out
}

// isEmptyContent 判断消息内容是否为空
func isEmptyContent(content any) bool {
	switch v := content.(type) {
	case string:
		return v == ""
	case []any:
		if len(v) == 0 {
			return true
		}
		for _, part := range v {
			if m, ok := part.(map[string]any); ok {
				if t, ok := m["text"].(string); ok && t != "" {
					return false
				}
				if _, ok := m["image_url"]; ok {
					return false
				}
			}
		}
		return true
	default:
		return true
	}
}

// trimToTurns 从尾部取最多 turns*2 条，优先成对（user+assistant），落单的 user 保留。
func trimToTurns(msgs []WindowMessage, turns int) []WindowMessage {
	max := turns * 2
	if len(msgs) <= max {
		return msgs
	}

	// 从尾部往前数，找最近的 turns 轮
	// 一轮 = 1 user + 1 assistant；从尾部的最后一条开始往前数
	cnt := 0
	start := len(msgs)
	for i := len(msgs) - 1; i >= 0 && cnt < max; i-- {
		cnt++
		start = i
		// 如果到达 user 角色，回退看有没有配对的 assistant
		// 逻辑：优先保留完整的 user+assistant 对
	}

	// 调整：确保最后一条是 user（如果被截断到 assistant 前面，补上 user）
	if start > 0 && msgs[start].Role == "assistant" {
		// 找上一个 user
		for i := start - 1; i >= 0; i-- {
			if msgs[i].Role == "user" {
				start = i
				break
			}
		}
	}

	result := msgs[start:]
	// 最后一条如果是 assistant（只有 assistant），找前一个 user 补上
	if len(result) > 0 && result[0].Role == "assistant" {
		for i := start - 1; i >= 0; i-- {
			if msgs[i].Role == "user" {
				result = append([]WindowMessage{msgs[i]}, result...)
				break
			}
		}
	}

	return result
}

// trimByToken 按 Token 预算裁剪，最旧的消息先删，本轮 user 消息（最后一条 role=user）不删。
func trimByToken(msgs []WindowMessage, maxTokens int) []WindowMessage {
	if maxTokens <= 0 {
		return msgs
	}

	// 找最后一条 user 消息的索引（内容永不删除）
	protectedIdx := -1
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role == "user" {
			protectedIdx = i
			break
		}
	}

	// 从最旧（头部）开始逐条删除，直到 token 在预算内
	// 删除时跳过 protectedIdx
	head := 0
	result := msgs
	for tokenCount(result) > maxTokens && head < len(result) {
		if head == protectedIdx {
			head++
			continue
		}
		// 删除 head 这条
		result = append(result[:head], result[head+1:]...)
		if protectedIdx > head {
			protectedIdx--
		}
		// head 不变（下条旧消息现在在 head 位置）
	}

	return result
}

// tokenCount 估算 messages 总 token 数（内联实现，避免循环依赖）。
func tokenCount(msgs []WindowMessage) int {
	n := 0
	for _, m := range msgs {
		switch v := m.Content.(type) {
		case string:
			n += textTokens(v)
		case []any:
			for _, part := range v {
				if m, ok := part.(map[string]any); ok {
					switch m["type"] {
					case "text":
						if t, ok := m["text"].(string); ok {
							n += textTokens(t)
						}
					case "image_url":
						n += 300
					}
				}
			}
		}
	}
	return n
}

// textTokens 按中文/ASCII 估算 token 数（与 chat/sse.go 口径一致）。
func textTokens(s string) int {
	if s == "" {
		return 0
	}
	ascii := 0
	n := 0
	for _, r := range s {
		n++
		if r < 128 {
			ascii++
		}
	}
	cjk := n - ascii
	est := ascii/4 + cjk
	if est < 1 {
		return 1
	}
	return est
}
