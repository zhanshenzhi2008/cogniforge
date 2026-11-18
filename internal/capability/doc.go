// Package capability 能力层：可被 runtime / 编排 / 评测复用的积木。
//
// 子包：provider / memory / mcp / skill / skillimport / knowledge / webfetch。
//
// 归属约定（已锁定）：
//   - provider：模型供应商、API Key、活跃模型（不是执行循环）
//   - memory：记忆存取与窗口裁剪（不是「拼本轮 context」；拼 context 在 runtime）
//   - 不设独立 context 包：context 由 runtime 当场组装
//
// 依赖：infra / shared / model。
// 禁止依赖 app、router、gateway、runtime。
package capability
