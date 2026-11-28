# Requirements: 知识库 visibility=org（本站共享）

## 背景

库级字段 `visibility` 已存在，但创建/更新强制 `private`；列表与检索一律按 `user_id` 隔离。UI 文案「组织内」为占位。本轮**不建**部门/组织表。

## 目标

| ID | 需求 | 优先级 |
|----|------|--------|
| KB-VIS-1 | `visibility` 取值仅 `private` \| `org`；非法值回落 `private` | P0 |
| KB-VIS-2 | `org` = 本部署**全体登录用户可读**（list/get/文档只读/检索/RAG/绑 Agent） | P0 |
| KB-VIS-3 | 写操作（改设置、删库、上传/改/删文档、lifecycle、reparse）仅所有者 | P0 |
| KB-VIS-4 | `private` 库对非所有者不可见、不可检索 | P0 |
| KB-VIS-5 | Web 设置可切换并保存；非所有者 UI 只读；去掉「即将支持」 | P0 |

## 非目标

- 部门表、组织成员、按角色共享、文档级 ACL
- 改 Python RAG 鉴权（集合仍按 kbID；鉴权在 Go）
- 改旧 `docs/` PRD 树

## 验收口径

见同目录 `validation.md`。
