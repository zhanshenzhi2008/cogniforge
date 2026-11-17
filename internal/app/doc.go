// Package app 业务应用层：对外用例与路由组装。
//
// 子包：
//   - auth / user / prompt  账号、会话、提示词工程 HTTP 用例
//   - router               路由聚合（组装各层 handler）
//
// 说明：chat/agent 的 HTTP 仍与执行逻辑同在 runtime 包内（历史结构）；
// 后续可再拆 handler 到 app，执行核留在 runtime。
package app
