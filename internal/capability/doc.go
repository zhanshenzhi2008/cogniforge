// Package capability 能力层：可被 runtime / 编排 / 评测复用的积木。
//
// 子包：provider / memory / mcp / skill / skillimport / knowledge / webfetch。
// HTTP handler 已迁至 app 对应子包；本层只留 Service / Registry / 领域逻辑。
//
// 归属约定（已锁定）：
//   - provider / memory 留本层（不是 runtime）
//   - context 由 runtime 组装，不单独立包
//
// 依赖：infra / shared / model。禁止依赖 app、router、gateway、runtime。
package capability
