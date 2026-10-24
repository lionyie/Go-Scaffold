# 开发指南

[English](development.md) | [简体中文](development.zh-CN.md)

## 首次运行

```bash
pnpm run init
pnpm run generate
```

`init` 安装固定版本的代码生成工具与 workspace 依赖，不需要 `make`。

## 日常开发

开三个终端：

```bash
pnpm run dev:server   # :8000 / :9000
pnpm run dev:web      # http://localhost:5173
pnpm run dev:admin    # http://localhost:5175/admin/
```

内置管理员为 `admin / admin123`。首次登录会打开一个不可关闭的改密弹窗；完成前，
仅登录态可用的接口（`/auth/me`、登出、修改密码）可以访问，需要权限的接口会返回
`MUST_CHANGE_PASSWORD`。

## 新增后端模块

1. 在 `contracts/proto/<domain>/v1/*.proto` 中定义契约，并带上
   `google.api.http` 注解。
2. 运行 `pnpm run generate`。
3. 在 `services/internal/biz/<domain>.go` 实现用例，在
   `services/internal/data/database/<domain>.go` 实现仓储。
4. 在 `services/internal/service/<domain>.go` 添加传输适配器，并在
   `services/internal/service/service.go` 与
   `services/internal/server/http.go` / `grpc.go` 中注册。
5. 在 `services/internal/server/authz/mapping.go` 添加 operation 到权限的映射。
   `pnpm run verify-authz` 会在缺少映射或权限码不存在时失败。
6. 在 `apps/admin/src/views` 下构建页面，并使用 `@scaffold/shared` 的
   `Permission` 常量添加路由与菜单项。

## 代码生成

`pnpm run generate`：

1. 清空代码生成产物（保留 `services/api/identity/v1/redact.go`），
2. 重新生成权限常量，
3. 运行三个 buf 模板，
4. 重新生成 TypeScript schema，
5. 重新生成 wire，
6. 运行 `go mod tidy`，
7. 运行 `verify-authz`。

流水线必须是幂等的：跑两次之后文件不应有变化。CI 用 `git diff --exit-code`
强制这一点。

## 扩展与示例

```bash
pnpm run module:add -- oauth-github
pnpm run module:add -- example-notes
pnpm run module:list
pnpm run module:remove -- example-notes
```

详见 [extensions.zh-CN.md](extensions.zh-CN.md)。安装或删除扩展会重写生成的
注册表；当扩展声明 `"codegen": true` 时还会自动执行代码生成。

## 数据库

本地环境使用仓库根目录下 `data/scaffold.db` 的 SQLite。删除了列的 schema 变更
需要重置：

```bash
pnpm run db:reset -- --yes
```

`services/internal/data/migrations/` 会在 `AutoMigrate` 之前被读取。当前为空；
未来接入 `golang-migrate` 的命名规范见该目录的 README。

## 运行时插件开发

```bash
pnpm run build:plugins
pnpm run dev:server
pnpm run dev:admin        # 侧边栏 → Hello plugin
```

需要热更新时，用 `SCAFFOLD_PLUGINS_DEV=true` 启动后端，并运行插件 Vite 开发服务器：

```bash
SCAFFOLD_PLUGINS_DEV=true pnpm run dev:server
pnpm --filter @scaffold/plugin-hello dev
```

清单、路径约定、后端模式、鉴权与排错详见
[runtime-plugins/README.zh-CN.md](../runtime-plugins/README.zh-CN.md)。

## 测试与 lint

```bash
pnpm run verify-authz
pnpm run verify-generated
cd services && go vet ./... && go test ./...
pnpm run typecheck
pnpm run lint:web
buf lint
```

`go test -race` 需要 cgo。没有 C 编译器的机器请去掉 `-race` 运行；CI 在
Linux 上会带 `-race` 执行。

## 生产环境注意事项

- 将 `SCAFFOLD_JWT_SECRET` 与 `SCAFFOLD_OAUTH_STATE_SECRET` 设置为非默认值，
  否则生产启动会 fail fast。
- 首次生产启动前设置 `SCAFFOLD_SEED_ADMIN_PASSWORD`，或交互式修改 seed 密码。
  只要 seed 密码仍待轮换，生产环境就拒绝启动。
- 内存版 OAuth store 只支持单副本。水平扩容前请替换为 Redis 或数据库实现。
