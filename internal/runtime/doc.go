// Package runtime 运行时层：Chat / Agent 执行核。
//
// 子包：
//   - chat   ChatService、会话存储、RAG/MCP 工具注入、流式
//   - agent  AgentService（CRUD 领域逻辑；HTTP 在 app/agent）
//
// HTTP handler 已迁至 app/{chat,agent}。
// 调用方向：runtime → capability/{provider,memory,…}，禁止反向。
package runtime
