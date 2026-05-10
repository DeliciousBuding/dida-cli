# 快速开始

这份文档面向普通用户和操作者，用于安装、验证、登录并开始使用 DidaCLI。

## 安装

Unix-like 系统：

```bash
curl -fsSL https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.sh | sh
```

Windows PowerShell：

```powershell
iwr https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.ps1 -UseB | iex
```

可选环境变量：

| 变量 | 作用 |
| --- | --- |
| `DIDA_VERSION` | 安装指定 release tag，例如 `v0.1.0`。 |
| `DIDA_INSTALL_DIR` | 覆盖安装目录。 |
| `DIDA_REPO` | 覆盖 GitHub 仓库，例如安装 fork。 |

## 验证

```bash
dida version
dida doctor --json
```

默认的 `doctor --json` 只检查本地状态。登录后可以用 `--verify` 做一次只读
Web API 健康检查：

```bash
dida doctor --verify --json
```

## Web API 登录

Web API 通道使用本地浏览器登录后的会话 cookie。

```bash
dida auth login --browser --json
dida doctor --verify --json
dida auth status --verify --json
```

不要把 cookie 发到聊天或 issue 里。确实需要手动导入时，用 stdin：

```bash
dida auth cookie set --token-stdin
```

## Agent 首读

```bash
dida agent context --json
dida agent context --outline --json
```

这个命令会返回紧凑上下文：项目、文件夹、标签、过滤器、今日任务、未来任务和四象限。
`--outline` 会把任务列表变成 task id 引用，并用去重的 `taskIndex`
承载紧凑任务对象，适合 token 预算更紧的 Agent。

## Schema 发现

```bash
dida schema list --json
dida schema show task.create --json
dida schema show openapi.login --json
dida channel list --json
```

生成写操作前先查 schema。它会告诉 Agent 哪些命令支持 `--dry-run`、`--yes` 和紧凑输出。

## 官方 MCP

官方 MCP 是 token 通道，和 Web API cookie 登录分开。

```bash
DIDA365_TOKEN=... dida official doctor --json
dida official token set --token-stdin --json
dida official token status --json
dida official tools --limit 20 --json
dida official project list --json
dida official project data <project-id> --json
dida official task query --query today --json
dida official task filter --project <project-id> --status 0 --json
dida official task batch-add --args-json '{"tasks":[{"title":"任务"}]}' --dry-run --json
```

## 官方 OpenAPI

官方 OpenAPI 是 OAuth REST 通道，和 Web API / 官方 MCP 都分开。

```bash
dida openapi doctor --json
dida openapi client set --id <client-id> --secret-stdin --json
dida openapi login --browser --json
dida openapi project list --json
dida openapi project create --args-json '{"name":"Project","viewMode":"list","kind":"TASK"}' --dry-run --json
dida openapi focus list --from 2026-04-01T00:00:00+0800 --to 2026-04-02T00:00:00+0800 --type 1 --json
dida openapi habit list --json
dida openapi habit checkin <habit-id> --args-json '{"stamp":20260407,"value":1}' --dry-run --json
```

运行 `dida openapi login --browser --json` 前，先在开发者后台把 OAuth redirect URL
配置成 `dida openapi doctor --json` 输出的 `default_redirect_uri`。

`dida openapi login --browser --json` 会打开浏览器，并在 OAuth 回调完成后只输出一个
最终 JSON envelope。无浏览器的手动流程请用 `dida openapi auth-url --json`
和 `dida openapi listen-callback --json`。
也可以继续使用 `DIDA365_OPENAPI_CLIENT_ID` 和
`DIDA365_OPENAPI_CLIENT_SECRET`；环境变量优先于本地保存的 client 配置。

## 常用读取

```bash
dida project list --json
dida task today --compact --json
dida task upcoming --days 14 --limit 50 --compact --json
dida task search --query "论文" --limit 20 --compact --json
dida completed today --compact --json
dida trash list --cursor 20 --compact --json
dida stats general --json
```

## 常用写入

生成的写操作先预览：

```bash
dida task create --project <project-id> --title "读论文" --dry-run --json
dida task update <task-id> --project <project-id> --title "认真读论文" --dry-run --json
```

确认预览正确后再执行：

```bash
dida task create --project <project-id> --title "读论文" --json
dida task complete <task-id> --project <project-id> --json
```

破坏性操作必须显式加 `--yes`：

```bash
dida task delete <task-id> --project <project-id> --dry-run --json
dida task delete <task-id> --project <project-id> --yes --json
```

## Agent 提示

本节面向 LLM/Agent 使用。优先使用 JSON 输出；写操作前先查
`dida schema list --json`；生成的写入先跑 `--dry-run`；不要要求用户把
cookie 或 token 发到聊天里。
