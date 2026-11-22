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

### 权限来源

| 来源 | 产生方式 | 可删除 | 可改 code |
| --- | --- | --- | --- |
| `builtin` | `contracts/permissions.yaml` | 否 | 否 |
| `plugin` | 插件清单 | 否 | 否 |
| `custom` | `POST /api/v1/permissions`（API） | 是 | 否 |

启动时的种子只同步 builtin 目录并清理过期的 builtin 行，因此 custom 与 plugin
权限不会在重启后丢失。删除 custom 权限时，会在同一事务里清理对应的
`role_permissions`。

自定义权限本身只是目录条目：在某个 operation 被映射到它之前（写在
`Requirements` 里，或通过下面的运行时绑定），它不会保护任何接口。

### 运行时接口绑定

`mapping.go` 里的编译期映射是默认值。后端
`PUT /api/v1/operations/permission` 可以在运行时覆盖某个 operation 的权限。
管理端不再提供该页面，它是运维 API。

- `services/internal/pkg/operations/operations_gen.go` 由 OpenAPI 规格自动生成，
  保证目录与真实路由一致。
- 覆盖关系存放在 `operation_permissions(operation, permission_code)`。
- 公开接口和仅登录接口是锁定的，不能改绑。
- 其余接口可以绑定任意已存在的权限码，包括自定义权限。自定义权限正是通过这一步
  才真正开始保护接口。
- 绑定结果缓存 30 秒，写入时主动失效。
- 删除自定义权限时会一并删除引用它的接口绑定，相关接口回退到编译期映射。
- 改绑需要 `permission:assign`，查看目录需要 `permission:list`。

### 会话角色

会话 token 携带调用者的角色（`roles` claim），鉴权直接读取这些角色，不再每次
请求查数据库。

- 本地密码登录从本地 RBAC 表解析角色
- OIDC 登录直接使用 `realm_access.roles`。对接 Keycloak 时，名称等于权限码的
  realm role（例如 `user:list`）即授予该权限
- 名为 `admin` 的角色等价于 `is_admin`（完全绕过）
- `/api/v1/auth/me` 返回会话中的角色

本地密码登录作为兜底保留。管理端的权限管理页面已移除：SSO 用户的权限在 Keycloak
中管理，密码用户的权限在本地角色页管理。

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

### Keycloak（OIDC）

`pnpm run module:add oauth-oidc` 会安装通用 OIDC provider，目标文件为
`services/internal/auth/oauth/providers/oidc.go`。开发用 realm 位于
`deploy/keycloak/scaffold-realm.json`。

启动 Keycloak（可选 override）：

```bash
docker compose -f deploy/compose/docker-compose.yml \
               -f deploy/compose/docker-compose.keycloak.yml up -d keycloak
```

它监听 `http://localhost:8080`，导入 realm `scaffold`，并创建：

- 客户端 `scaffold-admin`（confidential，secret 为 `dev-scaffold-admin-secret`，
  启用 PKCE S256）
- `http://localhost:8000` 与 `http://127.0.0.1:8000` 的回调地址
- 开发用户 `admin@example.com / admin123`，realm 角色 `admin`
- 开发用户 `viewer@example.com / viewer123`，composite 角色 `viewer`
- 每个权限码对应一个 realm role（例如 `user:list`）

realm role 就是权限码。ID token 里的 `realm_access.roles` 会写入会话，因此在
Keycloak 中分配 `user:list` 即授予 `user:list` 权限。`admin` realm role 等价于
`is_admin`。

后端配置：

```text
SCAFFOLD_OAUTH_STATE_SECRET=<随机值>
SCAFFOLD_OAUTH_OIDC_ISSUER=http://localhost:8080/realms/scaffold
SCAFFOLD_OAUTH_OIDC_CLIENT_ID=scaffold-admin
SCAFFOLD_OAUTH_OIDC_CLIENT_SECRET=dev-scaffold-admin-secret
SCAFFOLD_OAUTH_OIDC_SCOPES=openid email profile
SCAFFOLD_OAUTH_OIDC_DISPLAY_NAME=Keycloak
SCAFFOLD_OAUTH_AUTO_REGISTER=true
```

账号关联规则：

- 已验证邮箱会把 Keycloak 身份关联到同邮箱的本地用户
- 种子管理员默认使用 `admin@example.com`，可用 `SCAFFOLD_SEED_ADMIN_EMAIL` 覆盖
- `auto_register=true` 时，未知身份会创建本地用户，但不具备管理员权限
- 本地密码登录与 SSO 并存

安全说明：

- PKCE S256 与 OIDC nonce 都会校验
- 所有环境都必须设置 `SCAFFOLD_OAUTH_STATE_SECRET`
- 完整 compose 部署时，server 容器通过
  `extra_hosts: localhost:host-gateway` 访问宿主机发布的 Keycloak 端口

Keycloak 是可选的 provider：本地 RBAC 与密码登录仍是默认形态，脚手架不依赖
Keycloak 才能运行。

### 多副本部署

`OAuthStore` 自带单副本的内存实现。state 与 ticket 在重启后丢失，也不会跨
Pod 共享。水平扩容前请替换为 Redis 或数据库实现。

## go-admin 参考

RBAC 模型与管理端交互借鉴了
[go-admin](https://github.com/go-admin-team/go-admin)，但本脚手架使用简单的
角色-权限表而非 Casbin，也不复制 go-admin 代码。`extensions/upstream.json`
固定了参考 commit，`pnpm run module:check-upstream` 会报告上游是否移动。

## 相关

- 指南：[架构](architecture.zh-CN.md) · [扩展机制](extensions.zh-CN.md) · [开发指南](development.zh-CN.md)
- 角色：[后端开发](roles/backend.zh-CN.md) · [运维与 SRE](roles/devops.zh-CN.md) · [测试与 QA](roles/testing-qa.zh-CN.md)
- 全部角色：[角色指南](roles/README.zh-CN.md)
