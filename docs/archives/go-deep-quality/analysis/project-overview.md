# DidaCLI — 架构概览

**生成日期**: 2026-06-25
**仓库**: D:\Code\Projects\tools\DidaCLI
**Go 版本**: 1.26.4
**外部依赖**: 零（仅标准库）

## 技术栈

| 层 | 详情 |
|---|---|
| 语言 | Go 1.26.4 |
| CLI 框架 | 无 — 手动标志解析 |
| HTTP 客户端 | net/http |
| 认证 | Cookie `t` / Bearer token / OAuth |
| 测试 | go test（无外部框架） |
| CI/CD | GitHub Actions (ubuntu-latest) |
| 分发 | GitHub Releases + npm + install.sh/ps1 |

## 目录结构

```
cmd/dida/main.go         (13行) 入口点
internal/
  cli/                   (39文件) 命令路由、标志解析、输出格式化 — 最大包
  webapi/                (8文件)  Web API HTTP 客户端 (Cookie认证)
  officialmcp/           (2文件)  MCP JSON-RPC 客户端 (Bearer)
  openapi/               (4文件)  OAuth + REST 客户端
  auth/                  (3文件)  本地凭证存储 + 浏览器登录
  config/                (2文件)  配置目录解析
  model/                 (2文件)  数据规范化
docs/                    (16文件) 用户和参考文档
```

## 命令架构

28 个顶级命令通过 `rootCommands()` 注册表分派。三通道隔离：

| 通道 | 命令 | 认证 |
|---|---|---|
| Web API | task, project, folder, tag, comment, sync, ... | Cookie `t` |
| Official MCP | official * | DIDA365_TOKEN |
| OpenAPI | openapi * | OAuth |

命令执行模式：`executeRead` / `executeMutation` (Web API) 或每个通道手动管理。

## 数据流

```
CLI → auth/webapi|openapi|officialmcp → HTTP
```
严格单向，无循环导入。
