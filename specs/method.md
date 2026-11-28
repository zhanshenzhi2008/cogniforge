# Method

本仓遵循全局 SDD 规则：`~/.cursor/rules/sdd-development.mdc`（colincalnan 风格）。

## 要点

- 活规格只在 `specs/`；入口见仓库根 `AGENTS.md`
- 行为变更：roadmap 挂项 → feature 三件套（`requirements.md` / `plan.md` / `validation.md`）→ 再改代码
- 可跳过 feature spec：typo、纯注释、纯格式、不改对外行为的重构
- 旧 `docs/` 默认归档；运维手册例外见 `AGENTS.md`
- 未知写 `_TBD_`，不编造栈或承诺

Feature 目录命名：`specs/YYYY-MM-DD-<slug>/`。
