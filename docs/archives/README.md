# 归档索引

## go-deep-quality

- **描述**: DidaCLI Go 代码库深度完善 — e2e 测试、文档、规范、结构设计
- **日期**: 2026-06-25
- **成果**: 3 安全修复、19 新测试(含 8 E2E)、3 文档、2 CI 质量门、7 文件拆分、Token CRUD 统一
- **Commits**: 42373c6, 2451b9b, 8c4a4fc
- **详情**: [MASTER.md](go-deep-quality/progress/MASTER.md)

## release-governance-optimization

- **描述**: DidaCLI CI/CD、release、npm publish、tag、changelog、provenance 治理优化；GitHub 记录见 [DeliciousBuding/dida-cli](https://github.com/DeliciousBuding/dida-cli)
- **日期**: 2026-07-06 至 2026-07-07
- **模式**: GITHUB_STANDARD
- **成果**: PR #2 恢复并合并、`v0.2.5` 发布、npm README 修复、CI/release gates、CodeQL、Scorecard、SHA-pinned Actions、release archive attestation wiring
- **详情**: [MASTER.md](release-governance-optimization/progress/MASTER.md)

## shell-completion-productization

- **描述**: 增加 `dida completion <bash|zsh|fish|powershell>`，并同步 schema、README、命令参考、roadmap、changelog
- **日期**: 2026-07-07
- **模式**: LOCAL_ONLY
- **成果**: 4 种 shell completion 脚本、本地测试、用户文档和路线图更新
- **详情**: [MASTER.md](shell-completion-productization/progress/MASTER.md)

## doctor-upgrade-check

- **描述**: 增加 `dida doctor --check-upgrade`，在 doctor 诊断中显式报告 GitHub Release 更新状态
- **日期**: 2026-07-07
- **模式**: LOCAL_ONLY
- **成果**: 默认保持本地检查；显式检查时输出 `upgrade_check`，升级查询失败只作为诊断项
- **详情**: [MASTER.md](doctor-upgrade-check/progress/MASTER.md)

## staticcheck-quality-gate

- **描述**: 将 Staticcheck 接入本地、CI、release validate、release-check 和仓库治理校验
- **日期**: 2026-07-07
- **模式**: LOCAL_ONLY
- **成果**: `make staticcheck` 使用 `honnef.co/go/tools/cmd/staticcheck@v0.7.0`；治理脚本保护 CI/release/PR checklist 入口
- **详情**: [MASTER.md](staticcheck-quality-gate/progress/MASTER.md)

## cli-coverage-improvement

- **描述**: 增加本地 CLI 回归测试和 `make coverage-cli`，提高 `internal/cli` 覆盖率基线
- **日期**: 2026-07-07
- **模式**: LOCAL_ONLY
- **成果**: `internal/cli` 覆盖率从 43.9% 提升到 50.8%；覆盖 task/project dry-run、本地 help、sync-backed reads、OpenAPI task dry-run
- **详情**: [MASTER.md](cli-coverage-improvement/progress/MASTER.md)
