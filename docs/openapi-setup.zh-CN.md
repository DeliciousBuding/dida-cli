# OpenAPI 设置指南

这份文档是 DidaCLI 官方 Dida365 OpenAPI 通道的正式设置指南。

适用场景：

- 在 Dida365 开发者后台创建或配置应用
- 把 `client_id` 和 `client_secret` 保存到本机供 DidaCLI 使用
- 正确配置 OAuth `redirect_uri`
- 完成一次 OAuth 登录，并在后续复用本地保存的 access token

这份文档不适用于：

- Web API cookie 通道
- 官方 MCP 的 `DIDA365_TOKEN` 通道

## 你需要准备什么

开始前请确认你有：

- 一个 Dida365 开发者应用
- 对应的 `client_id`
- 对应的 `client_secret`
- 能接收本地回调的当前机器

DidaCLI 默认本地回调地址是：

```text
http://127.0.0.1:17890/callback
```

除非你明确改了本地端口或主机，否则就用这条完整地址。

## 开发者后台怎么填

在 Dida365 开发者后台：

1. 打开你的应用设置
2. 找到 OAuth redirect URL / callback URL 配置项
3. 填入下面这条完整地址：

```text
http://127.0.0.1:17890/callback
```

注意：

- 默认用的是 `127.0.0.1`，不是 `localhost`
- 必须带 `/callback`
- 必须带端口 `17890`
- 不能只填域名或主机
- 如果你本地改了端口，后台里也必须同步改成完全一致的 URL

## 在 CLI 里保存 ID 和 Secret

`client_secret` 建议只走 stdin，不要进入 shell 历史：

```bash
dida openapi client set --id <client-id> --secret-stdin --json
```

保存后先检查本地状态：

```bash
dida openapi client status --json
dida openapi doctor --json
```

重点看 `dida openapi doctor --json` 输出里的：

- `default_redirect_uri`
- `default_scope`
- `next`

如果 `default_redirect_uri` 和你后台填的不一致，先不要继续，先改配置。

## 两种登录方式

### 方式 1：自动打开浏览器

适合本机 GUI/默认浏览器正常工作的情况：

```bash
dida openapi login --browser --json
```

它会自动做这些事：

- 启动本地 callback 监听
- 尝试打开浏览器
- 等待 OAuth 回调
- 自动交换 `code`
- 把 access token 保存到本地

如果自动打开浏览器失败，JSON 错误里会带：

- `error.details.authorization_url`
- `error.details.redirect_uri`
- `error.details.next`

### 方式 2：手动复制链接

这是最稳的方式，推荐在浏览器自动弹不出来或远程桌面环境下使用。

第 1 步：

```bash
dida openapi listen-callback --json
```

这个终端先不要关。

第 2 步：

```bash
dida openapi auth-url --json
```

复制输出里的 `authorization_url`，在浏览器里打开并完成登录授权。

第 3 步：

浏览器跳回后，第一个终端会输出：

- `code`
- `state`
- `redirect_uri`

第 4 步：

```bash
dida openapi exchange-code --code <code> --json
```

## 如何确认已经成功

登录完成后先查 token：

```bash
dida openapi status --json
```

再做一次真实 API 读取：

```bash
dida openapi project list --json
```

如果 `project list` 能成功返回，这台机器上的 OpenAPI 通道就已经打通了。

## 常见报错

### `invalid_request`: redirect URI must be registered

含义：

- 开发者后台没有注册匹配的 redirect URL

解决：

- 把 `dida openapi doctor --json` 输出的 `default_redirect_uri`
  一字不差地填到开发者后台

### `unsupported_response_type`

含义：

- 你手动拼了一个不完整的授权 URL

解决：

- 不要手拼
- 用 `dida openapi auth-url --json`
  或 `dida openapi login --browser --json`

### 浏览器打开 `127.0.0.1` 回调时报 `ERR_CONNECTION_REFUSED`

含义：

- 浏览器到了 callback 地址，但本地没有监听器在运行

解决：

- 先开 `dida openapi listen-callback --json`
- 再打开一条新生成的 `authorization_url`
- 不要复用旧的 callback URL

### `exchange-code` 显示保存成功，但后续读取说 token 不存在

含义：

- 多半是测试时混用了不同进程或不同环境

解决：

- 优先用安装在 PATH 上的 `dida` 做后续验证：

```bash
dida openapi status --json
dida openapi project list --json
```

## 安全注意事项

- 不要把 `client_secret`、access token、refresh token 发到聊天、issue 或仓库里
- 不要把本地 token 文件提交进仓库
- OpenAPI 必须和下面两条通道分开：
  - Web API cookie
  - 官方 MCP `DIDA365_TOKEN`

## 最短可用路径

如果你只想最快打通一遍，按这个顺序走：

```bash
dida openapi client set --id <client-id> --secret-stdin --json
dida openapi doctor --json
dida openapi listen-callback --json
dida openapi auth-url --json
```

然后：

1. 把 `authorization_url` 复制到浏览器
2. 完成授权
3. 用返回的 `code` 执行 `exchange-code`
4. 最后用 `dida openapi project list --json` 验证
