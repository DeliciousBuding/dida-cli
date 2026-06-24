# DidaCLI — 风险评估

**生成日期**: 2026-06-25

## CRITICAL 风险

| # | 风险 | 详情 |
|---|---|---|
| R1 | **CLI 测试覆盖 37.8%** | 超过 60% 的命令路径未测试。ROADMAP 已知 |
| R2 | **无自动化 E2E** | 全部 live 测试为手动 |
| R3 | **无 Windows CI** | 仅 release smoke 覆盖 Windows |
| R4 | **AGENTS.md 过薄** | 仅 4 行，缺少分支策略/PR 审查/安全策略 |

## HIGH 风险

| # | 风险 | 详情 |
|---|---|---|
| R5 | **无 macOS CI** | macOS 二进制编译但从未测试 |
| R6 | **无 CLAUDE.md / 项目 memory** | Agent 无项目级别指引 |
| R7 | **OAuth 监听器 host 校验缺口** | --host 0.0.0.0 可暴露回调到局域网 |
| R8 | **无 shell completion** | ROADMAP G2，阻碍人工可用性 |

## MEDIUM 风险

| # | 风险 |
|---|---|
| R9 | 说明.md 孤立文档 |
| R10 | CI 无 -race / fuzz / benchmark |
| R11 | Homebrew/Scoop 模板未发布 |
| R12 | official call 无 dry-run 层 |
| R13 | 无 test coverage 报告 |

## S.U.P.E.R 架构健康摘要

- **S (单一职责)**: cli/ 包违规严重 — 3 个文件超 700 行
- **U (单向流)**: 健康 — 无循环导入
- **P (接口优先)**: 弱 — 无 mockable 接口，4 个包重复 token CRUD
- **E (环境无关)**: 健康 — 环境变量覆盖，路径使用 filepath
- **R (可替换)**: 混合 — HTTP 客户端可替换性差，无 DI

## 最高价值未启动项 (ROADMAP)

1. CLI 测试覆盖到 60%+
2. Shell completion
3. doctor upgrade check
4. Homebrew/Scoop 发布
5. Live OpenAPI 验证
