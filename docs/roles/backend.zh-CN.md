# 后端开发指南

[English](backend.md) | [简体中文](backend.zh-CN.md)

> 适合谁：在 `services/` 下开发 Go（Kratos v3）后端的同学，涵盖 protobuf 契约、业务逻辑、持久化与鉴权。

## 职责与边界

后端遵循严格的单向依赖链：

```text
services/internal/server  ->  service  ->  biz  ->  data
```

- `server`：Kratos HTTP/gRPC 传输层、中间件、认证鉴权、OAuth 回调。
- `service`：薄薄的适配层，负责 proto 消息与 biz 用例之间的转换。
- `biz`：用例与仓储接口（不涉及存储细节）。
- `data`：GORM 仓储实现、种子数据、健康检查。

硬性规则：

- 依赖只能向下，`biz`/`data` 不得 import `server` 或 `service`。
- 核心代码不得 import 任何具体扩展；扩展可以依赖 `biz`、`data`、`server/authz` 和 `extensions/registry`，反之不行。
- `contracts/` 只放 `.proto` 文件和 `permissions.yaml`，不放实现代码。

## 环境与运行

```bash
pnpm run init        # 一次性：安装锁定的代码生成工具与工作区依赖
pnpm run generate    # 重新生成所有契约产物
pnpm run dev:server  # HTTP :8000，gRPC :9000
```

- 入口：`services/cmd/api/main.go`（Wire 依赖注入见 `services/cmd/api/wire.go`）。
- 配置：`configs/config.local.yaml`；环境变量前缀 `SCAFFOLD_`，占位符写法 `${NAME:default}`。
- 默认值：HTTP `0.0.0.0:8000`，gRPC `0.0.0.0:9000`，SQLite 位于 `file:../data/scaffold.db`，已开启 WAL 与外键。
- 切到 MySQL 只需改配置（`SCAFFOLD_DB_DRIVER=mysql`、`SCAFFOLD_DB_SOURCE=...`），无需改代码。
- 种子管理员为 `admin / admin123`，首次登录强制改密码。

## 目录地图

| 路径 | 内容 |
| --- | --- |
| `services/cmd/api` | 二进制入口、Wire 装配 |
| `services/internal/server` | HTTP/gRPC 服务、编解码、中间件、`authz/mapping.go` |
| `services/internal/service` | proto <-> biz 适配器、`Services` 聚合 |
| `services/internal/biz` | 用例、仓储接口、领域错误 |
| `services/internal/data` | GORM 仓储（`database/`）、模型（`model/`）、迁移钩子 |
| `services/internal/auth` | JWT/密码工具、OAuth（`oauth/`） |
| `services/internal/plugin` | 运行时插件加载器与注册表 |
| `services/internal/webui` | 内嵌 SPA 处理器（`/`、`/admin/`） |
| `services/internal/pkg` | 公共助手：`auth`、`rbac`、`id`、`pagination` |
| `contracts/proto/**/v1` | protobuf 唯一事实来源 |

## 请求生命周期

1. HTTP 中间件按以下顺序执行（`services/internal/server/http.go`）：
   `recovery -> logging -> LoginRateLimit -> Authn -> Authz`。
2. 请求体使用 protojson 解码（`services/internal/server/codec.go`）：
   字段名为 camelCase，枚举用名称而非数字，64 位整数序列化为字符串，
   未知字段丢弃，响应使用 `EmitUnpopulated`。
3. `Authn` 校验 JWT 访问令牌；`Authz` 用生成的操作名查 `authz/mapping.go`，
   未登记的操作一律拒绝（fail-closed）。
4. `/proxy/plugins/` 与 `/api/v1/auth/oauth/` 通过 `HandlePrefix` 注册，
   绕过 Kratos 中间件链，自行完成认证。

## 新增后端模块完整链路

1. 在 `contracts/proto/<domain>/v1/*.proto` 中定义契约，并加 `google.api.http` 注解。
2. 运行 `pnpm run generate`，刷新 `services/api/**`、OpenAPI 与 TS schema。
3. 在 `services/internal/biz/<domain>.go` 实现用例，并在其中声明仓储接口。
4. 在 `services/internal/data/database/<domain>.go` 实现仓储，对应模型放在 `services/internal/data/model/`。
5. 在 `services/internal/service/<domain>.go` 添加传输层适配器。
6. 在 `services/internal/service/service.go`（`Services` 结构体与 `ProviderSet`）以及 `services/internal/server/http.go` / `grpc.go` 中注册服务。
7. 把每个新操作加入 `services/internal/server/authz/mapping.go`（`Public`、仅登录、或某个 `rbac.Perm*` 编码）。
8. 如需新权限编码，先加到 `contracts/permissions.yaml` 再重新生成。
9. 运行 `pnpm run verify-authz`；操作未登记或权限编码不存在都会报错。
10. 在 `services/` 下运行 `go vet ./...` 和 `go test ./...`。

## API 参考（Swagger UI）

`pnpm run generate` 也会生成嵌入后端的 OpenAPI 文档。非生产环境下服务端通过
`/swagger/` 提供：

- UI：`http://localhost:8000/swagger/`
- Spec：`http://localhost:8000/swagger/openapi.yaml`

带 `google.api.http` 注解的新 RPC 在重新生成后会出现在这里。生产环境会强制关闭该
端点，可用 `SCAFFOLD_SWAGGER_ENABLED` 控制；详见
[开发指南 · API 参考](../development.zh-CN.md#api-参考与-swagger-ui)。

## 数据与迁移

- 当 `data.database.auto_migrate` 为 true 时，启动时执行 GORM `AutoMigrate`（见 `services/internal/data/database/data.go`）；新的持久化类型要加进那里的 `AutoMigrate` 调用。
- SQLite 连接池被限制为单写者，避免 busy 错误。
- `services/internal/data/migrations/` 刻意保持为空；启动钩子会读取该目录，发现有 `.sql` 文件但迁移器未启用时给出警告。未来 `golang-migrate` 的目录约定（`NNNN_name.up.sql` / `NNNN_name.down.sql`）见 `services/internal/data/migrations/README.md`。
- 种子逻辑（`data.database.seed`）创建默认管理员、角色与权限绑定。

## 认证与 RBAC

- JWT 访问令牌有效期 15 分钟；刷新令牌 7 天，且以哈希形式存储。密码使用 bcrypt。
- 登录限流：同一 IP+用户名每分钟失败 5 次即返回 429。
- 权限目录为 `contracts/permissions.yaml`，生成到 `services/internal/pkg/rbac/permissions_gen.go`（Go）与 `packages/shared/src/permissions.gen.ts`（TS）。
- `is_admin` 绕过所有权限校验；其余主体一律走 Authz 映射，未知操作 fail-closed。
- 种子管理员必须先改密码才能使用带权限的接口；生产环境（`APP_ENV=production`）对默认 JWT/OAuth 密钥和未轮换的种子密码直接启动失败。

## 不可手改清单

以下均为生成产物，请改源头后重新生成，不要手改：

- `services/api/**`（唯一例外：`services/api/identity/v1/redact.go`）
- `services/internal/conf/conf.pb.go`
- `services/internal/pkg/rbac/permissions_gen.go`
- `services/internal/auth/oauthproviders_gen/`
- `services/internal/server/authz/extensions_gen.go`
- `services/internal/extensions/registry_gen/`
- `services/cmd/api/wire_gen.go`
- `packages/api-client/generated/**`、`packages/api-client/src/generated/**`
- `packages/shared/src/permissions.gen.ts`、`apps/*/src/router/features.gen.ts`

## 测试与质量门禁

- 测试只存在于 Go 侧，例如 `services/internal/auth/oauth/oauth_test.go`、`services/internal/pkg/rbac/permissions_consistency_test.go`、`services/internal/server/middleware/authz_test.go`、`services/internal/data/database/data_test.go`、`services/internal/webui/fallback_test.go`、`services/internal/plugin/registry_test.go`。
- 本地门禁：在 `services/` 下运行 `go vet ./...` 与 `go test ./...`。CI 运行的是 `go test -race ./...`，涉及并发的改动提交前请在本地加 `-race` 跑一遍。
- CI（`.github/workflows/ci.yml`）：`codegen`（generate + verify-authz + 漂移检查）、`go`（vet + test -race）、`web`、`extensions`、`docker`。

## 常见坑

- 忘记 authz 映射：新 RPC 没有在 `authz/mapping.go` 登记会被 fail-closed 拒绝，表现像权限 bug。
- protojson 线格式：客户端收发都是 camelCase，int64 是字符串；不要在 proto 类型旁边手搓 `encoding/json` 的 DTO。
- 插件代理与 OAuth 路由不走 Kratos 中间件链；不要假设 `/proxy/plugins/` 或 `/api/v1/auth/oauth/` 下的处理器已经过 `Authn`。
- 手改生成文件：CI 漂移检查会失败；应修改 `.proto` 或 `permissions.yaml` 后重新生成。
- SQLite 写竞争：写操作保持短小；连接池已固定为单连接。

## 相关

- [架构总览](../architecture.zh-CN.md)
- [认证与授权](../auth.zh-CN.md)
- [开发流程](../development.zh-CN.md)
- [架构负责人指南](architecture-lead.zh-CN.md)
