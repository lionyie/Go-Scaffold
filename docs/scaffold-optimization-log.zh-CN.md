# Scaffold 优化执行日志

[English](scaffold-optimization-log.md) | [简体中文](scaffold-optimization-log.zh-CN.md)

格式：`任务ID | 变更文件 | 验收命令 | 结果 | 偏差说明`

| 任务 | 变更 | 验收命令 | 结果 | 偏差 |
| --- | --- | --- | --- | --- |
| Phase 0 | `apps/{web,admin}`、`services/`、`contracts/`、`extensions/`、`runtime-plugins/`、`deploy/`、`configs/`、`scripts/` | `pnpm run generate`；`go -C services build ./...`；`pnpm run typecheck` | 通过 | 仓库尚无 commit，见文末说明 |
| T1 | `extensions/`（原 modules）、`runtime-plugins/`（原 plugins）、注册表、文档、CI、ignore 规则 | `grep -rn "modules/templates\|internal/modules\|installed\.json\|plugins/example" ...` | 无匹配（exit 1） | `.turbo` 缓存仍引用旧插件路径，已删除 |
| T3 | `package.json`、`Makefile`、`scripts/{init,tool-versions,generate,verify-*}.mjs` | 无 make 环境下 `pnpm run init && pnpm run generate && pnpm run build` | 通过 | 新增 `scripts/build.mjs` 编排 web + admin + plugins + server |
| T2 | `scripts/module.mjs`、`extensions/installed.lock.json`、模板 | `add`；重复 `add`；`remove`；`update`；`remove --force` | 通过：重复安装被拒绝、update 写出 `.new`、force 删除清理干净 | 进程锁用简单 pid 锁，未用 `proper-lockfile` |
| T4 | `scripts/generated-dirs.mjs`、`scripts/generate.mjs`、`scripts/verify-generated.mjs` | 对生成产物取哈希 → `generate` → 再取哈希 | 完全一致（幂等） | `services/api` 清理时保留 `services/api/identity/v1/redact.go`；`services/internal/conf` 只清理 `conf.pb.go` |
| T5 | `cmd/genpermissions`、`rbac` 常量、`scripts/verify-authz.mjs`、管理端 `Permission` 用法、ESLint | `pnpm run generate`；`pnpm run verify-authz`；临时的缺失映射测试 | 通过：`authz ok: 19 operations, 19 requirements, 11 permissions` | 缺失映射测试在开发期完成，最终代码树通过校验 |
| T6 | `internal/auth/oauth`、`OAuthStore`、OAuth handler、provider 模板、测试 | `go -C services test ./internal/auth/oauth` | 通过（5 个安全用例） | 仅内存实现；Redis/DB 方案已写入文档 |
| T7 | `seed.go`、`main.go`、登录限流、改密流程 | 冒烟：登录 → 403 → 改密 → 重登 → 200；5 次错误登录后 429；`APP_ENV=production` 启动 | 通过：`401×5,429`；默认 JWT 与默认管理员均 fail fast | 本机无 C 编译器，未跑 `go test -race` |
| T8 | `scripts/db-reset.mjs`、`services/internal/data/migrations/`、`data.go`、`data_test.go` | `pnpm run db:reset -- --yes`；`go -C services test ./internal/data/database` | 通过 | 空目录 + AutoMigrate；golang-migrate 作为文档化的升级路径 |
| T9 | `scripts/rename.mjs` | 临时副本执行 `rename --name demo --yes`；`pnpm typecheck`；`go -C services build ./...` | 通过（typecheck 10/10） | `scaffold.json` 文件名不变；lock 文件重建而非文本替换 |
| T10 | `extensions/upstream.json`、`check-upstream` | `check-upstream --json`；`--strict` | 通过：离线状态、strict 退出码 1；固定 commit 来自更早的一次在线检查 | 本环境访问 GitHub 不稳定 |
| T11 | `module.json#targets`、`features.gen.ts` | 安装 `example-notes`；grep 两个生成文件 | 通过：web 与 admin 均有 `/notes` 路由与菜单 | 检查后示例已删除 |
| T12 | `apps/*/app.json`、Vite 配置、`copy-web-dist`、`webui` | `pnpm run build`；冒烟 `/`、`/admin/`、`/admin/users` | 通过：200 + SPA HTML；产出 `bin/scaffold.exe` | `build-server` 无需读 manifest；内嵌路径是静态的 |
| T13 | `services/internal/webui/fallback_test.go` | `go -C services test ./internal/webui` | 通过（fallback、静态资源、保留 API、重定向） | 无 |
| T14 | `runtime-plugins/README.md`、架构文档 | 人工 review | 通过 | 插件运行时保持实验性 |
| T15 | OAuth 测试、RBAC 矩阵、权限一致性、DB 迁移测试 | `go -C services test ./...`；`pnpm run verify-authz` | 通过 | 本地未带 `-race`；CI 在 Linux 上带 `-race` 运行 |

## 最终回归

```text
pnpm run generate              通过
pnpm run verify-authz          通过（19 operations / 19 requirements / 11 permissions）
go -C services vet ./...       通过
go -C services test ./...      通过
pnpm -r --if-present typecheck 通过
pnpm run lint:web              通过
buf lint                       通过
pnpm run build                 通过（web + admin + server）
API/SPA 冒烟                   通过
```

## 后续：强制改密与 Vuetify UI

| 项目 | 变更 | 验收 | 结果 |
| --- | --- | --- | --- |
| 强制改密 Bug | `/auth/me` 被改密守卫拦截，SPA 无法加载用户、无法渲染弹窗。authz 现在放行仅登录态的接口（`/auth/me`、登出），拦截需要权限的接口。管理端展示持久的改密弹窗，改密成功后轮换 access token。 | 无头 Chrome：`admin/admin123` 登录 → `/admin/profile` 出现弹窗 → 提交新密码 → 进入 `/admin/dashboard`，浏览器控制台 0 错误。 | 通过 |
| Element Plus → Vuetify 4 | 两个应用与共享 UI 迁移到 Vuetify 4 + MDI；共享明暗主题、玻璃质感、路由/卡片转场、Snackbar/确认弹窗、favicon、hero 文案本地化。所有依赖与模板中的 Element Plus 已移除。 | `pnpm -r typecheck`、`pnpm run lint:web`、`pnpm run build:web`、`pnpm run build:admin`，登录/首页/仪表盘宽屏截图。 | 通过 |

## 说明

- 仓库目前仍没有 commit。因此 `verify-generated` 会提示"no commits yet;
  skipping diff check"。完成首次提交后，CI 的生成物防漂移校验才会生效。
- `go test -race` 在 Windows 上需要 C 编译器，本机没有。同样的测试会在 CI 的
  Linux 上带 `-race` 运行。
- go-admin 仅做跟踪，不合并。go-admin 基于 Gin，本脚手架基于 Kratos。
