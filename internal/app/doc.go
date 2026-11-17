// Package app 业务应用层：对外用例与 HTTP 入口（chat / agent / workflow / knowledge / auth …）。
//
// 依赖规则：可调用 runtime / orchestration(engine) / capability / infra；是最外层业务，不给下层反向依赖。
// 现阶段 handler 仍与各业务包同目录；后续可逐步迁到 app/<domain>。
package app
