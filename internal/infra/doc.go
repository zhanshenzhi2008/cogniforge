// Package infra 基础设施层：DB / Redis 客户端 / 加解密 / 日志 / HTTP 客户端 / 邮件 / Trace。
//
// 依赖规则：可被上层引用；本层不得 import app / runtime / capability / orchestration 业务包。
package infra
