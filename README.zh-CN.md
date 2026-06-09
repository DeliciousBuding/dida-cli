<p align="center">
  <img src="assets/logo-icon.svg" alt="DidaCLI" width="100">
</p>

<h1 align="center">DidaCLI</h1>

<p align="center">
  <b>面向 <a href="https://dida365.com">Dida365</a> / <a href="https://ticktick.com">TickTick</a> 的 JSON 优先 CLI</b>
</p>

<p align="center">
  <a href="https://github.com/DeliciousBuding/dida-cli/blob/main/LICENSE"><img alt="License: MIT" src="https://img.shields.io/badge/License-MIT-blue?style=flat-square"></a>
  <img alt="Go 1.26+" src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat-square&logo=go&logoColor=white">
  <a href="https://github.com/DeliciousBuding/dida-cli/releases/latest"><img alt="Latest Release" src="https://img.shields.io/github/v/release/DeliciousBuding/dida-cli?style=flat-square&logo=github"></a>
  <a href="https://github.com/DeliciousBuding/dida-cli/actions/workflows/ci.yml"><img alt="CI" src="https://img.shields.io/github/actions/workflow/status/DeliciousBuding/dida-cli/ci.yml?branch=main&label=ci&style=flat-square&logo=github-actions"></a>
  <a href="https://www.npmjs.com/package/@delicious233/dida-cli"><img alt="npm" src="https://img.shields.io/npm/v/@delicious233/dida-cli?style=flat-square&logo=npm"></a>
</p>

<p align="center">
  <a href="README.md">English</a> ·
  <a href="https://deliciousbuding.github.io/dida-cli/">项目主页</a> ·
  <a href="docs/commands.md">命令参考</a>
</p>

---

DidaCLI 构建为单个 Go 二进制文件，不依赖外部 Go 模块。Web API Cookie、官方 MCP Token、OpenAPI OAuth 三条认证通道相互独立。命令返回稳定的 JSON 信封。

```bash
$ dida task today --compact --json
```

## 安装

```bash
npm i -g @delicious233/dida-cli          # npm
curl -fsSL https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.sh | sh
iwr https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.ps1 -UseB | iex
```

<details>
<summary>全部安装方式</summary>

### npm（推荐）

```bash
npm install -g @delicious233/dida-cli
```

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.sh | sh
```

### Windows（PowerShell）

```powershell
iwr https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.ps1 -UseB | iex
```

### Go

```bash
go install github.com/DeliciousBuding/dida-cli/cmd/dida@latest
```

### 锁定版本

```bash
DIDA_VERSION=vX.Y.Z curl -fsSL https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.sh | sh
```

安装后：

```bash
dida version && dida doctor --json
dida upgrade --check
```

</details>

## 快速开始

```bash
# 1. 使用 Dida365 浏览器 Cookie "t" 登录
dida auth cookie set --token-stdin --json

# 可选：用本地浏览器自动捕获 Cookie
dida auth login --browser --json

# 2. 验证
dida doctor --verify --json

# 3. 查看今日
dida +today --json

# 4. 创建任务（先预览）
dida task create --project <id> --title "发布 v1" --dry-run --json
```

## 命令

<details>
<summary>读取</summary>

```bash
dida task today --json
dida task upcoming --days 14 --json
dida task search --query "考试" --json
dida project list --json
dida tag list --json
dida completed today --json
dida pomo stats --json
```
</details>

<details>
<summary>写入</summary>

```bash
dida task create --project <id> --title "新任务" --json
dida task update <id> --project <p> --title "更新" --json
dida task complete <id> --project <p> --json
dida task move <id> --from <p> --to <p> --json
dida task delete <id> --project <p> --yes --json
```
</details>

<details>
<summary>官方通道（MCP & OpenAPI）</summary>

```bash
# MCP（Token）
DIDA365_TOKEN=dp_xxx dida official project list --json

# OpenAPI（OAuth）
dida openapi client set --id <id> --secret-stdin --json
dida openapi login --browser --json
dida openapi project list --json
```
</details>

完整参考：[docs/commands.md](docs/commands.md)

## 认证通道

| | Web API | 官方 MCP | 官方 OpenAPI |
|---|---|---|---|
| **认证** | 浏览器 Cookie | Token | OAuth |
| **覆盖面** | 官方通道外的 Web API 资源 | MCP 工具型 | 标准 REST |
| **配置** | 一次登录 | 获取 Token | 注册应用 |

三条认证通道独立。

## Agent 集成

```bash
dida schema list --compact --json        # 发现命令
dida agent context --outline --json      # 构建上下文
dida task create ... --dry-run --json    # 预览写入
```

| Agent | 安装 |
|---|---|
| Claude Code | 复制 [`skills/dida-cli/SKILL.md`](skills/dida-cli/SKILL.md) |
| Codex / 其他 | 参见 [docs/skill-installation.md](docs/skill-installation.md) |

支持 `--dry-run` 的资源写入应先预览。破坏性操作需要 `--yes`。CLI 不会打印完整 Token。参见 [Agent 使用指南](docs/agent-usage.md)。

## 文档

- [命令参考](docs/commands.md)
- [Agent 使用指南](docs/agent-usage.md)
- [API 覆盖面](docs/api-coverage.md)
- [OpenAPI 设置](docs/openapi-setup.zh-CN.md)

## 参与贡献

```bash
git clone https://github.com/DeliciousBuding/dida-cli.git && cd dida-cli
go test ./... && go build -o bin/dida ./cmd/dida
```

参见 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 许可证

[MIT](LICENSE)

---

DidaCLI 是面向 [Dida365](https://dida365.com) / [TickTick](https://ticktick.com) 兼容工作流的独立开源 CLI。请只在你控制的账号和自动化流程中使用，并遵守上游服务条款。
