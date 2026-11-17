// Package runtime 运行时层：Agent/Chat 执行循环（调模型、工具、读写 Memory）。
//
// 依赖规则：可依赖 capability / infra / shared；编排（workflow engine）与 runtime 并列，勿互相污染。
// Memory 概念上由 runtime 使用；物理包可仍独立，但调用方向必须是 runtime → memory。
package runtime
