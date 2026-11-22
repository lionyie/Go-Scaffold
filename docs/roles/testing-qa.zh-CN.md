# 测试与质量保障指南

[English](testing-qa.md) | [简体中文](testing-qa.zh-CN.md)

> 适合谁：负责运行测试套件、盯 CI 门禁、提交缺陷报告的工程与 QA 人员。

## 职责与现状

本仓库的自动化测试**只存在于 Go 侧**:`services/` 下正好有六个测试文件；前端测试**为零**——`apps/*` 和 `packages/*` 里每个 `test` 脚本都是 `"echo no-tests"`,`pnpm run test:web` 只是通过 Turbo 跑这些空脚本。CI 全绿不代表"产品可用",只代表"Go 测试 + lint + 类型检查通过"。汇报质量结论时请按这个口径说清楚。

## 测试地图

| 文件 | 覆盖内容 |
| --- | --- |
| `services/internal/auth/oauth/oauth_test.go` | OAuth state 过期、一次性 ticket 防重放、state 签名、PKCE、按邮箱关联账号 |
| `services/internal/pkg/rbac/permissions_consistency_test.go` | 生成的 `permissions_gen.go` 目录与 `contracts/permissions.yaml` 保持一致 |
| `services/internal/webui/fallback_test.go` | `/` 与 `/admin/` 的 SPA fallback、保留的 API 前缀、`/admin` 重定向 |
| `services/internal/data/database/data_test.go` | SQLite 迁移能创建 `users` / `oauth_accounts` 表；迁移器接受空目录 |
| `services/internal/server/middleware/authz_test.go` | Authz 中间件：公共路由、未知 operation 默认拒绝、缺少 principal、允许/拒绝、管理员绕过、强制改密 |
| `services/internal/plugin/registry_test.go` | 插件注册表：`/proxy/plugins/<id>/frontend/` 下的资源 URL、开发模式入口 |

## 运行方式

```bash
# 运行全部 Go 测试（在仓库根目录执行）
go -C services test ./...

# 竞态检测（需要 cgo;Windows 上需安装 gcc，如 MSYS2/MinGW）
go -C services test -race ./...

# 校验权限目录、authz 映射与前端权限码三者一致
pnpm run verify-authz

# 校验生成代码与源契约(proto、权限、模块)一致
pnpm run verify-generated
```

CI 在 Linux 上执行 `go test -race ./...`，本地不带 `-race` 的运行只能算冒烟检查。

## CI 门禁

定义在 `.github/workflows/ci.yml`，每次推送 `main` 和每个 PR 都会触发：

| 任务 | 执行内容 | 失败意味着 |
| --- | --- | --- |
| `codegen` | `pnpm run generate`、`pnpm run verify-authz`，然后 `git diff --exit-code` | 手工改了生成文件，或改了 proto/权限/模块后忘记重新生成 |
| `go` | 在 `services/` 下 `go vet ./...` 与 `go test -race ./...` | 编译错误、vet 违规、测试失败或数据竞争 |
| `web` | `lint:web`、`typecheck`、`build:web`、`build:admin` | ESLint 报错、TS 类型错误或 Vite 构建失败 |
| `extensions` | 对 `oauth-github` 和 `example-notes` 执行 `module.mjs add/remove`，外加 Go 构建 | 模块生命周期坏了，或留下了未提交的漂移 |
| `docker` | 构建 `deploy/docker/Dockerfile`（不推送，仅 main 分支） | 生产镜像无法构建 |

关键在于 `git diff --exit-code`：代码生成之后工作区必须干净，生成文件过期会直接卡 CI。

## 前端测试缺口

前端目前没有任何测试运行器：没有 Vitest、Playwright 或 Cypress。空脚本（`"test": "echo no-tests"`）只是为了让 `turbo run test` 能退出码为 0。要补上真实测试，必须先选型并接线：

- 单元/组件测试：在 workspace 中引入 Vitest，给每个包加 `test` 脚本，并配置 Turbo 任务，让 `pnpm run test:web` 真正执行。
- 端到端测试：引入 Playwright，目标指向 dev 服务器（5173/5175）或 Go 二进制（8000)，并新增一个 CI 任务。

在补齐之前，不要宣称有 UI 覆盖率——请使用下面的手工验收清单。

## 手工验收清单

在干净的本地实例上执行（`pnpm run dev:server`,web 5173、admin 5175，或构建后的二进制 8000):

- [ ] 用种子管理员（`admin` / `admin123`）登录——在任何其他页面加载前会被强制改密
- [ ] 改密后 admin 各页面正常加载，用户列表正常渲染
- [ ] 创建一个无角色的普通用户——访问受权限保护的 admin 路由会被重定向/拒绝，直接调 API 返回 403
- [ ] 授予只含一个权限的角色——恰好只有对应页面解锁，其余不变
- [ ] OAuth：发起 GitHub/微信流程、完成回调，验证 ticket 只能兑换一次，第二次兑换同一 ticket 失败
- [ ] 安装运行时插件——菜单出现对应入口，面板从 `/proxy/plugins/<id>/frontend/` 正常加载
- [ ] 在 `/admin/users` 之类的深层路由上刷新——SPA fallback 返回 `index.html` 而非 404
- [ ] `curl http://127.0.0.1:8000/healthz` 与 `/readyz` 均返回 200
- [ ] 同一账号连续输错密码 5 次——第 6 次返回 429（登录限流）

## 新增 Go 测试指南

沿用仓库现有模式：

- 优先表驱动测试；参考形态见 `webui/fallback_test.go`（用例结构体 + 循环 + `httptest`)。
- 测中间件时在接缝处造假而不是造假网络：`authz_test.go` 用 `fakeChecker` 替代 RBAC、用 `fakeTransport` 的 `Operation()` 返回路由，再通过 `kerrors.FromError` 断言 Kratos 错误码与 reason。
- 数据层测试用 SQLite 内存库：`gormlite.Open(":memory:")` 加 `db.AutoMigrate(...)`，参考 `data_test.go`。绝不要碰默认的 `file:../data/scaffold.db`。
- 保持测试确定性：不用真实时钟 sleep、不用随机端口、不访问网络。像 OAuth 测试那样把时间/TTL 作为参数注入（例如传负时长）。
- 需要测未导出行为时把测试放在同包（`package foo` 而非 `foo_test`)，现有六个文件都这么做。

## 缺陷报告建议字段

每份报告都包含以下字段（仓库里没有模板文件，直接复制下表）:

| 字段 | 示例 |
| --- | --- |
| 环境 | 本地 dev / Docker / k8s；操作系统；Go 与 Node 版本 |
| 提交 | `git rev-parse HEAD` 的输出 |
| 复现步骤 | 编号、最小化、能从干净状态复现 |
| 期望与实际 | 各一行 |
| 日志 | 后端 stderr、浏览器控制台、相关的 `curl -v` 输出 |
| 截图 | 仅当缺陷是视觉问题时 |

务必写明是在 dev 服务器还是构建后的二进制上复现的，以及 `pnpm run verify-generated` 是否干净——生成物漂移会产生"假 bug"。

## 相关

- [开发指南](../development.zh-CN.md)——日常开发循环、测试与 lint 一节
- [DevOps 与 SRE](devops.zh-CN.md)——CI、Docker 与部署门禁
- [后端开发](backend.zh-CN.md)——测试应当针对的 Go 分层
- [架构负责人 · 已知不一致](architecture-lead.zh-CN.md#已知不一致)——那些容易被误报为 bug 的现存怪癖
