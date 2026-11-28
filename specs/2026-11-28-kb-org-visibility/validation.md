# Validation: 知识库 visibility=org

- [x] API：用户 A 建 `org` 库 → 用户 B 能 list / get / search（单测覆盖 list/get/read；Search 走同一 `loadKBForRead`）
- [x] API：用户 B 对 A 的 `org` 库 update / upload 失败（单测覆盖 update/delete/write）
- [x] API：用户 A 的 `private` 库对 B 不可见
- [x] Web：设置可切换 `private`/`org` 并保存
- [x] Web：非所有者进入详情时写操作禁用/隐藏
- [x] `go test` knowledge 包（含新增用例）通过
