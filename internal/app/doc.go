// Package app 业务应用层：对外 HTTP 用例与路由组装。
//
// 子包：
//   - auth / user / prompt
//   - chat / agent（执行核在 runtime）
//   - provider / memory / mcp / skill / skillimport / knowledge（能力实现在 capability）
//   - quota / eval / monitor / rbac（策略实现在 governance）
//   - token（签发实现在 infra/token）
//   - router
package app
