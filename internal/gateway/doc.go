// Package gateway HTTP 边缘网关层（单体内部，不是独立微服务）。
//
// 放这里：
//   - CORS / Recovery / 请求日志
//   - JWT 鉴权中间件（AuthRequired / AuthOptional）
//   - RBAC 门禁中间件（RequireAdmin / RequirePermission）
//   - 限流中间件（调用 governance/quota，本身不存配额策略）
//
// 不放这里：
//   - 登录/注册/改密（→ app/auth，业务用例）
//   - 配额策略与计数实现（→ governance/quota）
//   - API Key CRUD（→ infra/token 或后续 app）
package gateway
