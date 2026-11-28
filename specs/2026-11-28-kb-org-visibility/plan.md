# Plan: 知识库 visibility=org

## 步骤

- [x] 1. Go：`normalizeVisibility` / `loadKBForRead` / `loadKBForWrite`
- [x] 2. Go：List 含 `user_id=me OR visibility=org`；读路径用 Read；写路径用 Write
- [x] 3. Go：创建/更新持久化 `org`（去掉强制 private）
- [x] 4. Go 单测：跨用户读/写边界
- [x] 5. Web：设置可见范围 + 非所有者只读 + i18n
- [x] 6. 勾 validation；roadmap Now #3 → complete

## 主要改动文件

- `internal/capability/knowledge/service.go`（+ 可选 `access.go`）
- `internal/capability/knowledge/service_test.go` / 新测试
- `internal/model/kb.go`（注释）
- `cogniforge-web/pages/knowledge/[id].vue`、`i18n/messages.ts`、`useKnowledgeBases.ts`
