# 架构与技术负责人指南

[English](architecture-lead.md) | [简体中文](architecture-lead.zh-CN.md)

> 适合谁：负责这个 Go（Kratos v3）+ TypeScript monorepo 的架构师或技术负责人，需要把控分层边界、代码生成流水线、技术债与评审标准。

## 职责

- 维护分层边界与五条边界规则，在评审中拦截跨层引用。
- 维护生成流水线（`pnpm run generate`）和禁止手改的生成文件清单。
- 维护权限目录契约（`contracts/permissions.yaml`）及其校验机制。
- 跟踪已知不一致与技术债，主动安排偿还，而不是任其漂移。
- 制定评审标准：任何 proto 或权限变更都必须附带重新生成的代码并通过门禁。

## 分层图与依赖方向

摘自[架构文档](../architecture.zh-CN.md#分层)：

```text
apps/web, apps/admin          Vue 3 + Vite + Vuetify 4
        │  HTTP/JSON (protojson)
services/internal/server      Kratos HTTP/gRPC、认证鉴权、OAuth 回调
services/internal/service     proto <-> biz 适配层
services/internal/biz         用例与仓储接口
services/internal/data        GORM 仓储、种子数据、健康检查
services/internal/auth        JWT、密码哈希、OAuth provider 注册表
services/internal/plugin      实验性运行时插件加载器
services/internal/webui       内嵌的 web/admin SPA 处理器
        ▲ 由 buf 生成
contracts/proto/**/v1         protobuf 契约
```

| 层 | 可以引用 | 禁止引用 |
| --- | --- | --- |
| `server` | `service`、`authz`、`middleware` | 直接引用 `data` |
| `service` | `biz`、生成的 `services/api` 类型 | `data`、`server` |
| `biz` | 自己定义的 `data` 接口 | `service`、`server` |
| `data` | GORM、conf | 上层任何包 |
| 扩展 | `biz`、`data`、`server/authz`、`extensions/registry` | —— |

依赖只能向下：`server → service → biz → data`；核心代码不得 import 某个具体扩展。HTTP 中间件顺序为 recovery → logging → LoginRateLimit → Authn → Authz（见 `services/internal/server/http.go`）。

## 五条边界规则

概括自[架构：边界规则](../architecture.zh-CN.md#边界规则)：

1. `contracts/` 只放 `.proto` 文件和 `permissions.yaml`，不放任何实现代码。
2. `services/api/` 全部是生成代码；唯一的手写文件是 `services/api/identity/v1/redact.go`。
3. `packages/api-client/{generated,src/generated}`、`packages/shared/src/permissions.gen.ts` 以及 `services/internal/` 下的 `*_gen` 包均为生成产物，禁止手改。
4. 所有 `dist/` 产物（`apps/*/dist`、`runtime-plugins/*/dist`、`services/internal/webui/dist`）都是构建输出，已 gitignore。
5. 每个构建单元拥有自己的清单（`services/go.mod`、各 `package.json`、`deploy/docker/Dockerfile`），根目录文件只负责编排。

## 单一事实源与生成链

`contracts/proto/**/v1/*.proto` 与 `contracts/permissions.yaml` 是事实源。`pnpm run generate` 产出：

| 来源 | 生成产物 |
| --- | --- |
| `contracts/proto/**/v1` | `services/api/**` 下的 Go protobuf/gRPC/Kratos HTTP 代码 |
| `contracts/proto` | `packages/api-client/generated/openapi.yaml` 与 `packages/api-client/src/generated/schema.d.ts` |
| `services/internal/conf/conf.proto` | `services/internal/conf/conf.pb.go` |
| `contracts/permissions.yaml` | `services/internal/pkg/rbac/permissions_gen.go`、`packages/shared/src/permissions.gen.ts` |
| `services/cmd/api/wire.go` | `wire_gen.go` |
| `extensions/installed.lock.json` | `*_gen` 注册表与 `apps/*/src/router/features.gen.ts` |

禁改清单：`services/api/**`（`identity/v1/redact.go` 除外）、`conf.pb.go`、`permissions_gen.go`、所有 `*_gen` 文件/目录、`packages/api-client/generated/**`、`packages/api-client/src/generated/**`、`apps/*/src/router/features.gen.ts`。CI 会重跑 `generate`，出现任何 `git diff` 即失败。

## API 参考（Swagger UI）

同一个 `protoc-gen-openapi` 生成器会产出两份已提交文档：

- `packages/api-client/generated/openapi.yaml`，供 TypeScript 客户端使用。
- `services/internal/swagger/generated/openapi.yaml`，嵌入 Go 二进制，并在非生产
  环境通过 `/swagger/` 提供 Swagger UI。

两者都属于生成物（见上面的禁改清单）。曝光策略是配置中的 `swagger.enabled` 加上
`APP_ENV=production` 的硬关闭；任何削弱生产门禁的改动都需要评审。详见
[开发指南 · API 参考](../development.zh-CN.md#api-参考与-swagger-ui)。

## 扩展 vs 运行时插件 vs 核心

| | 核心 | 编译期扩展 | 运行时插件（实验性） |
| --- | --- | --- | --- |
| 位置 | `services/internal/`、`apps/` | `extensions/templates/`，由 `scripts/module.mjs` 安装 | `runtime-plugins/<id>/` |
| 生命周期 | 始终编译进二进制 | 安装时复制进仓库，由 `installed.lock.json` 跟踪 | 服务启动时加载 |
| 能否新增 proto/Go 服务 | 能 | 能（`codegen: true` 会重跑生成） | 不能 |
| 能否新增 SPA 路由/菜单 | 能 | 能，通过生成的 `features.gen.ts` | 能，通过 `plugin.json` |
| 鉴权集成 | `authz/mapping.go` | 写入 `authz/extensions_gen.go` | 在 `/proxy/plugins/` 下自行鉴权 |

所有部署都需要的能力（认证、RBAC、用户管理）放核心；可选但一等公民的功能（OAuth provider、垂直切片示例）用扩展；不触碰代码生成与二进制的宿主无关面板或实验用运行时插件。

## 认证/RBAC 的生产语义

- JWT 访问令牌 15 分钟，刷新令牌 7 天且滚动轮换；密码使用 bcrypt；刷新令牌以 SHA-256 哈希存储。
- 登录限流：同一 IP + 用户名每分钟失败 5 次返回 HTTP 429。
- 种子管理员 `admin / admin123` 带 `must_change_password=true`；改密前所有带权限的接口返回 `MUST_CHANGE_PASSWORD`，由中间件强制。生产环境下种子密码未轮换会拒绝启动；首次启动前设置 `SCAFFOLD_SEED_ADMIN_PASSWORD`。
- `is_admin=true` 绕过权限校验；未知的 authz 操作一律 fail closed，返回 403。
- authz 映射（`services/internal/server/authz/mapping.go`）必须覆盖每个生成的操作，`pnpm run verify-authz` 强制校验。
- OAuth 的 `OAuthStore` 是内存实现：state 与一次性 ticket 重启即丢、跨副本不共享；水平扩容前必须换成 Redis 或数据库实现。
- `/proxy/plugins/` 与 `/api/v1/auth/oauth/` 路由不走 Kratos 中间件链，自行完成认证。

## 单二进制双 SPA 部署模型

构建出的 Go 服务同时托管两个前端：`/` 指向 `apps/web/dist`，`/admin/` 指向 `apps/admin/dist`，由 `services/internal/webui` 按最长前缀匹配选择挂载点。`api/`、`proxy/`、`healthz`、`readyz` 为保留前缀，永远不会回退到 SPA。静态资源缓存 1 年，`index.html` 不缓存。线上 JSON 使用 protojson：字段为 camelCase、枚举用名称、64 位整数序列化为字符串。

## 已知不一致

以下为准的官方清单；修复方向是向前收敛，而不是掩盖：

- `packages/shared/src/index.ts` 导出 `APP_NAME = "Vela"`，而仓库其他品牌均为 "scaffold"。
- 两个应用的 `apps/*/src/i18n/index.ts` 都把语言持久化在 `vela.locale` 键下。
- `apps/web/src/api/client.ts` 使用请求头 `x-vela-retried`，而 `apps/admin/src/api/client.ts` 使用 `x-scaffold-retried`。
- `apps/admin/src/directives/permission.ts` 定义了 `vPermission`，但从未在 `main.ts` 中注册。
- `PageHeader` 需要每个视图自行导入，而 `AppSnackbar` 与 `ConfirmDialog` 在两个 `main.ts` 中都是全局注册。
- 前端没有任何测试；每个 package 的 test 脚本都是 `echo no-tests`。
- OAuth store 只有内存实现（见上文）。
- 登录限流器也是内存态，失败次数限制按副本独立生效，不是全局限制。
- `services/internal/data/migrations/` 目前只有 README，尚无迁移工具。
- CI 只构建 Docker 镜像但不推送（`.github/workflows/ci.yml` 的 `docker` job，仅 main 分支）。

## 技术债与演进建议

按影响排序；以下事项均尚未完成：

1. 补齐前端测试（Vitest + 组件测试），替换掉 `echo no-tests` 占位脚本。
2. 提供持久化 `OAuthStore` 实现（Redis 或 SQL），支持多副本部署。
3. 在 `services/internal/data/migrations/` 下引入真正的数据库迁移机制，替代 GORM auto-migrate + seed。
4. 统一品牌残留（`APP_NAME`、`vela.locale`、`x-vela-retried`），并为已存储的 localStorage 键提供迁移说明。
5. 全局注册 `vPermission` 或删除该指令；同时为全局 UI 组件确定一种统一模式。
6. 如确有发布镜像需求，在 tag/release 触发器后增加 CI 的 Docker 推送步骤。

## 变更评审清单

- [ ] 改了 proto → 重跑 `pnpm run generate`，提交生成 diff，生成路径内无手改。
- [ ] 新增权限 → 写入 `contracts/permissions.yaml`，在 `services/internal/server/authz/mapping.go` 中映射，`pnpm run verify-authz` 通过。
- [ ] 检查跨层引用（只允许 `server → service → biz → data`；核心不得 import 扩展）。
- [ ] 生成文件未被手改；CI 的 codegen job 重跑后 `git diff` 为零。
- [ ] `pnpm run typecheck`、`pnpm run lint:web`、`go -C services test ./...` 全部通过。
- [ ] 行为或边界发生变化时，同步更新文档（`docs/architecture.md`、`docs/auth.md` 或本系列）。

## 相关

- [架构](../architecture.zh-CN.md)
- [认证、RBAC 与 SSO](../auth.zh-CN.md)
- [扩展](../extensions.zh-CN.md)
- [运行时插件](../../runtime-plugins/README.zh-CN.md)
