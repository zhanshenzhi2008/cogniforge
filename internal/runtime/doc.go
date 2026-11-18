// Package runtime 运行时层：Chat / Agent 执行循环。
//
// 子包：
//   - chat   Playground/对话流式执行、拼 context（RAG + Memory 窗口）、MCP 工具
//   - agent  Agent 对话执行（CRUD/HTTP 暂仍同包，后续可拆到 app）
//
// 本层负责「怎么跑」：组装本轮 context、调模型、跑工具、流式输出。
// Provider / Memory 是能力积木，在 capability；调用方向必须是
// runtime → capability/{provider,memory}，禁止反向。
//
// 依赖：capability / governance(quota) / infra / shared。
package runtime
