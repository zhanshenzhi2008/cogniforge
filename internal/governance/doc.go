// Package governance 监控治理层：可观测性、配额、权限、评测与质量门禁。
//
// 子包：
//   - monitor  运行监控 / 指标
//   - quota    用量与限额
//   - rbac     角色权限
//   - eval     评测（Prompt/Agent 质量，属治理而非业务 CRUD）
//
// 依赖规则：可依赖 capability / infra / shared / model；勿依赖具体业务页面编排细节。
package governance
