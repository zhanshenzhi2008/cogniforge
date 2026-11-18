// Package app 业务应用层：对外 HTTP 用例与路由组装。
//
// 子包：
//   - auth / user / prompt
//   - chat / agent（执行核在 runtime）
//   - provider / memory / mcp / skill / skillimport / knowledge（能力 CRUD HTTP；实现在 capability）
//   - router
package app
