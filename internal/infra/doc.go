// Package infra 基础设施层。
//
// 子包：database / crypto / logger / httpclient / mail / trace /
// middleware / config / token / modelcache / scheduler。
//
// 依赖规则：可被上层引用；本层不得 import app / runtime / capability / orchestration / governance。
package infra
