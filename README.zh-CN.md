# Scaffold

[English](README.md) | [简体中文](README.zh-CN.md)

一个面向 Go + TypeScript 项目的多端脚手架。默认内置管理端、RBAC、按需安装的第三方登录扩展，以及脚本化的项目改名流程，让你不用重写工具链就能把它变成真实项目。

## 内置内容

- **多端**：`apps/web`（公共站）与 `apps/admin`（管理端），Vue 3 + Vite + Vuetify 4，自带明暗主题、玻璃质感与页面转场。
- **UI 体系**：共享 Vuetify 主题、MDI 图标、Snackbar/确认弹窗，以及首次登录强制的改密弹窗。
- **后端**：`services/` 单一 Go module，Kratos HTTP/gRPC、protobuf 契约、GORM、JWT、RBAC、插件宿主、内嵌前端资源。
- **契约**：`contracts/proto` 下的 `.proto` 源文件，生成 Go 与 TypeScript 客户端。
- **扩展**：`extensions/` 下的编译期功能模块与可运行示例，由 `scripts/module.mjs` 管理。
- **运行时插件**：`runtime-plugins/` 下的实验性加载机制，不属于 MVP 验收项。

## 目录结构

```text
apps/
  web/                  公共站（端口 5173）
  admin/                管理端（端口 5175，部署在 /admin/）
services/
  cmd/api/              Go 入口 + wire
  api/                  生成的 Go protobuf 代码
  internal/
    biz/ data/ service/ 身份与权限的用例、仓储、传输适配
    server/             HTTP/gRPC、鉴权、OAuth 回调
    auth/               JWT、密码、OAuth provider 注册表
    extensions/         扩展示例的注册表
    plugin/             运行时插件加载器（实验性）
    webui/              内嵌 web + admin 构建产物
packages/               跨端共享 TypeScript 库
contracts/              proto 源文件 + 权限目录
extensions/             功能/示例模板与安装锁文件
runtime-plugins/        实验性运行时插件
deploy/                 Docker、Compose、Kubernetes
configs/                local/docker/k8s 运行时配置
scripts/                代码生成、扩展管理、改名、构建、清理
docs/                   架构、开发、认证、扩展、改名
```

## 快速开始

前置要求：Go 1.27+、Node 22.12+、pnpm 10+。`make` 可选。

```bash
pnpm run init          # 安装固定版本的代码生成工具与 workspace 依赖
pnpm run generate      # 生成权限常量、protobuf、OpenAPI/TS、wire

# 终端 1
pnpm run dev:server    # HTTP :8000，gRPC :9000

# 终端 2
pnpm run dev:web       # http://localhost:5173

# 终端 3
pnpm run dev:admin     # http://localhost:5175/admin/
```

使用 `admin / admin123` 登录。首次登录会强制修改密码；之后即可管理用户、角色与权限。

本地开发默认开启 Swagger UI：`http://localhost:8000/swagger/`；生产环境会强制关闭。

## 常用命令

`pnpm` 脚本是唯一真实入口，`make <target>` 只是别名。

| 目标               | 命令                                    |
| ------------------ | --------------------------------------- |
| 安装工具与依赖     | `pnpm run init`                         |
| 重新生成契约       | `pnpm run generate`                     |
| 校验生成物         | `pnpm run verify-generated`             |
| 校验鉴权映射       | `pnpm run verify-authz`                 |
| 启动后端           | `pnpm run dev:server`                   |
| 启动公共站         | `pnpm run dev:web`                      |
| 启动管理端         | `pnpm run dev:admin`                    |
| 构建全部           | `pnpm run build`                        |
| 类型检查           | `pnpm run typecheck`                    |
| 前端 lint          | `pnpm run lint:web` + `buf lint`        |
| 重置本地数据库     | `pnpm run db:reset -- --yes`            |
| 安装功能模块       | `pnpm run module:add -- oauth-github`   |
| 安装示例           | `pnpm run module:add -- example-notes`  |
| 删除扩展           | `pnpm run module:remove -- <id>`        |
| 检查 go-admin 上游 | `pnpm run module:check-upstream`        |
| 项目改名           | `pnpm run rename -- --name myapp --yes` |

## 第三方登录

第三方登录都是按需安装的扩展，只装需要的即可：

```bash
pnpm run module:add -- oauth-github
pnpm run module:add -- oauth-google
pnpm run module:add -- oauth-wechat
pnpm run module:add -- oauth-qq
pnpm run module:add -- oauth-oidc
```

每个 provider 从环境变量读取配置（例如 `SCAFFOLD_OAUTH_GITHUB_CLIENT_ID` 与 `_CLIENT_SECRET`），配置齐全后会出现在管理端登录页。详见[认证与 SSO](docs/auth.zh-CN.md)。

## Keycloak（OIDC，可选）

脚手架能对接任意 OIDC 提供方，但**不依赖**它。本地开发用 Keycloak realm 作为可选
compose override 提供：

```bash
docker compose -f deploy/compose/docker-compose.yml \
               -f deploy/compose/docker-compose.keycloak.yml up -d keycloak   # http://localhost:8080
```

带匹配的 OIDC 配置启动后端：

```bash
SCAFFOLD_OAUTH_STATE_SECRET=dev-only-oauth-state-secret \
SCAFFOLD_OAUTH_OIDC_ISSUER=http://localhost:8080/realms/scaffold \
SCAFFOLD_OAUTH_OIDC_CLIENT_ID=scaffold-admin \
SCAFFOLD_OAUTH_OIDC_CLIENT_SECRET=dev-scaffold-admin-secret \
SCAFFOLD_OAUTH_OIDC_DISPLAY_NAME=Keycloak \
SCAFFOLD_OAUTH_AUTO_REGISTER=true \
pnpm run dev:server
```

使用 `admin@example.com / admin123` 登录。已验证邮箱会把 Keycloak 身份关联到种子管理员；
Keycloak 的 realm role 就是权限码（`user:list`、`role:create` 等），名为 `admin` 的角色
等价于全部权限；`viewer@example.com / viewer123` 是只读账号。本地密码登录仍然保留。
详见[认证与 SSO](docs/auth.zh-CN.md)。

## 运行时插件（实验性）

仓库内置了一个可运行示例 `runtime-plugins/hello-panel/`：注册路由与侧边栏入口，使用
宿主提供的 Vuetify 组件（不重复打包），并通过受鉴权的插件代理调用静态 API。

插件从构建后的 `dist/` 目录提供。如果 Hello 插件加载时报 `index.js` 404，
先执行一次 `pnpm run build:plugins`。

管理端的**插件**页还支持从公开 Git 仓库或上传 ZIP 安装预构建插件；安装到
`data/plugins`，无需重启即可卸载。

```bash
pnpm run build:plugins                       # 构建所有已安装插件
pnpm run dev:server                          # :8000
pnpm run dev:admin                           # 侧边栏 → Hello plugin

# 插件源码热更新
SCAFFOLD_PLUGINS_DEV=true pnpm run dev:server
pnpm --filter @scaffold/plugin-hello dev      # :5174
```

删除插件目录即可卸载。清单字段、路径约定、后端模式、鉴权与排错详见
[runtime-plugins/README.zh-CN.md](runtime-plugins/README.zh-CN.md)。

## 部署

```bash
pnpm run build
node scripts/db-reset.mjs --yes
docker compose -f deploy/compose/docker-compose.yml up --build
kubectl apply -f deploy/k8s/
```

单二进制同时内嵌两个 SPA：`/` 是公共站，`/admin/` 是管理端。

## 文档

完整索引见[文档目录（中文）](docs/README.zh-CN.md)。

**刚加入项目？** 按角色选择阅读：[docs/roles/](docs/roles/README.zh-CN.md)：

- [新人全栈入门](docs/roles/onboarding.zh-CN.md)
- [后端开发](docs/roles/backend.zh-CN.md)
- [前端 / UI 开发](docs/roles/frontend-ui.zh-CN.md)
- [运维 / SRE](docs/roles/devops.zh-CN.md)
- [插件 / 扩展开发](docs/roles/plugin-author.zh-CN.md)
- [测试 / QA](docs/roles/testing-qa.zh-CN.md)
- [架构 / 技术负责人](docs/roles/architecture-lead.zh-CN.md)

主题参考：

- [架构](docs/architecture.zh-CN.md)
- [开发流程](docs/development.zh-CN.md)
- [API 参考](docs/development.zh-CN.md#api-参考与-swagger-ui)
- [认证与 SSO](docs/auth.zh-CN.md)
- [扩展机制](docs/extensions.zh-CN.md)
- [项目改名](docs/renaming.zh-CN.md)
- [优化执行日志](docs/scaffold-optimization-log.zh-CN.md)

英文版入口见 [docs/README.md](docs/README.md)。
