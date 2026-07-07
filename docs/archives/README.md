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

## webapi-openapi-read-coverage

- **描述**: 继续补充 Web API 读命令和 OpenAPI 读命令的本地 CLI 测试
- **日期**: 2026-07-07
- **模式**: LOCAL_ONLY
- **成果**: `internal/cli` 覆盖率从 50.8% 提升到 61.3%；覆盖 folder/tag/quadrant/agent context/closed reads 和 OpenAPI project/task/focus/habit reads
- **详情**: [MASTER.md](webapi-openapi-read-coverage/progress/MASTER.md)

## roadmap-governance-freshness

- **描述**: 增加 roadmap freshness 校验，防止 `ROADMAP.md` 的当前版本和下一里程碑继续滞后
- **日期**: 2026-07-07
- **模式**: LOCAL_ONLY
- **成果**: `make release-check VERSION=v0.2.5` 现在会运行 `scripts/validate-roadmap.sh` 和对应 mutation tests
- **详情**: [MASTER.md](roadmap-governance-freshness/progress/MASTER.md)

## website-product-polish

- **描述**: 对齐 GitHub Pages 首页与 `v0.2.5` 当前安装、认证、诊断、schema、latest task 和 completion 路径
- **日期**: 2026-07-07
- **模式**: LOCAL_ONLY
- **成果**: 首页去掉旧命令示例和易失效父级资源路径；`make release-check VERSION=v0.2.5` 现在会运行 website copy 校验和 mutation tests
- **详情**: [MASTER.md](website-product-polish/progress/MASTER.md)

## package-manager-template-automation

- **描述**: 将 Homebrew、Scoop、packaging README 和 winget notes 从手工维护改为由 release `checksums.txt` 生成
- **日期**: 2026-07-07
- **模式**: LOCAL_ONLY
- **成果**: 新增 `scripts/update-packaging-templates.sh` 和 mutation tests；`make release-check VERSION=v0.2.5` 会验证生成器
- **详情**: [MASTER.md](package-manager-template-automation/progress/MASTER.md)

## release-strategy-goreleaser-decision

- **描述**: 记录 `v0.3.x` 保留当前 release workflow、暂不迁移 GoReleaser 的决策和重评条件
- **日期**: 2026-07-07
- **模式**: LOCAL_ONLY
- **成果**: 新增 `docs/research/release-strategy-goreleaser.md` 和 `scripts/validate-release-strategy.sh`；`make release-check VERSION=v0.2.5` 会阻止 roadmap 回到 undecided 状态
- **详情**: [MASTER.md](release-strategy-goreleaser-decision/progress/MASTER.md)

## package-manager-repo-export

- **描述**: 从已验证的 Homebrew/Scoop 模板导出可单独建仓库的 tap/bucket 根目录，但不自动创建或发布外部仓库
- **日期**: 2026-07-07
- **模式**: LOCAL_ONLY
- **成果**: 新增 `scripts/export-package-manager-repos.sh` 和测试；`make release-check VERSION=v0.2.5` 现在覆盖外部仓库布局导出
- **详情**: [MASTER.md](package-manager-repo-export/progress/MASTER.md)

## package-manager-release-artifact

- **描述**: 在 release workflow 创建 GitHub Release 后，自动生成并上传 Homebrew tap / Scoop bucket 仓库导出 artifact
- **日期**: 2026-07-07
- **模式**: LOCAL_ONLY
- **成果**: 新增 `package-manager-export` release job；仓库治理校验保护 artifact 名称、路径、保留期和导出命令
- **详情**: [MASTER.md](package-manager-release-artifact/progress/MASTER.md)

## research-audit-freshness

- **描述**: 将 objective/distribution 审计文档更新到 `v0.2.5`，并增加校验防止 research audit 回到旧 release 基线
- **日期**: 2026-07-07
- **模式**: LOCAL_ONLY
- **成果**: 新增 `scripts/validate-research-audit.sh` 和 mutation tests；`make release-check VERSION=v0.2.5` 现在覆盖 research audit freshness
- **详情**: [MASTER.md](research-audit-freshness/progress/MASTER.md)

## roadmap-distribution-freshness

- **描述**: 将 `ROADMAP.md` 详细分发状态从旧 release 证据更新到 `v0.2.5`，并扩展 roadmap freshness 校验
- **日期**: 2026-07-07
- **模式**: LOCAL_ONLY
- **成果**: F1/F2/F3 分发段落对齐当前 release/npm/package-manager artifact；`validate-roadmap` 会拒绝旧分发基线回流
- **详情**: [MASTER.md](roadmap-distribution-freshness/progress/MASTER.md)
