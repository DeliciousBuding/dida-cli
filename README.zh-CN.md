<h1 align="center">DidaCLI</h1>

<p align="center">
  <b>面向 Dida365 / TickTick 的整洁、适合 Agent 的命令行工具。</b>
</p>

<p align="center">
  <a href="README.md">English README</a> ·
  <a href="https://deliciousbuding.github.io/dida-cli/">项目主页</a> ·
  <a href="#快速开始">快速开始</a> ·
  <a href="#常用命令">常用命令</a> ·
  <a href="#给-agent-使用">给 Agent 使用</a> ·
  <a href="docs/quickstart.zh-CN.md">命令文档</a>
</p>

---

## 简介

DidaCLI 是一个用 Go 编写的滴答清单 / TickTick 命令行工具，目标不是做一层薄封装，而是做一个稳定、可测试、适合人和 Agent 共同使用的 CLI。

它优先接入 Dida365 Web API，而不是只依赖公开 Open API。这样能覆盖更完整的个人账号操作面，同时保持命令显式、输出稳定、行为可审计。

## 设计目标

- 命令要短，结构要清晰，适合人工直接使用。
- `--json` 输出要稳定，适合 Codex、Hermes、Claude Code 这类 Agent 自动解析。
- 写操作要有边界：支持 `--dry-run`，破坏性动作要求明确确认。
- 安全性要够用：不默认打印 cookie，不开放原始写 API 通道，不把敏感输入强塞进命令历史。
- 文档、Schema、Skill 都要跟得上，不能只有代码。
- 通道策略要清楚：Web API 负责覆盖面，官方 MCP 负责正规 token 接入与能力基线。
- 第三通道要预留：官方 OpenAPI 走 OAuth，适合后续做标准 REST 集成。

## 已支持能力

- 同步与账户视图：`sync all`、`sync checkpoint`、`agent context`
- 任务体系：任务读写、搜索、今日、未来、优先级、子任务、移动、完成、删除
- 项目体系：项目、文件夹、标签、列、评论
- 历史与统计：已完成任务、closed history、trash、统计总览
- 专注与习惯：Pomodoro、timeline、habit、checkins
- 元数据读取：提醒、分享、日历、模板、搜索、用户会话、Web 侧设置
- `schema list/show`：给 Agent 的命令契约面
- `raw get`：只读探测通道

## 快速开始

### 一行安装

macOS / Linux：

```bash
curl -fsSL https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.sh | sh
```

Windows PowerShell：

```powershell
iwr https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.ps1 -UseB | iex
```

安装脚本会从 GitHub Releases 下载最新版本，校验 `checksums.txt`，安装到用户本地 bin 目录，并运行：

```bash
dida version
dida doctor --json
```

可选环境变量：

| 变量 | 作用 |
| --- | --- |
| `DIDA_VERSION` | 安装指定 release tag，例如 `v0.1.0`。 |
| `DIDA_INSTALL_DIR` | 覆盖安装目录。 |
| `DIDA_REPO` | 覆盖 GitHub 仓库，例如安装 fork。 |

### 登录与首读

```bash
dida auth login --browser --json
dida doctor --json
dida agent context --json
dida schema list --json
```

### 官方通道

```bash
DIDA365_TOKEN=... dida official doctor --json
dida openapi doctor --json
```

Web API、官方 MCP、官方 OpenAPI 是三套不同认证通道，不要混用 cookie 或 token。

### Agent 提示

本节面向 LLM/Agent 使用。优先使用 JSON 输出；写操作前先查
`dida schema list --json`；生成的写入先跑 `--dry-run`；不要要求用户把
cookie 或 token 发到聊天里。

### 从源码安装

开发时使用这个路径：

```bash
git clone https://github.com/DeliciousBuding/dida-cli.git
cd dida-cli
go test ./...
go build -o bin/dida ./cmd/dida
```

完整文档：

- [快速开始](docs/quickstart.zh-CN.md)
- [LLM / Agent Quickstart](docs/llm-quickstart.md)
- [分发说明](docs/distribution.md)

## 常用命令

### 读取

```bash
dida doctor --json
dida official doctor --json
dida official tools --limit 20 --json
dida official show list_projects --json
dida official call list_projects --json
dida official call list_undone_tasks_by_time_query --args-json "{\"query_command\":\"today\"}" --json
dida official task query --query "today" --json
dida official task batch-add --args-json "{\"tasks\":[{\"title\":\"任务\"}]}" --dry-run --json
dida openapi doctor --json
dida openapi status --json
dida openapi login --json
dida openapi auth-url --json
dida openapi exchange-code --code <code> --json
dida openapi project list --json
dida openapi focus list --from 2026-04-01T00:00:00+0800 --to 2026-04-02T00:00:00+0800 --type 1 --json
dida openapi habit list --json
dida schema list --json
dida agent context --json
dida auth status --verify --json
dida settings get --json
dida settings get --include-web --json
dida project list --json
dida task today --json
dida task upcoming --days 14 --json
dida completed today --json
dida closed list --status 2 --from 2026-05-01 --to 2026-05-09 --json
dida trash list --cursor 20 --compact --json
dida search all --query "计算机" --limit 20 --json
dida stats general --json
dida template project list --limit 20 --json
dida user sessions --limit 10 --json
dida pomo stats --json
dida pomo timeline --limit 20 --json
dida habit checkins --habit <habit-id> --json
```

### 写入

```bash
dida task create --project <project-id> --title "新任务" --dry-run --json
dida task create --project <project-id> --title "新任务" --json
dida task update <task-id> --project <project-id> --title "改标题" --json
dida task complete <task-id> --project <project-id> --json
dida task delete <task-id> --project <project-id> --yes --json
dida project create --name "新项目" --dry-run --json
dida folder create --name "工作" --dry-run --json
dida tag create planning --dry-run --json
```

完整命令列表看 [docs/commands.md](docs/commands.md)。

## 给 Agent 使用

推荐先拿上下文，再决定后续动作：

```bash
dida doctor --json
dida schema list --json
dida agent context --json
dida auth status --verify --json
```

推荐写入流程：

```bash
dida task create --project <project-id> --title "Agent-created task" --dry-run --json
dida task create --project <project-id> --title "Agent-created task" --json
```

仓库内自带 Skill：

```text
skills/dida-cli/SKILL.md
```

给 Codex、Claude Code、OpenClaw、Hermes Agent 的安装说明见 [docs/skill-installation.md](docs/skill-installation.md)。

## Web API 说明

当前 CLI 已覆盖一大批实测可用的 Web API，包括：

- `/batch/check/...`
- `/user/preferences/settings`
- 任务 / 项目 / 文件夹 / 标签 / 评论
- `/pomodoros...`、`/habit...`
- `/statistics/general`
- `/projectTemplates/all`
- `/search/all`
- 只读 `raw get`

官方 MCP 与 Web API 的对比说明见：

- [docs/research/official-mcp-vs-webapi.md](docs/research/official-mcp-vs-webapi.md)
- [docs/research/official-openapi-guide.md](docs/research/official-openapi-guide.md)

更细的端点级说明见：

- [docs/web-api.md](docs/web-api.md)
- [docs/api-coverage.md](docs/api-coverage.md)
- [docs/research/api-surfaces.md](docs/research/api-surfaces.md)

## 项目结构

```text
cmd/dida/          CLI 入口
internal/auth/     登录、cookie、浏览器采集
internal/cli/      命令分发与 JSON 输出
internal/model/    归一化任务/项目模型
internal/webapi/   Dida365 Web API 客户端
docs/              文档
skills/dida-cli/   仓库内 Agent Skill
```

## 开发

```bash
go test ./...
go build -o bin/dida ./cmd/dida
```

CI 默认执行：

- `go test ./...`
- `go vet ./...`
- `govulncheck`

## License

MIT，见 [LICENSE](LICENSE)。
