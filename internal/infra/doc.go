// Package infra 基础设施层。
//
// 子包：database / crypto / logger / httpclient / mail / trace /
// config / token / modelcache / scheduler。
//
// HTTP 中间件（CORS / JWT / RBAC 门禁）已迁至 gateway。
//
// 依赖规则：可被上层引用；本层不得 import app / runtime / capability /
// orchestration / governance / gateway。
package infra
