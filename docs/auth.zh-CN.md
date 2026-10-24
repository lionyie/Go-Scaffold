# 认证、RBAC 与 SSO

[English](auth.md) | [简体中文](auth.zh-CN.md)

## 本地登录

`POST /api/v1/auth/login` 使用 bcrypt 校验密码，返回 access token（15 分钟）和
轮换的 refresh token（7 天）。令牌是使用 `SCAFFOLD_JWT_SECRET` 签名的 JWT；
refresh token 只以 SHA-256 哈希落库。

内置管理员为 `admin / admin123`，其 `must_change_password=true`。在
`POST /api/v1/auth/password` 成功之前，`GET /api/v1/auth/me`、登出等仅要求登录
态的接口保持可用，而所有需要权限的接口返回 `MUST_CHANGE_PASSWORD`。这样管理端
才能渲染出持久的改密弹窗——限制由 authz 中间件强制，而不只是前端行为。改密成功
后，SPA 会先轮换 access token 再进入仪表盘。

登录失败有内存限流：同一 IP + 用户名每分钟失败 5 次即返回 HTTP 429。

## RBAC 模型

```text
users(id, username, email, password_hash, is_admin, must_change_password, status, ...)
roles(id, name, display_name, description, is_system)
permissions(code, name, resource, action)
user_roles(user_id, role_id)
role_permissions(role_id, permission_code)
oauth_accounts(id, user_id, provider, provider_user_id, email, email_verified, ...)
refresh_tokens(id, user_id, token_hash, session_id, expires_at, revoked_at)
```

- 权限目录在 `contracts/permissions.yaml`，生成到 Go（`rbac.Catalog`、
  `rbac.All`）和 TypeScript（`Permission`）。
- `services/internal/server/authz/mapping.go` 将每个生成的 operation 映射为
  `Public`、空 code（仅要求登录态）或某个权限码。
- 未登记的 operation 默认拒绝（403）。
- `is_admin=true` 绕过权限检查。
- `pnpm run verify-authz` 会在生成的 operation 缺少映射、映射条目过期、或权限
  码不存在时失败。

默认角色：`admin`（全部权限）和 `viewer`（只读）。管理端的路由与菜单守卫使用
生成的 `Permission` 常量，ESLint 会拒绝 `apps/admin/src` 下出现权限字符串字面量。

## 第三方登录

OAuth 与具体 provider 解耦。核心端点始终存在；provider 是在各自 `init()` 中
注册的扩展。

| 端点 | 用途 |
| --- | --- |
| `GET /api/v1/auth/oauth/providers` | 已启用 provider 的公开列表 |
| `POST /api/v1/auth/oauth/{provider}/start` | 返回 `{authUrl}` |
| `GET /api/v1/auth/oauth/{provider}/callback` | Provider 重定向；换取 code 并跳回管理端 |
| `POST /api/v1/auth/oauth/exchange` | 用一次性 ticket 换取令牌 |
| `GET /api/v1/auth/oauth/accounts` | 列出调用者已绑定的账号 |
| `DELETE /api/v1/auth/oauth/{provider}` | 解绑某个 provider |

流程：

1. 管理端登录页加载已启用的 provider。
2. `start` 生成随机 state，用 `SCAFFOLD_OAUTH_STATE_SECRET` 签名，创建 PKCE
   verifier/challenge 与 nonce，存入 `OAuthStore`，并返回 provider 授权 URL。
3. 回调校验 state 签名，消费它（单次有效），换取 code，并归一化用户资料。
4. 已绑定的身份直接登录。仅当 `email_verified=true` 时才会用已验证邮箱自动
   绑定；否则只有 `auth.oauth.auto_register=true` 才会创建账号。
5. 令牌放入一次性、60 秒有效的 ticket。回调重定向到
   `SCAFFOLD_OAUTH_ADMIN_REDIRECT_URL` 并带 `?ticket=...`，由 SPA 换取令牌。
   令牌永远不会出现在 URL 中。
6. 回调设置 `Referrer-Policy: no-referrer`。

### Provider 环境变量

```text
SCAFFOLD_OAUTH_GITHUB_CLIENT_ID / _CLIENT_SECRET
SCAFFOLD_OAUTH_GOOGLE_CLIENT_ID / _CLIENT_SECRET
SCAFFOLD_OAUTH_WECHAT_APP_ID / _APP_SECRET
SCAFFOLD_OAUTH_QQ_APP_ID / _APP_KEY
SCAFFOLD_OAUTH_OIDC_ISSUER / _CLIENT_ID / _CLIENT_SECRET / _SCOPES
```

在各 provider 控制台登记回调地址：

```text
http(s)://<host>/api/v1/auth/oauth/<provider>/callback
```

微信（开放平台扫码）与 QQ 互联不返回邮箱地址，它们永远不参与基于邮箱的绑定。
OIDC 会用 issuer 的 JWKS 校验 ID token 签名，并检查 `iss`、`aud`、`exp` 与
nonce。

### 多副本部署

`OAuthStore` 自带单副本的内存实现。state 与 ticket 在重启后丢失，也不会跨
Pod 共享。水平扩容前请替换为 Redis 或数据库实现。

## go-admin 参考

RBAC 模型与管理端交互借鉴了
[go-admin](https://github.com/go-admin-team/go-admin)，但本脚手架使用简单的
角色-权限表而非 Casbin，也不复制 go-admin 代码。`extensions/upstream.json`
固定了参考 commit，`pnpm run module:check-upstream` 会报告上游是否移动。
