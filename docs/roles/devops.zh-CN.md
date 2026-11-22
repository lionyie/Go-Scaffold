# 运维与 SRE 指南

[English](devops.md) | [简体中文](devops.zh-CN.md)

> 适合谁：负责在容器、Compose 或 Kubernetes 中构建、部署和运行本脚手架的运维与 SRE 同学。

## 职责与交付模型

本仓库交付的是**一个 Go 二进制**（`bin/scaffold`），它内嵌了两个 Vue 单页应用并直接对外提供服务：

- `/` 提供来自 `apps/web/dist` 的公开站点。
- `/admin/` 提供来自 `apps/admin/dist` 的管理后台。

**没有 nginx，也没有独立的 Web 服务器**。`services/internal/webui/handler.go` 中的 `webui` 处理器保留了 `api/`、`proxy/`、`healthz`、`readyz` 这些前缀，其余路径全部回退到 SPA。静态资源缓存一年，`index.html` 不缓存，因此前端发版就是替换二进制本身。

## 构建产物与端口

| 端口 / 产物 | 用途 |
| --- | --- |
| 5173 | `apps/web` Vite 开发服务器（base `/`） |
| 5175 | `apps/admin` Vite 开发服务器（base `/admin/`） |
| 5174 | 运行时插件 HMR 开发服务器（`pnpm --filter @scaffold/plugin-hello dev`） |
| 8000 | 后端 HTTP（生产环境同时托管两个 SPA） |
| 9000 | 后端 gRPC |
| 80 | Kubernetes Service 的 HTTP 端口（映射到 8000） |
| 3306 | Compose 栈暴露的 MySQL |

产物：

- `apps/web/dist/` 与 `apps/admin/dist/`，由 `pnpm run build:web` / `pnpm run build:admin` 生成。
- Windows 下为 `bin/scaffold.exe`，其他平台为 `bin/scaffold`，由 `pnpm run build:server`（`scripts/build-server.mjs`）生成。
- `pnpm run build` 通过 `scripts/build.mjs` 执行完整流水线。

## 配置系统

`configs/` 下有三份配置文件：

- `configs/config.local.yaml` —— 本地开发；默认使用 SQLite，路径为 `file:../data/scaffold.db`，开启 WAL。
- `configs/docker.yaml` —— 容器内使用；默认 MySQL，JSON 格式日志。
- `configs/k8s.yaml` —— Kubernetes 清单使用；结构与 Docker 配置一致。

运维上需要记住的规则：

- 环境变量使用 **`SCAFFOLD_` 前缀**，覆盖 YAML 文件中的值（Kratos env source）。
- YAML 值支持 **`${NAME:default}` 占位符**，例如 `"${JWT_SECRET:dev-only-change-me}"`、`"${DB_SOURCE:...}"`。因此 `SCAFFOLD_DB_SOURCE` 覆盖 `data.database.source`，`SCAFFOLD_LOG_FORMAT` 覆盖 `log.format`，以此类推。
- 二进制通过 `-conf <路径>` 启动；容器入口默认 `/app/configs/docker.yaml`，k8s Deployment 覆盖为 `/app/configs/k8s.yaml`。

## 生产必备密钥与启动门禁

启用生产模式前必须设置：

- `SCAFFOLD_JWT_SECRET` —— JWT 签名密钥。
- `SCAFFOLD_OAUTH_STATE_SECRET` —— OAuth state 签名密钥。
- `SCAFFOLD_SEED_ADMIN_PASSWORD` —— 首次启动时种子 `admin` 账号的密码（默认 `admin123`）。
- `APP_ENV=production` —— 开启门禁。

当 `APP_ENV=production` 时，`services/cmd/api/main.go` 会在 JWT 密钥或 OAuth state 密钥为空或仍是开发默认值时**直接 panic 退出**；`services/internal/data/database/seed.go` 会拒绝带着未更换的种子管理员密码启动。如果 Pod 因这类 panic 反复重启，那是配置问题，不是应用缺陷。

## 本地全栈 Compose

`deploy/compose/docker-compose.yml` 会拉起 **MySQL 8.4 加服务端**：

```bash
docker compose -f deploy/compose/docker-compose.yml up --build
```

- MySQL 暴露在 `3306`，库名/用户名/密码均为 `scaffold`，数据持久化到 `mysql-data` 卷。
- 服务端通过 `depends_on` + `condition: service_healthy`（mysqladmin ping 健康检查）等待 MySQL 就绪。
- 服务端把 `SCAFFOLD_DB_SOURCE` 指向 MySQL，并使用**占位密钥**（`change-me-in-production`）——请勿把这套配置原样用于本地评估以外的场景。
- 停止用 `docker compose -f deploy/compose/docker-compose.yml down`（加 `-v` 可同时删除数据库卷）。

## 镜像构建

`deploy/docker/Dockerfile` 是三阶段构建：

1. **node:24-alpine** —— 安装 pnpm workspace，构建 `apps/web`、`apps/admin` 和运行时插件，然后把前端产物复制到 Go embed 路径。
2. **golang:1.27-alpine** —— 执行 `CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.Version=${VERSION}"`；用 `--build-arg VERSION=<tag>` 注入版本号。
3. **gcr.io/distroless/static-debian12:nonroot** —— 最终镜像只包含 `/app/scaffold`、`/app/configs` 和 `/app/runtime-plugins`；以 `nonroot` 用户运行，暴露 8000/9000。

本地构建：

```bash
docker build -f deploy/docker/Dockerfile -t scaffold:local .
```

## Kubernetes

清单位于 `deploy/k8s/`：

- `namespace.yaml` —— `scaffold-system`。
- `deployment.yaml` —— 单副本，镜像 `scaffold:latest`，`imagePullPolicy: IfNotPresent`，启动参数 `-conf /app/configs/k8s.yaml`。
- `service.yaml` —— 端口 80 -> 8000（http）、9000 -> 9000（grpc）。
- `secret.example.yaml` —— `scaffold-secrets`（`jwt-secret`、`oauth-state-secret`）的模板；apply 前请替换占位值。

探针：readiness 走 `/readyz`（初始延迟 5 秒，周期 10 秒），liveness 走 `/healthz`（初始延迟 10 秒，周期 20 秒）。

当前缺口，需要提前规划：

- [ ] 没有 MySQL 清单 —— `SCAFFOLD_DB_SOURCE` 指向的 `mysql:3306` 需要你自行提供。
- [ ] 没有 Ingress 和 TLS —— 请自行在 Service 前接 ingress 或负载均衡。
- [ ] 没有 HPA，且 `replicas: 1` 实际上是硬性要求（见下方扩缩容限制）。
- [ ] `scaffold:latest` + `IfNotPresent` 意味着节点缓存的镜像可能过期；正式部署请按发布打独立 tag。

## 数据库

- **本地：** SQLite（`file:../data/scaffold.db`，WAL 模式）。备份方式是在服务停止后复制该文件。
- **生产：** 通过 `SCAFFOLD_DB_SOURCE` 使用 MySQL（DSN 见 Compose 与 k8s 示例）。
- **表结构：** 当 `data.database.auto_migrate` 为 true 时（所有出厂配置的默认值），启动时执行 GORM `AutoMigrate`。`migrations/` 目录为空 —— 目前还没有手写迁移，因此有破坏性的表结构变更需要谨慎：升级前先备份（`mysqldump` 或卷快照），并优先采用增量式（只加不删）的模型变更。
- 种子数据（`data.database.seed`）会在空库上创建初始 admin 用户、角色和权限。

## 健康检查、日志与可观测性

- `GET /healthz` —— 存活探针；`GET /readyz` —— 就绪探针。两者都在 8000 端口，不受 SPA 回退影响。
- 日志输出到 stdout；`log.format` 本地为 `text`，`configs/docker.yaml` 和 `configs/k8s.yaml` 中为 `json`（可用 `SCAFFOLD_LOG_FORMAT` / `SCAFFOLD_LOG_LEVEL` 覆盖）。
- **目前没有接入指标和链路追踪** —— 没有 Prometheus 端点，也没有 OpenTelemetry exporter。在此之前，日志聚合是你唯一的观测手段。

## API 参考（Swagger UI）

非生产部署可以在 `/swagger/` 暴露生成的 OpenAPI 文档，由 `swagger.enabled`
（`SCAFFOLD_SWAGGER_ENABLED`）控制：

- `configs/config.local.yaml` 默认开启。
- `configs/docker.yaml` 与 `configs/k8s.yaml` 默认关闭。
- `APP_ENV=production` 时即使开关为 true 也会强制关闭。
- 关闭时 `/swagger/` 返回 404，不属于 SPA 回退。

staging 环境可在容器环境变量中加入 `SCAFFOLD_SWAGGER_ENABLED: "true"`，并且不要设置
`APP_ENV=production`。生产环境不要通过公网 Ingress 或负载均衡暴露 `/swagger/`。

## 扩缩容限制

目前水平扩容受到以下约束：

- OAuth state/ticket 存储是**内存态**，请求一旦在副本间漂移，OAuth 登录就会失败。在引入共享存储之前，保持 `replicas: 1` 或使用会话亲和（sticky session）。
- 刷新令牌存入数据库，但**登录限流器是内存态**（每个副本独立），插件资源从容器文件系统（`/app/runtime-plugins`）读取，因此**每个副本必须内置同一套插件** —— 插件无法在运行时只加到某一个副本上。
- 在提高副本数之前，先规划共享状态存储（Redis 或数据库实现）。

## CI 与发布

`.github/workflows/ci.yml` 包含五个 job：

| Job | 内容 |
| --- | --- |
| codegen | 执行 `pnpm run generate`、`pnpm run verify-authz`，然后 `git diff --exit-code` 检查代码生成漂移 |
| go | 在 `services/` 下执行 `go vet ./...` 和 `go test -race ./...` |
| web | `pnpm run lint:web`、`typecheck`、`build:web`、`build:admin` |
| extensions | 安装/移除 `oauth-github` 与 `example-notes` 扩展模块并验证工作区干净 |
| docker | 构建 `deploy/docker/Dockerfile`，tag 为 `scaffold:ci`，**push: false**，仅 main 分支 |

运维缺口：目前没有任何环节发布镜像。接入镜像仓库时，请在 docker job 中补充登录与推送步骤（或单独建 release workflow），并按 git SHA 或版本号打 tag，而不是用 `latest`。

## 故障排查表

| 症状 | 可能原因 | 排查方式 |
| --- | --- | --- |
| 容器启动即崩溃，panic 提到密钥 | `APP_ENV=production` 但密钥仍是默认值 | 查看 Pod 日志；检查 Secret 中的 `SCAFFOLD_JWT_SECRET`、`SCAFFOLD_OAUTH_STATE_SECRET` |
| 崩溃信息提到种子管理员密码 | 生产环境仍在使用 `admin123` | 首次启动前设置 `SCAFFOLD_SEED_ADMIN_PASSWORD` |
| 浏览器 502 / 连接被拒 | 后端未监听 8000 | `docker compose ps`、`kubectl get pods`、port-forward 8000 |
| readiness 探针失败 | 进程已起但 `/readyz` 不正常 | `kubectl describe pod`；在 8000 端口 curl `/readyz` |
| OAuth 登录在跳转后失败 | state 密钥未设置，或多副本内存态不共享 | 检查 `SCAFFOLD_OAUTH_STATE_SECRET`；核对副本数 / 会话亲和 |
| 后台页面能打开但 API 调用失败 | 请求没有到达后端 `/api` | 开发环境查 Vite 代理目标；k8s 查 Service 80 -> 8000 |
| 插件页面 404 或资源过旧 | 镜像中缺少插件或 dev 模式未开 | 检查镜像内 `/app/runtime-plugins/<id>`；核对 `SCAFFOLD_PLUGINS_DEV` |
| 开发服务器无法访问 | 预期端口被占用 | 5173（web）、5175（admin）、5174（插件）、8000/9000（后端） |
| gRPC 客户端连不上 | 只暴露了 80 端口 | Service 同时映射 9000 -> 9000；检查 LB/Ingress 是否放行 9000 |

## 相关

- [架构说明](../architecture.zh-CN.md)
- [认证与授权](../auth.zh-CN.md)
- [开发指南](../development.zh-CN.md)
