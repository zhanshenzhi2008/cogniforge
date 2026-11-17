// Package orchestration 编排层：工作流图调度与节点执行（与 Agent runtime 并列）。
//
// 子包：
//   - engine    工作流引擎与节点
//   - workflow  工作流 CRUD / 执行 API
//
// 依赖规则：可依赖 capability（如 provider/mcp）与 infra；Agent 单轮循环归 runtime，不放本层。
package orchestration
