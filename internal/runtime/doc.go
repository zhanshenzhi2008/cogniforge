// Package runtime 运行时层：Agent / Chat 执行循环（调模型、工具、读写 Memory）。
//
// 子包：
//   - chat   Playground/对话流式执行、RAG 注入、MCP 工具
//   - agent  Agent CRUD 与对话执行入口
//
// 依赖规则：可依赖 capability / governance(quota) / infra / shared；
// Memory 调用方向必须是 runtime → capability/memory。
package runtime
