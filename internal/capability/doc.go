// Package capability 能力层：模型供应商、Memory、MCP/Skill、知识库核心、网页抓取等可复用能力。
//
// 子包：provider / memory / mcp / skill / skillimport / knowledge / webfetch。
// 依赖规则：可依赖 infra / shared / model；不得依赖 app（HTTP 用例）与具体页面编排入口。
package capability
