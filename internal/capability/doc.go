// Package capability 能力层：可被 runtime / 编排 / 评测复用的积木。
//
// 子包：provider / memory / mcp / skill / skillimport / knowledge / webfetch / prompt。
// HTTP handler 在 app；本层只留 Service / 领域逻辑。
//
// 禁止依赖 app、router、gateway、runtime。
package capability
