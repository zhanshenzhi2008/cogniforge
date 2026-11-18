package chat

import "errors"

// ErrNoActiveProvider 没有可用的默认模型或 API Key。
// 对话不应再返回 mock 文本，应提示用户去「模型」页配置。
var ErrNoActiveProvider = errors.New("请先到「模型」页填写 API Key，并设置一个默认模型后再对话")
