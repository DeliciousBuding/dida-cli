<p align="center">
  <img src="assets/hero.svg" alt="DidaCLI - 面向 Dida365 / TickTick 的整洁、适合 Agent 的命令行工具" width="100%">
</p>

<p align="center">
  <a href="https://github.com/DeliciousBuding/dida-cli/actions/workflows/ci.yml"><img alt="CI" src="https://img.shields.io/github/actions/workflow/status/DeliciousBuding/dida-cli/ci.yml?branch=main&label=ci&logo=github-actions"></a>
  <a href="https://github.com/DeliciousBuding/dida-cli/releases/latest"><img alt="Latest Release" src="https://img.shields.io/github/v/release/DeliciousBuding/dida-cli?logo=github"></a>
  <a href="https://github.com/DeliciousBuding/dida-cli/blob/main/LICENSE"><img alt="License" src="https://img.shields.io/github/license/DeliciousBuding/dida-cli"></a>
  <img alt="Go" src="https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white">
  <a href="https://github.com/DeliciousBuding/dida-cli/releases"><img alt="Downloads" src="https://img.shields.io/github/downloads/DeliciousBuding/dida-cli/total?logo=github"></a>
</p>

<p align="center">
  <b>面向 <a href="https://dida365.com">滴答清单</a> / <a href="https://ticktick.com">TickTick</a> 的整洁、适合 Agent 的命令行工具</b>
</p>

<p align="center">
  <a href="README.md">English</a> &nbsp;&middot;&nbsp;
  <a href="https://deliciousbuding.github.io/dida-cli/">项目主页</a> &nbsp;&middot;&nbsp;
  <a href="docs/quickstart.zh-CN.md">快速开始</a> &nbsp;&middot;&nbsp;
  <a href="docs/commands.md">命令参考</a> &nbsp;&middot;&nbsp;
  <a href="docs/README.md">文档索引</a>
</p>

---

## 为什么选 DidaCLI

滴答清单有很好的 UI，但没有稳定的命令行工具来做自动化。DidaCLI 用一个 Go 二进制文件填补了这个空白，同时接入滴答清单的私有 Web API、官方 MCP 和官方 OpenAPI -- 为人类 **和** AI Agent 设计，提供可预测、结构化的任务操作。

```
$ dida +today --json
{
  "ok": true,
  "command": "task today",
  "meta": { "count": 3 },
  "data": {
    "tasks": [
      { "title": "Review lab notes", "priority": 3, "status": 0 },
      { "title": "Push to main",     "priority": 1, "status": 0 },
      { "title": "Write tests",      "priority": 0, "status": 0 }
    ]
  }
}
```

## 亮点

| | 特性 | 描述 |
|---|---|---|
| **30+ 命令** | 完整 CRUD | 任务、项目、文件夹、标签、列、评论、过滤器、习惯、番茄钟、回收站、搜索、统计 |
| **Agent 安全 JSON** | 一致的信封结构 | 每个 `--json` 响应使用 `{ ok, command, meta, data, error }` |
| **三通道认证** | Web API + MCP + OpenAPI | 浏览器 Cookie、官方 Token、OAuth -- 互不混用 |
| **Dry-run 写入** | 先预览再执行 | 所有写命令支持 `--dry-run` 预览请求体 |
| **零依赖** | 单个二进制 | 纯 Go stdlib，无 CGO，无运行时依赖 |
| **六平台** | 交叉编译 | Windows / Linux / macOS，amd64 和 arm64 |

## 安装

**macOS / Linux:**

```bash
curl -fsSL https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.sh | sh
```

**Windows PowerShell:**

```powershell
iwr https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.ps1 -UseB | iex
```

**Go:**

```bash
go install github.com/DeliciousBuding/dida-cli/cmd/dida@latest
```

<details>
<summary><b>锁定特定版本</b></summary>

```bash
# macOS / Linux
DIDA_VERSION=v0.2.0 curl -fsSL https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.sh | sh

# Windows
$env:DIDA_VERSION="v0.2.0"; iwr https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.ps1 -UseB | iex
```

</details>

## 快速开始

```bash
# 1. 登录（打开浏览器，仅捕获 t cookie）
dida auth login --browser --json

# 2. 验证环境
dida doctor --verify --json

# 3. 查看今日任务
dida +today --json

# 4. 创建任务（先 dry-run）
dida task create --project <id> --title "Ship v1" --dry-run --json
dida task create --project <id> --title "Ship v1" --json

# 5. 获取 Agent 上下文包
dida agent context --outline --json
```

## 命令速览

<details>
<summary><b>读取数据</b></summary>

```bash
dida +today --json                       # 今日任务
dida task upcoming --days 14 --json      # 未来两周
dida task search --query "exam" --json   # 搜索任务
dida project list --json                 # 所有项目
dida folder list --json                  # 所有文件夹
dida tag list --json                     # 所有标签
dida completed today --json              # 今日已完成
dida pomo stats --json                   # 番茄钟统计
dida stats general --json                # 账户统计
```

</details>

<details>
<summary><b>写入数据</b></summary>

```bash
dida task create --project <id> --title "新任务" --json
dida task update <task-id> --project <id> --title "更新标题" --json
dida task complete <task-id> --project <id> --json
dida task move <task-id> --project <id> --to-project <dest-id> --json
dida task delete <task-id> --project <id> --yes --json
dida project create --name "新项目" --json
dida tag create my-tag --json
```

</details>

<details>
<summary><b>官方通道</b></summary>

```bash
# 官方 MCP（基于 Token）
DIDA365_TOKEN=dp_xxx dida official doctor --json
dida official project list --json
dida official task query --query "today" --json

# 官方 OpenAPI（基于 OAuth）
dida openapi client set --id <client-id> --secret-stdin --json
dida openapi login --browser --json
dida openapi project list --json
```

</details>

完整命令参考：[docs/commands.md](docs/commands.md)

## 架构

```
                    dida-cli
            ┌──────────┼──────────┐
            │          │          │
        Web API    Official MCP  OpenAPI
        (cookie)    (token)     (OAuth)
            │          │          │
            └──────────┼──────────┘
                       │
              ┌────────┴────────┐
              │   CLI 层        │  30+ 命令，JSON 信封
              │   internal/cli/ │  Schema 注册表，dry-run
              ├─────────────────┤
              │   模型层        │  归一化 Task/Project/Column
              │ internal/model/ │  搜索、过滤、upcoming
              ├─────────────────┤
              │   API 客户端    │  HTTP + MCP 协议
              │ internal/webapi │
              │ internal/officialmcp
              │ internal/openapi│
              └─────────────────┘
```

## Agent 集成

DidaCLI 从一开始就为 AI Agent 工作流设计。Agent 可以：

1. **发现命令** -- `dida schema list --compact --json`
2. **构建上下文** -- `dida agent context --outline --json`
3. **预览写入** -- 写入前用 `--dry-run` 预览
4. **解析响应** -- 从稳定的 JSON 信封中提取数据

仓库内自带 Agent Skill：[`skills/dida-cli/SKILL.md`](skills/dida-cli/SKILL.md)

| Agent | 安装说明 |
|---|---|
| Claude Code | 复制 `skills/dida-cli/SKILL.md` 到你的 skills 目录 |
| Codex | 参见 [docs/skill-installation.md](docs/skill-installation.md) |
| Hermes | 参见 [docs/skill-installation.md](docs/skill-installation.md) |

## 为什么不直接用官方 API？

| | DidaCLI Web API | 官方 OpenAPI | 官方 MCP |
|---|---|---|---|
| **认证** | 浏览器 Cookie | OAuth 应用 | Token |
| **覆盖面** | 最广（私有端点） | 项目、任务、专注、习惯 | 工具型（MCP 协议） |
| **写入安全** | Dry-run + 确认 | Dry-run | Dry-run（本地预览） |
| **Agent 模式** | JSON 信封、Schema | JSON 响应 | MCP 工具 Schema |
| **配置成本** | 一次浏览器登录 | 注册 OAuth 应用 | 获取 Token |

Web API 覆盖面最大，OpenAPI 适合标准 REST 集成，MCP 适合官方工具接入。三者认证通道独立，绝不混用。

## 文档

- [快速开始](docs/quickstart.zh-CN.md) -- 2 分钟上手
- [命令参考](docs/commands.md) -- 每个命令、每个参数
- [Agent 使用指南](docs/agent-usage.md) -- AI Agent 如何使用 DidaCLI
- [API 覆盖面](docs/api-coverage.md) -- 覆盖了哪些端点
- [OpenAPI 设置](docs/openapi-setup.zh-CN.md) -- OAuth 通道配置

## 参与贡献

欢迎贡献。参见 [CONTRIBUTING.md](CONTRIBUTING.md)。

```bash
git clone https://github.com/DeliciousBuding/dida-cli.git
cd dida-cli
go test ./...
go build -o bin/dida ./cmd/dida
```

## 许可证

[MIT](LICENSE)
