# DidaCLI — S.U.P.E.R 模块清单

**生成日期**: 2026-06-25

## S.U.P.E.R 总评矩阵

| 包 | S | U | P | E | R | 均分 |
|---|---|---|---|---|---|---|
| cmd/dida | 5 | 5 | 5 | 5 | 5 | **5.0** |
| internal/config | 5 | 5 | 5 | 5 | 5 | **5.0** |
| internal/model | 5 | 5 | 5 | 5 | 4 | **4.8** |
| internal/auth | 4 | 5 | 4 | 5 | 4 | **4.4** |
| internal/officialmcp | 5 | 5 | 3 | 5 | 3 | **4.2** |
| internal/webapi | 5 | 5 | 3 | 4 | 3 | **4.0** |
| internal/openapi | 4 | 5 | 3 | 5 | 3 | **4.0** |
| internal/cli | 3 | 5 | 4 | 4 | 4 | **4.0** |

**项目均分: 4.3 / 5.0**

## Top 10 违规热点 (按严重程度)

| # | 严重度 | 文件 | 问题 |
|---|---|---|---|
| 1 | HIGH | `openapi/client.go` | 错误路径缺少 token 脱敏 |
| 2 | HIGH | `openapi/oauth.go` | RefreshToken 字段存在但未实现自动刷新 |
| 3 | HIGH | `webapi/client.go` | HTTPClient 默认无超时 (http.DefaultClient) |
| 4 | MEDIUM | `cli/openapi_cmd.go`(1660L), `official_cmd.go`(1484L), `account_read_cmd.go`(748L) | 巨型文件混合多职责 |
| 5 | MEDIUM | `webapi/sync.go` + `model/normalize.go` | 类型转换辅助函数重复 |
| 6 | MEDIUM | `auth/`, `model/`, `officialmcp/`, `openapi/` | Token 文件 CRUD 模式 4 次重复 |
| 7 | MEDIUM | `auth/cookie.go` | 凭证写入非原子操作 |
| 8 | MEDIUM | `openapi/oauth.go` | client_secret 明文存储 |
| 9 | LOW | `webapi/client.go` | deviceHeader() 硬编码 "OS X" / "Firefox 95.0" |
| 10 | LOW | `auth/browser.go` | 嵌入 179 行 Python 脚本 |

## 各包详细评估

### internal/cli (39文件)

- **强项**: 单向导入，一致的错误信封，schema 系统
- **弱项**: 手动标志解析在 ~10 个文件中重复，巨型文件混合多职责，无 DI

### internal/webapi (8文件)

- **强项**: 单一职责（每文件一个 API 域），集中化 cookie 认证
- **弱项**: 无 Client 接口（不可 mock），无默认超时

### internal/openapi (4文件)

- **强项**: 环境变量优先级正确
- **弱项**: 无 token 刷新、错误路径无脱敏、明文 client_secret

### internal/auth (3文件)

- **强项**: 路径遍历防护，规范的 token 标准化
- **弱项**: 非原子写入，嵌入 Python 脚本
