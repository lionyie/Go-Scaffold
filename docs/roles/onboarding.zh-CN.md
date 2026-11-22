# 新人全栈入门

[English](onboarding.md) | [简体中文](onboarding.zh-CN.md)

> 适合谁：刚加入的开发者，需要在第一天克隆本 monorepo、把 Go 后端和两个 Vue 前端跑起来，并完成一次安全的首次改动。

## 第一天路线图

- [ ] 装好前置工具（Go、Node、pnpm）并克隆仓库。
- [ ] 执行 `pnpm run init`，安装工作区依赖和锁定版本的代码生成工具。
- [ ] 执行 `pnpm run generate`，重新生成契约代码、权限常量和 API 客户端。
- [ ] 开三个终端分别运行：`pnpm run dev:server`、`pnpm run dev:web`、`pnpm run dev:admin`。
- [ ] 用 `admin / admin123` 登录后台，并完成强制改密。
- [ ] 浏览 `http://localhost:5173`（公开站点）和 `http://localhost:5175/admin/`（管理后台）。
- [ ] 完成一次小的前端改动和一次小的后端改动（见下文）。
- [ ] 提交首个 PR 前，在本地跑一遍质量门禁。

## 前置工具与版本

| 工具 | 版本要求 | 说明 |
| --- | --- | --- |
| Go | 1.27+ | `services/go.mod` 声明 `go 1.27.0`。 |
| Node | 22.12+ | 见 `package.json` 的 `engines`；仓库 `.nvmrc` 固定为 24。 |
| pnpm | 10 | 由 `packageManager: pnpm@10.33.2` 锁定，可用 `corepack enable` 启用。 |
| make | 可选 | 非必需；所有流程都通过 pnpm 脚本和 `scripts/*.mjs` 完成。 |

## 5 分钟跑起来

```bash
pnpm run init        # 安装依赖 + 锁定版本的代码生成工具
pnpm run generate    # proto -> Go/TS 代码、权限常量、OpenAPI schema
```

然后打开三个终端：

```bash
pnpm run dev:server   # Go 后端：HTTP :8000，gRPC :9000
pnpm run dev:web      # 公开站点：http://localhost:5173
pnpm run dev:admin    # 管理后台：http://localhost:5175/admin/
```

两个 Vite 开发服务器都会把 `/api`、`/proxy`、`/healthz`、`/readyz` 代理到 `http://127.0.0.1:8000`，前后端同源，无需配置 CORS。后端默认配置来自 `configs/config.local.yaml`：SQLite 落在 `data/scaffold.db`（自动创建），HTTP 监听 `0.0.0.0:8000`，gRPC 监听 `:9000`。

## 登录与强制改密

种子数据会创建管理员 `admin / admin123`（见 `services/internal/data/database/seed.go`）。首次登录流程：

1. 后台路由守卫检测到账号处于待轮换状态，弹出不可关闭的改密对话框。
2. 完成改密前，只有认证自服务接口（`/auth/me`、登出、改密）可用；带权限的接口一律返回错误原因 `MUST_CHANGE_PASSWORD`。
3. 改密完成后控制台全部解锁。生产环境在种子密码未轮换时拒绝启动；首次启动前请设置 `SCAFFOLD_SEED_ADMIN_PASSWORD`。

登录限流为同一 IP+用户名每分钟最多失败 5 次（返回 429），不要对开发服务器写脚本暴力重试。

## 仓库地图

| 路径 | 内容 |
| --- | --- |
| `apps/` | 两个 Vue 3.5 + Vite 8 + Vuetify 4 单页应用：`apps/web`（base `/`）和 `apps/admin`（base `/admin/`）。 |
| `services/` | Go（Kratos v3）后端；入口 `services/cmd/api/main.go`；`services/internal/` 下分层为 `server -> service -> biz -> data`。 |
| `packages/` | 共享 TS 包：`api-client`、`ui`、`shared`、`config-vite`、`plugin-runtime`。 |
| `contracts/` | 事实来源：`proto/**/v1/*.proto` 与 `permissions.yaml`。 |
| `extensions/` | 编译期扩展模板与 `installed.lock.json`。 |
| `runtime-plugins/` | 运行时插件（如 `hello-panel`），通过 `/proxy/plugins/` 提供服务。 |
| `deploy/` | `docker/Dockerfile`、`compose/docker-compose.yml`、`k8s/*.yaml`。 |
| `configs/` | 后端配置文件：`config.local.yaml`、`docker.yaml`、`k8s.yaml`。 |
| `scripts/` | 仓库自动化：`init.mjs`、`generate.mjs`、`module.mjs`、`dev-server.mjs` 等。 |
| `docs/` | 文档，包括 `docs/roles/` 下的角色系列。 |

## 第一个改动建议

**最小安全前端改动：** 修改 `apps/web/src/i18n/locales/zh-CN.ts` 中的一条文案（同时改 `en-US.ts` 中对应的 key），刷新 `http://localhost:5173`。Vite HMR 即时生效，且不涉及任何生成文件。

**最小安全后端改动：** 在既有用例（如 `services/internal/biz/user.go`）里加一行日志或调整一条错误信息，然后重启 `pnpm run dev:server`。因为没有改动 proto 契约，无需重新生成代码。

以下情况**必须**运行 `pnpm run generate`：修改 `contracts/proto/**/v1/*.proto` 或 `contracts/permissions.yaml`。切勿手工编辑生成产物（`services/api/**`、`services/internal/conf/conf.pb.go`、`permissions_gen.go`、`packages/api-client/src/generated/**`、`apps/*/src/router/features.gen.ts`）。

## 质量门禁

推送前在本地跑一遍，CI 会全部强制执行：

```bash
pnpm run typecheck      # turbo 跨 apps/packages 做类型检查
pnpm run lint:web       # eslint 检查 apps、packages、extensions、scripts
pnpm run verify-authz   # 校验 authz 映射与权限目录一致性
go -C services test ./...   # Go 单元测试（有 cgo 时加 -race）
```

`pnpm run generate` 还必须保持幂等；CI 会重新生成并对任何 `git diff` 报错。

## 常见坑

- **端口冲突：** 8000（后端 HTTP）、9000（gRPC）、5173（web）、5175（admin）；插件 HMR 模式还会占用 5174。重启前先清掉残留进程。
- **改了生成文件：** 上述生成路径下的手工改动会被 `pnpm run generate` 覆盖。请改契约，再重新生成。
- **`go test -race`：** 依赖 cgo。没有 C 编译器的机器请直接跑 `go test ./...`；race 构建由 Linux CI 执行。
- **首次登录“卡住”：** 强制改密弹窗按设计会挡住带权限的页面，完成改密即可，不要试图绕过。
- **JSON 中的 64 位整数：** 响应使用 protojson，int64 字段以字符串返回、枚举以名称返回，前端不要盲目 `parseInt`。

## 术语表

| 术语 | 本仓库中的含义 |
| --- | --- |
| SPA | 单页应用；本仓库有两个（`apps/web`、`apps/admin`）。 |
| proto 契约 | 定义 RPC 与 HTTP 绑定的 `contracts/proto/**/v1/*.proto` 文件。 |
| 生成文件 | 由 `pnpm run generate` 产出的任何文件，一律只读。 |
| 扩展（extension） | `extensions/templates/` 下的编译期模块，用 `pnpm run module:add` 安装。 |
| 运行时插件 | `runtime-plugins/` 下的即插即用插件，运行时经 `/proxy/plugins/` 加载。 |
| RBAC | 基于角色的访问控制；权限目录在 `contracts/permissions.yaml`。 |
| authz 映射 | `services/internal/server/authz/mapping.go` 中的操作到权限映射表。 |
| 单二进制 | 构建出的 Go 服务内嵌两个 SPA：`/` 提供 web，`/admin/` 提供 admin。 |

## 相关

- [backend.zh-CN.md](backend.zh-CN.md)
- [frontend-ui.zh-CN.md](frontend-ui.zh-CN.md)
- [devops.zh-CN.md](devops.zh-CN.md)
- [开发指南](../development.zh-CN.md)
