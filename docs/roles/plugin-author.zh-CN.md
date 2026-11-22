# 插件与扩展开发指南

[English](plugin-author.md) | [简体中文](plugin-author.zh-CN.md)

> 适合谁：需要以编译期扩展或实验性运行时插件的方式为脚手架增加功能的开发者。

## 两种模型对比

| 维度 | 编译期扩展 | 运行时插件（实验性） |
| --- | --- | --- |
| 位置 | `extensions/templates/<id>/` | `runtime-plugins/<id>/` |
| 安装方式 | `pnpm run module:add -- <id>`（把文件拷入仓库） | 放入目录即可，服务器启动时加载 |
| 清单文件 | `module.json` | `plugin.json` |
| 能否修改核心/生成代码 | 能 —— `copy` 可指向仓库任意位置，并可触发代码生成 | 不能 —— 只发布浏览器 ESM + CSS + 可选的静态/代理后端 |
| 后端逻辑 | 编进二进制的真实 Go 代码 | `server/` 下的静态文件，或代理到上游 |
| 前端 | 源码并入应用，路由/菜单生成到 `features.gen.ts` | 从 `/proxy/plugins/<id>/frontend/` 动态加载 |
| 免重建安装/卸载 | 否 | 前端与资源可以；清单变更需要重启服务器 |
| 成熟度 | 有锁文件与回滚的正式流程 | 实验性：无沙箱、无签名校验 |

## 选型决策树

- 需要修改或新增**核心 Go 服务、proto 或生成代码**？→ 编译期扩展。
- 开发时想要**独立 HMR**、不重建宿主应用？→ 运行时插件（`SCAFFOLD_PLUGINS_DEV=true`）。
- 需要**不重建** Go 二进制和应用包就能装/卸功能？→ 运行时插件（重启只会重新加载清单）。
- 需要由 RBAC 目录和 `verify-authz` 兜底的**真实权限边界**？→ 编译期扩展；运行时插件的权限只在 `/proxy/plugins/<id>/api/` 边界生效。
- 只是加一个页面、一个侧边栏入口或简单的只读 API？→ 先用运行时插件，不够用再升级为扩展。

## 运行时插件

每个插件是 `runtime-plugins/` 下的一个目录，目录名必须等于清单里的 `id`。示例 `runtime-plugins/hello-panel/` 开箱可运行。

### plugin.json

关键字段（完整表格见[运行时插件 README](../../runtime-plugins/README.zh-CN.md)）：

| 字段 | 含义 |
| --- | --- |
| `id` | 小写 kebab-case；必须等于目录名 |
| `version` | 语义化版本 |
| `frontend.entry` | 相对 `dist/` 的路径，如 `index.js` |
| `frontend.devEntry` | `plugins.dev=true` 时使用的绝对 URL |
| `frontend.styles` | 相对 `dist/` 的路径 |
| `frontend.routes[]` | 路由记录；`component` 是模块 `components` 导出的键；可选 `nav` 增加侧边栏入口 |
| `backend.mode` | `none`、`static` 或 `proxy` |
| `backend.staticDir` | `static` 模式的目录；默认 `server` |
| `backend.upstream` | `proxy` 模式的上游 URL |
| `backend.permission` | 调用插件 API 所需的权限 |
| `permissions[]` | 写入数据库种子的权限码；控制目录可见性 |

### 资源与 API

- 前端资源：`GET /proxy/plugins/<id>/frontend/<file>` 由 `<plugin>/dist/` 提供。`entry: "index.js"` 对应 `/proxy/plugins/hello-panel/frontend/index.js`。
- 插件 API：`GET /proxy/plugins/<id>/api/<path>` 由 `server/`（static 模式）提供，或转发到 `backend.upstream`（proxy 模式）。
- 这些路由绕过 Kratos 中间件链、自行鉴权：必须持有有效访问令牌；仍须改密用户会收到 `MUST_CHANGE_PASSWORD`；`backend.permission` 不满足时返回 `FORBIDDEN`。
- 代理会附加 `X-Scaffold-Plugin-Id`、`X-Scaffold-User-Id`、`X-Scaffold-Username`；默认超时 30 秒。生产环境中上游主机必须列入 `plugins.allowed_upstreams`。

### 构建与开发

```bash
pnpm --filter @scaffold/plugin-hello build   # 单个插件
pnpm run build:plugins                       # 全部已安装插件

# HMR：两个终端
SCAFFOLD_PLUGINS_DEV=true pnpm run dev:server
pnpm --filter @scaffold/plugin-hello dev     # http://localhost:5174
```

`plugins.dev=true` 时目录返回 `devEntry` 而不是构建产物的 `entryUrl`，宿主直接从 Vite 导入。清单变更仍需重启后端，因为注册表在启动时加载一次。

### 宿主集成

- 宿主全局安装了 Vuetify，插件 SFC 可直接使用 `<VCard>`、`<VBtn>` 等组件，无需打包 Vuetify。
- 宿主通过 `window.__SCAFFOLD_PLUGIN_HOST__` 暴露全局对象（见 `packages/plugin-runtime/src/host.ts`），`ctx.host.vuetifyLib` 提供 `useTheme` 等组合式函数。
- 卸载即删除目录，没有锁文件，也没有代码生成步骤。

## 编译期扩展

扩展是 `extensions/templates/` 下的模板，由 `scripts/module.mjs` 拷入仓库（pnpm 别名：`module:add|module:remove|module:update|module:list|module:check-upstream`）。

### module.json

完整示例见 `extensions/templates/example-notes/module.json`。重要字段：

- `copy`：`{ from, to }` 对，分别相对于模板目录和仓库根。
- `registries`：拷入 `services/internal/extensions` 的 Go 文件由 `registry_gen` 空白导入；拷入 `services/internal/auth/oauth/providers` 的 OAuth provider 由 `oauthproviders_gen` 空白导入。
- `authz` / `authzImport` / `authzAlias`：写入 `services/internal/server/authz/extensions_gen.go`，使 `pnpm run verify-authz` 能覆盖到。
- `targets`：路由与菜单写入 `apps/{web,admin}/src/router/features.gen.ts`（切勿手改这些文件）。
- `codegen: true`：安装/移除后运行完整的 `pnpm run generate` 流水线。
- `generatedDirs`：卸载时删除。

### 安全与生命周期

- `extensions/installed.lock.json` 记录模板版本、安装时的 `baseHash`、逐文件 SHA-256、生成文件与 targets。
- `add` 校验依赖与冲突，先在 `.scaffold-tmp/` 暂存，失败时回滚已创建的文件。
- `remove` 拒绝删除你改过的文件（除非 `--force`），也拒绝删除被其他扩展依赖的扩展。
- `update` 只覆盖安装后未改动的文件；你改过的文件会以 `<file>.new` 写到目标旁边，并附 diff 摘要。
- 并发命令通过 `extensions/.lock` 互斥；所有命令支持 `--dry-run` 与 `--json`。
- `check-upstream` 对照 `extensions/upstream.json` 中固定的 go-admin 参考提交，并把（被 gitignore 的）报告写到 `extensions/upstream-report.md`；把它当作手工迁移清单即可。

## 权限

- 运行时插件的 `permissions[]` 会写入数据库种子，只控制目录可见性；`backend.permission` 才是真正的 API 门禁。插件前端的权限检查只是装饰——插件 JavaScript 以宿主权限运行。
- 如果功能需要**新增**权限码（扩展常见）：
  - [ ] 在 `contracts/permissions.yaml` 中添加。
  - [ ] 重新生成，使其进入 `services/internal/pkg/rbac/permissions_gen.go` 和 `packages/shared/src/permissions.gen.ts`。
  - [ ] 在 `services/internal/server/authz/mapping.go`（或扩展的 `authz` 块）中映射操作。
  - [ ] 运行 `pnpm run verify-authz`——未知操作默认拒绝（fail closed）。

## 测试与验证

- [ ] 构建：`pnpm run build:plugins`，确认 `dist/index.js` 与 `dist/style.css` 存在。
- [ ] 目录：`GET /api/v1/plugins` 返回你的清单，含 `entryUrl` 与路由。
- [ ] 资源：`GET /proxy/plugins/<id>/frontend/index.js` 返回 200。
- [ ] API：未认证 → 401；已认证但缺 `backend.permission` → 403 `FORBIDDEN`；已授权 → 200。
- [ ] 扩展安装：`pnpm run module:add -- example-notes`，然后在 `services/` 下 `go build ./...` 并运行 `pnpm run typecheck`。
- [ ] `pnpm run module:list` 能看到该扩展；`remove` 后工作树恢复干净。
- [ ] CI 执行相同检查：codegen + `verify-authz` + `git diff`、`go vet` + `go test -race`，以及扩展 add/remove 循环。

## 限制

- 运行时插件是实验性功能：无沙箱、无签名校验、无版本锁定——只安装你信任的插件。
- 宿主提供的 Vue、Vuetify、Pinia、vue-i18n **不得**打进插件包；它们通过 `packages/config-vite/plugin.ts` 外置。
- 清单变更需要重启后端；每个 id 只能有一个插件。
- 资源同源提供，没有 SSR。
- 多副本部署必须让每个副本带同样的插件（打进镜像或挂载共享卷）。插件注册表启动时加载一次，同样只适用于单副本假设。

## 相关

- [扩展](../extensions.zh-CN.md) —— module.mjs 命令、锁文件、内置模板
- [运行时插件 README](../../runtime-plugins/README.zh-CN.md) —— 清单参考与排障
- [认证、RBAC 与 SSO](../auth.zh-CN.md) —— `backend.permission` 背后的权限模型
- [架构](../architecture.zh-CN.md) —— 生成产物与边界规则
- [开发](../development.zh-CN.md) —— 代码生成流水线与日常工作流
