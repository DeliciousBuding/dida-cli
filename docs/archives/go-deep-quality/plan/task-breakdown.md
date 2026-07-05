# DidaCLI 深度完善计划

**方向**: 对 Go 代码库进行 e2e 真实测试、文档完善、规范梳理、结构设计
**模式**: GITHUB_STANDARD（本地文件跟踪）
**生成日期**: 2026-06-25

## Phase 1: 安全修复 (S.U.P.E.R HIGH)

| # | 任务 | 优先级 | 文件 | 验收标准 |
|---|---|---|---|---|
| 1.1 | OpenAPI 错误路径 token 脱敏 | P0 | `openapi/client.go` | summarizeBody 脱敏 Bearer token |
| 1.2 | webapi.Client 默认超时 | P0 | `webapi/client.go` | NewClient() 设置默认超时 |

## Phase 2: CLI 测试覆盖提升 (37.8% → 60%+)

| # | 任务 | 优先级 | 范围 | 验收标准 |
|---|---|---|---|---|
| 2.1 | task 命令全路径测试 | P0 | task_cmd.go | create/update/complete/delete 实际命令路径 |
| 2.2 | tag/folder/column 写命令测试 | P0 | tag_cmd.go, folder_cmd.go, column_cmd.go | 全部写操作覆盖 |
| 2.3 | official MCP 子命令测试 | P1 | official_cmd.go | 读命令 + dry-run 写命令 |
| 2.4 | openapi 子命令测试 | P1 | openapi_cmd.go | 读命令 + dry-run 写命令 |
| 2.5 | 读类命令测试 (pomo/habit/user/share等) | P1 | account_read_cmd.go, productivity_cmd.go | 全部读命令覆盖 |
| 2.6 | 集成/E2E 测试框架 | P0 | tests/e2e/ | fake transport 的 E2E 测试 |

## Phase 3: 文档完善

| # | 任务 | 优先级 | 范围 | 验收标准 |
|---|---|---|---|---|
| 3.1 | AGENTS.md 扩展 | P0 | AGENTS.md | 分支策略、PR审查、提交规范、安全策略 |
| 3.2 | CLAUDE.md 创建 | P0 | CLAUDE.md | 项目指引、memory 索引 |
| 3.3 | 说明.md 整合 | P1 | docs/ → 说明.md | 移入 docs/ 并索引 |
| 3.4 | SKILL.md 更新 | P1 | skills/dida-cli/SKILL.md | 补全 attachment download |

## Phase 4: 规范与 CI

| # | 任务 | 优先级 | 范围 | 验收标准 |
|---|---|---|---|---|
| 4.1 | Windows CI | P0 | ci.yml | test + build on windows-latest |
| 4.2 | -race + -coverprofile in CI | P1 | ci.yml | race detector + coverage |
| 4.3 | 凭证原子写入 | P1 | auth/cookie.go | 临时文件 + rename |
| 4.4 | OAuth 监听器 host 校验 | P1 | openapi_cmd.go | 拒绝非 loopback host |

## Phase 5: 结构优化

| # | 任务 | 优先级 | 范围 | 验收标准 |
|---|---|---|---|---|
| 5.1 | account_read_cmd.go 拆分 | P2 | cli/ | 按域拆分为 user_cmd.go, share_cmd.go 等 |
| 5.2 | Token-file CRUD 统一 | P2 | auth/ | 抽象通用 token store 接口 |
| 5.3 | 类型转换辅助函数去重 | P2 | model/ + webapi/ | 统一到 model/ |
