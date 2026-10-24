# 架构

[English](architecture.md) | [简体中文](architecture.zh-CN.md)

## 目标

- 一个仓库可以承载多个可运行端（`apps/*`）和多个后端服务（`services/cmd/*`）。
- 契约是唯一真源；生成代码提交到仓库，并由 CI 校验漂移。
- Admin 与 RBAC 是脚手架的默认组成部分。
- 可选能力以扩展形式存在，而不是无条件的依赖。

## 分层

```text
apps/web, apps/admin          Vue 3 + Vite + Vuetify 4
        │
        │  HTTP/JSON (protojson)
        ▼
services/internal/server      Kratos HTTP/gRPC、authn/authz、OAuth 回调
services/internal/service     proto <-> biz 适配
services/internal/biz         用例、仓储接口
services/internal/data        GORM 仓储、seed、健康检查
services/internal/auth        JWT、密码哈希、OAuth provider 注册表
services/internal/plugin      实验性运行时插件加载器
services/internal/webui       内嵌 web + admin SPA 处理器
        ▲
        │ 由 buf 生成
contracts/proto/**/v1         protobuf 契约
```

依赖只向下指向：`server → service → biz → data`。扩展可以依赖 `biz`、`data`、
`server/authz` 和 `extensions/registry`，但核心代码不得导入任何具体扩展。

## 边界规则

1. `contracts/` 只包含 `.proto` 文件和 `permissions.yaml`，不放任何 Go 或
   TypeScript 实现。
2. `services/api/` 是生成区。唯一的手写文件是
   `services/api/identity/v1/redact.go`，它必须与生成的请求类型同包。
3. `packages/api-client/src/generated/`、`packages/api-client/generated/`、
   `packages/shared/src/permissions.gen.ts` 以及 `services/internal/` 下的
   `*_gen` 包都是生成物，不要手改。
4. `apps/*/dist`、`runtime-plugins/*/dist`、`services/internal/webui/dist`
   是构建产物，已被 gitignore。
5. 每个构建单元自带清单：`services/go.mod`、每个 `package.json`、
   `deploy/docker/Dockerfile`。根目录文件只做编排。

## 生成物

| 生成物 | 来源 | 落点 |
| --- | --- | --- |
| Go protobuf / gRPC / Kratos HTTP | `contracts/proto/**/v1` | `services/api/**/v1` |
| OpenAPI | `contracts/proto` | `packages/api-client/generated/openapi.yaml` |
| TypeScript schema | OpenAPI | `packages/api-client/src/generated/schema.d.ts` |
| 配置 schema | `services/internal/conf/conf.proto` | `services/internal/conf/conf.pb.go` |
| 权限常量 | `contracts/permissions.yaml` | `services/internal/pkg/rbac/permissions_gen.go`、`packages/shared/src/permissions.gen.ts` |
| Wire DI | `services/cmd/api/wire.go` | `services/cmd/api/wire_gen.go` |
| 扩展注册表 | `extensions/installed.lock.json` | 生成的 `*_gen` 包和 `features.gen.ts` |
| Web/admin 产物 | `apps/*/dist` | 打包时复制到 `services/internal/webui/dist/{web,admin}` |

## 扩展与运行时插件

- **Feature Extension（功能扩展）**：编译期由 `scripts/module.mjs` 复制进代码树的
  模块，可以添加 protobuf 契约、Go 服务和页面。
- **Example（示例）**：`kind: "example"` 的功能扩展。内置的 `example-notes`
  是一条完整的纵向切片。
- **Runtime Plugin（运行时插件）**：`runtime-plugins/` 下的目录，由服务端在启动时加载。
  它包含清单、浏览器 ESM，以及可选的 static/proxy 后端。内置的 `hello-panel` 演示了
  Vuetify 页面与静态 API。已标记为实验性，不属于 MVP 验收项。
  清单的 `entry`/`styles` 相对插件的 `dist/`；宿主通过
  `/proxy/plugins/<id>/frontend/` 提供。宿主应用全局安装 Vuetify，因此插件复用同一批
  组件实例，不会重复打包 Vuetify。

两类扩展的注册表都是生成的。手写代码只 import 生成的注册表包，因此安装或删除
扩展永远不会改动核心 server 或 router 文件。

## 单二进制，双 SPA

`make build` / `pnpm run build` 会内嵌两个构建：

- `/` → `apps/web/dist`
- `/admin/` → `apps/admin/dist`

`services/internal/webui` 按前缀选择挂载点，并对 history 模式路由返回
`index.html`。API 和代理前缀永远不会回退到 SPA。

## UI 体系

两个应用都使用 Vuetify 4，共享主题来自 `packages/ui`：

- `scaffoldLight` 与 `scaffoldDark` 主题，持久化在
  `localStorage["scaffold.theme"]`，可在顶栏切换。
- MDI 图标、胶囊按钮、圆角卡片，以及导航和顶栏的玻璃质感
  （`backdrop-filter`）。
- `packages/ui` 提供共享的 `PageHeader`、Snackbar 和确认弹窗，保证两个应用
  交互语言一致。
- 路由转场与卡片入场动画；登录页使用动态渐变背景。

## 配置

运行时配置在 `configs/{config.local,docker,k8s}.yaml`。环境变量统一使用
`SCAFFOLD_` 前缀。生产环境启动时拒绝默认的 JWT 与 OAuth state 密钥。
