# MASTER.md — DidaCLI 深度完善

**任务**: DidaCLI Go 代码库深度完善 — e2e 测试、文档、规范、结构
**模式**: LOCAL_ONLY
**开始日期**: 2026-06-25
**最后更新**: 2026-06-25
**Commit**: 42373c6

## 文档索引

- [项目概览](../analysis/project-overview.md)
- [模块清单](../analysis/module-inventory.md)
- [风险评估](../analysis/risk-assessment.md)
- [任务分解](../plan/task-breakdown.md)

## 进度

- [x] Phase 1: 安全修复 (3/3 tasks) — commit 42373c6
  - [x] 1.1 OpenAPI 错误路径 token 脱敏
  - [x] 1.2 webapi.Client 默认超时 (30s)
  - [x] 1.3 OAuth 回调监听器 host 校验 (新增)
- [x] Phase 2: 测试覆盖 (11 new tests) — commit 42373c6
  - [x] 2.1 task 命令测试 (4 tests: update, complete, filter, due-counts)
  - [x] 2.2 tag/column/folder 写测试 (4 tests: rename, merge, column, folder)
  - [x] 2.3 读命令测试 (5 tests: share, calendar, stats, pomo, user)
  - 覆盖率: 37.8% → 43.0%
- [x] Phase 3: 文档完善 (3/3 tasks) — commit 42373c6
  - [x] 3.1 AGENTS.md 扩展 (4→8 sections)
  - [x] 3.2 CLAUDE.md 创建
  - [x] 3.3 说明.md → docs/attachment-download.md
- [x] Phase 4: 规范与 CI (2/2 tasks) — commit 42373c6
  - [x] 4.1 Windows CI (test + build matrix)
  - [x] 4.2 -race + -coverprofile in CI
- [ ] Phase 5: 结构优化 (0/3 tasks) — deferred P2

## 当前状态

Phase 1-4 完成。Phase 5 (结构优化: 文件拆分、token CRUD 统一、辅助函数去重) 为 P2，后续迭代处理。
