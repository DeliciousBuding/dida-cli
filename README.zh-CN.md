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

## 功能覆盖

DidaCLI 通过 `dida schema list --compact --json` 暴露 139 条本地命令契约。主要功能面如下：

| 范围 | 覆盖内容 | 入口 |
|---|---|---|
| 输出与安全 | 稳定 JSON 信封、紧凑任务输出、本地 schema 发现、写入预览、破坏性确认、Token 脱敏 | `--json`, `--compact`, `--dry-run`, `--yes`, `schema` |
| 认证与诊断 | 浏览器 Cookie 登录、stdin 导入 Cookie、认证验证、退出登录、本地配置和端点检查 | `auth`, `doctor` |
| 通道选择 | Web API、官方 MCP、官方 OpenAPI 三条认证模型相互独立，并提供本地通道选择说明 | `channel list` |
| Agent 上下文 | 一次调用返回项目、文件夹、标签、筛选器、今日、即将到期、四象限分组，支持 outline 模式 | `agent context` |
| 同步 | 全量同步和基于 checkpoint 的增量同步，返回规范化视图和兼容原始增量 | `sync all`, `sync checkpoint` |
| 任务 | 今日、最新捕获任务、活动列表、任务详情、搜索、即将到期、到期计数、创建、更新、完成、删除、移动、父子任务关系 | `+today`, `task` |
| 任务字段 | 正文、富文本描述、开始和截止时间、时区、优先级、标签、清单项、看板列、提醒、重复元数据、全天、浮动任务 | `task create`, `task update` |
| 项目与组织 | 项目列表、项目任务、项目 CRUD、文件夹 CRUD、标签 CRUD/重命名/合并、筛选器、看板列读取和实验性创建 | `project`, `folder`, `tag`, `filter`, `column` |
| 评论与文件 | 评论列表、创建、更新、删除，评论附件上传，附件配额，已有任务附件下载 | `comment`, `attachment` |
| 历史记录 | 已完成任务、关闭历史、回收站删除任务 | `completed`, `closed`, `trash` |
| 效率数据 | 番茄偏好、记录、计时、统计、时间线、任务番茄记录，习惯偏好、习惯、分组、打卡，艾森豪威尔四象限 | `pomo`, `habit`, `quadrant` |
| 账号元数据 | 设置、每日提醒、共享联系人、项目共享状态、日历订阅和账号、账号统计、项目模板、全局搜索、用户资料/状态/会话 | `settings`, `reminder`, `share`, `calendar`, `stats`, `template`, `search`, `user` |
| 官方 MCP | Token 管理、工具发现、schema 展示、原始工具调用、项目读取、任务搜索/筛选/批量写入、习惯读写、专注记录读取/删除 | `official` |
| 官方 OpenAPI | OAuth 客户端配置、浏览器或手动 OAuth、项目/任务/专注/习惯封装、已完成和筛选任务读取 | `openapi` |
| 原始只读探测 | 只允许 GET 的 Web API 探测入口，并返回结构化错误细节 | `raw get` |
| 分发与升级 | 单二进制构建、npm/安装脚本、包管理模板、发布归档、校验和验证自升级 | `upgrade` |

## 命令示例

<details>
<summary>Web API 读取</summary>

```bash
dida task today --json
dida task latest --limit 10 --project inbox --compact --json
dida task list --filter all --limit 50 --compact --json
dida task upcoming --days 14 --json
dida task search --query "考试" --json
dida task due-counts --json
dida project list --json
dida project tasks <project-id> --limit 50 --compact --json
dida folder list --json
dida tag list --json
dida filter list --json
dida column list <project-id> --json
dida completed today --json
dida closed list --status 2 --limit 50 --json
dida trash list --cursor 20 --compact --json
dida settings get --include-web --json
dida reminder daily --json
dida attachment quota --json
dida share contacts --json
dida share project shares <project-id> --json
dida calendar subscriptions --json
dida stats general --json
dida template project list --limit 50 --json
dida search all --query "考试" --limit 20 --json
dida user profile --json
dida pomo stats --json
dida habit list --json
dida quadrant list --json
dida sync all --json
```
</details>

### 最新捕获任务

从微信或其他入口连续推送多个相关任务，并把上下文和材料一起放进收集箱时，优先用
`task latest` 读取最近捕获的任务：

```bash
dida task latest --limit 10 --project inbox --compact --json
```

这个命令按创建时间倒序返回活动任务，`--project inbox` 会映射到同步数据里的真实收集箱项目。
如果创建时间缺失，会回退到修改时间。去掉 `--project inbox` 可以读取所有项目里的最新活动任务。

<details>
<summary>Web API 写入</summary>

```bash
dida task create --project <id> --title "新任务" --dry-run --json
dida task update <id> --project <p> --title "更新" --dry-run --json
dida task complete <id> --project <p> --dry-run --json
dida task move <id> --from <p> --to <p> --dry-run --json
dida task parent <id> --parent <parent-id> --project <p> --dry-run --json
dida task delete <id> --project <p> --dry-run --yes --json
dida project create --name "新项目" --dry-run --json
dida folder create --name "工作" --dry-run --json
dida tag rename old-name new-name --dry-run --json
dida column create --project <p> --name "进行中" --dry-run --json
dida comment create --project <p> --task <id> --text "看起来没问题" --dry-run --json
dida comment create --project <p> --task <id> --text "见附件" --file ./probe.png --dry-run --json
dida attachment download --project <p> --task <id> --attachment <a> --output ./file.doc --json
```
</details>

<details>
<summary>官方 MCP 和 OpenAPI</summary>

```bash
# MCP（Token）
DIDA365_TOKEN=dp_xxx dida official project list --json
dida official tools --limit 20 --json
dida official task filter --project <id> --status 0 --json
dida official task batch-add --args-json '{"tasks":[{"title":"Task"}]}' --dry-run --json
dida official habit list --json
dida official focus list --from-time 2026-05-01T00:00:00+08:00 --to-time 2026-05-09T23:59:59+08:00 --type 1 --json

# OpenAPI（OAuth）
dida openapi client set --id <id> --secret-stdin --json
dida openapi login --browser --json
dida openapi project list --json
dida openapi task create --args-json '{"projectId":"<project-id>","title":"Task"}' --dry-run --json
dida openapi task filter --args-json '{"projectIds":["<project-id>"],"status":[0]}' --json
dida openapi habit checkins --habit-ids <habit-id> --from 20260401 --to 20260407 --json
```
</details>

<details>
<summary>Agent、schema、原始读取和升级</summary>

```bash
dida schema list --compact --json
dida schema show task.create --json
dida channel list --json
dida agent context --outline --json
dida raw get /user/preferences/settings --json
dida upgrade --check --json
```
</details>

完整参考：[docs/commands.md](docs/commands.md)。机器可读参考：`dida schema list --compact --json`。

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
