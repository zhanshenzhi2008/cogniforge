package model

// =============================================================================
// 统一 API 响应结构
// =============================================================================

type ApiResponse struct {
	Code    int         `json:"code"`             // 业务状态码
	Message string      `json:"message"`          // 状态描述
	TraceID string      `json:"trace_id"`         // 请求追踪 ID
	Data    interface{} `json:"data,omitempty"`  // 业务数据
	Error   string      `json:"error,omitempty"`  // 错误信息（失败时）
}
