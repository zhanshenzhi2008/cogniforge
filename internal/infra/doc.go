// Package infra 基础设施层。
//
// 子包：database / crypto / logger / httpclient / mail / trace /
// config / token / modelcache / scheduler。
//
// HTTP 中间件在 gateway；token 的 HTTP 路由在 app/token（本层只留签发/校验实现）。
//
// 依赖规则：可被上层引用；本层不得 import app / runtime / capability /
// orchestration / governance / gateway。
package infra
