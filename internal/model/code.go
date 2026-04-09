package model

// =============================================================================
// 业务状态码定义
// =============================================================================

// 2xxx: 成功
const (
	CodeSuccess  = 2000 // 成功
	CodeCreated  = 2001 // 创建成功
	CodeUpdated  = 2002 // 更新成功
	CodeDeleted  = 2003 // 删除成功
	CodeAccepted = 2004 // 请求已接受（异步）
)

// 4xxx: 系统运行异常（数据库、网络、服务不可用等）
const (
	CodeSystemError        = 4001 // 系统内部错误
	CodeDatabaseError      = 4002 // 数据库错误
	CodeCacheError         = 4003 // 缓存错误
	CodeNetworkError       = 4004 // 网络错误
	CodeServiceUnavailable = 4005 // 服务不可用
	CodeAIProviderError    = 4006 // AI 服务商错误
	CodeAIRequestTimeout   = 4007 // AI 请求超时
	CodeAIQuotaExhausted   = 4008 // AI 配额已用尽
	CodeModelNotSupport    = 4009 // 不支持的模型
)

// 5xxx: 业务校验异常（参数、数据、权限等）
const (
	CodeParamInvalid        = 5001 // 参数无效
	CodeResourceNotFound    = 5002 // 资源不存在
	CodeResourceConflict    = 5003 // 资源冲突（如重复创建）
	CodeResourceDeleted     = 5004 // 资源已删除
	CodeUnauthorized        = 5005 // 未认证（需登录）
	CodeForbidden           = 5006 // 无权限访问
	CodeTokenInvalid        = 5007 // Token 无效
	CodeTokenExpired        = 5008 // Token 已过期
	CodeEmailExists         = 5009 // 邮箱已被注册
	CodeEmailNotExists      = 5010 // 邮箱不存在
	CodePasswordIncorrect   = 5011 // 密码错误
	CodeUsernameExists      = 5012 // 用户名已被使用
	CodeVerifyCodeInvalid   = 5013 // 验证码无效或已过期
	CodeRateLimitExceeded   = 5014 // 请求频率超限
	CodeRequestTooLarge     = 5015 // 请求数据过大
)

// =============================================================================
// Code 到 Message 的映射
// =============================================================================

var codeMessages = map[int]string{
	// 2xxx: 成功
	CodeSuccess:  "成功",
	CodeCreated:  "创建成功",
	CodeUpdated:  "更新成功",
	CodeDeleted:  "删除成功",
	CodeAccepted: "请求已接受",

	// 4xxx: 系统运行异常
	CodeSystemError:        "系统内部错误",
	CodeDatabaseError:      "数据库错误",
	CodeCacheError:         "缓存错误",
	CodeNetworkError:       "网络错误",
	CodeServiceUnavailable: "服务暂时不可用",
	CodeAIProviderError:    "AI 服务暂时不可用",
	CodeAIRequestTimeout:   "AI 请求超时",
	CodeAIQuotaExhausted:   "AI 配额已用尽",
	CodeModelNotSupport:    "不支持的模型",

	// 5xxx: 业务校验异常
	CodeParamInvalid:       "参数无效",
	CodeResourceNotFound:   "资源不存在",
	CodeResourceConflict:   "资源冲突",
	CodeResourceDeleted:    "资源已删除",
	CodeUnauthorized:       "请先登录",
	CodeForbidden:          "无权限访问",
	CodeTokenInvalid:       "Token 无效",
	CodeTokenExpired:        "Token 已过期",
	CodeEmailExists:        "该邮箱已被注册",
	CodeEmailNotExists:     "邮箱不存在",
	CodePasswordIncorrect:  "密码错误",
	CodeUsernameExists:     "用户名已被使用",
	CodeVerifyCodeInvalid:  "验证码无效或已过期",
	CodeRateLimitExceeded:  "请求频率超限，请稍后重试",
	CodeRequestTooLarge:     "请求数据过大",
}

// GetMessage 根据 code 获取默认消息
func GetMessage(code int) string {
	if msg, ok := codeMessages[code]; ok {
		return msg
	}
	return "未知错误"
}

// IsSuccess 判断是否为成功响应
func IsSuccess(code int) bool {
	return code >= 2000 && code < 3000
}

// IsBizError 判断是否为业务校验异常
func IsBizError(code int) bool {
	return code >= 5000 && code < 6000
}

// IsSysError 判断是否为系统运行异常
func IsSysError(code int) bool {
	return code >= 4000 && code < 5000
}
