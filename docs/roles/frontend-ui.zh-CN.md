# 前端与 UI 开发指南

[English](frontend-ui.md) | [简体中文](frontend-ui.zh-CN.md)

> 适合谁：负责两个 Vue 3 单页应用、共享 Vuetify 设计系统和前端 API 层的开发者。

## 职责与两个 SPA

仓库包含两个相互独立的 Vue 3.5 + Vite 8 + Vuetify 4 单页应用，生产环境都由 Go 二进制统一托管：

| 应用 | 定位 | 开发端口 | Base | 入口 |
| --- | --- | --- | --- | --- |
| `apps/web` | 公开站点 | 5173 | `/` | `apps/web/src/main.ts` |
| `apps/admin` | 管理控制台 | 5175 | `/admin/` | `apps/admin/src/main.ts` |

做营销页、公开页面和匿名插件视图改 `apps/web`；做任何登录后的功能（用户、角色、权限、插件、个人资料）改 `apps/admin`。只有 `apps/admin` 带认证 store（`apps/admin/src/stores/auth.ts`）；`apps/web` 的 API 客户端用 `getToken: () => null` 配置，不携带凭证。

## 快速启动与代理

```bash
pnpm install
pnpm run dev:server   # Go 后端，http://127.0.0.1:8000
pnpm run dev:web      # apps/web，:5173
pnpm run dev:admin    # apps/admin，:5175
```

每个应用都有一个 `app.json`（如 `apps/admin/app.json`），声明 `name`、`base`、`devPort`、`embed`、`proxy`、`fallback`。Vite 开发服务器把 `/api`、`/proxy`、`/healthz`、`/readyz` 这几个前缀代理到 `http://127.0.0.1:8000`，因此前端始终同源调用后端，无需处理 CORS。

## 目录结构与共享包

每个应用在 `apps/*/src/` 下遵循同一套布局：

| 目录 | 内容 |
| --- | --- |
| `api/` | `client.ts`（openapi-fetch 接线）和 `errors.ts` |
| `components/` | 应用内组件（如 admin 的 `ChangePasswordDialog.vue`） |
| `layouts/` | `DefaultLayout.vue` 外壳 |
| `router/` | `index.ts`、`menu.ts`、生成的 `features.gen.ts` |
| `stores/` | Pinia setup store（`auth.ts` 仅 admin，`plugin.ts` 两个应用都有） |
| `styles/` | `index.css` 全局工具类 |
| `views/` | 路由级视图 |
| `i18n/` | `index.ts` 以及 `locales/en-US.ts`、`locales/zh-CN.ts` |

日常会引用的共享包：

| 包 | 提供 |
| --- | --- |
| `@scaffold/ui`（`packages/ui`） | `createScaffoldVuetify()`、`PageHeader`、`AppSnackbar`、`ConfirmDialog`、`notify`/`notifyApi`、`confirmDialog` |
| `@scaffold/shared`（`packages/shared`） | 生成的 `Permission` 常量、`APP_NAME`、`formatBytes` 等小工具 |
| `@scaffold/api-client`（`packages/api-client`） | 基于 openapi-fetch 的 `createApiClient()`、生成的 `schema.d.ts` 类型、`ApiError`/`toApiError` |
| `@scaffold/plugin-runtime`（`packages/plugin-runtime`） | `installPluginHost()`、`loadPlugins()`、插件清单类型 |

## Vuetify 设计系统

不要在应用里直接调用 `createVuetify()`。两个 `main.ts` 都调用 `@scaffold/ui` 的 `createScaffoldVuetify()`（`packages/ui/src/theme.ts`），它会：

- 注册全部 Vuetify 组件与指令，并接入 MDI 图标集（`@mdi/font`）；
- 定义 `scaffoldLight` 与 `scaffoldDark` 两套主题及共享色板；
- 设置组件默认值（胶囊按钮、`xl` 圆角卡片、outlined 且 comfortable 的输入框等）。

当前主题持久化在 `localStorage` 的 `scaffold.theme` 键；`main.ts` 启动时读取并把 `scaffoldDark` 或 `scaffoldLight` 传给 `createScaffoldVuetify()`。在 CSS 中通过 Vuetify 变量引用颜色，例如 `rgb(var(--v-theme-primary))`。

全局工具类在 `apps/*/src/styles/index.css`，对两套主题都生效：`.glass`、`.gradient-text`、`.hover-lift`、`.metric-card`、`.page-card`、`.blob`。优先复用这些类而不是写一次性样式，保证明暗主题都正常。

## 共享组件

`packages/ui` 导出三个组件和两个命令式助手：

- `PageHeader`：页面标题、可选描述和 `actions` 插槽。**未全局注册**，每个视图按需导入：`import { PageHeader } from '@scaffold/ui'`。
- `AppSnackbar` 和 `ConfirmDialog`：在两个 `main.ts` 里通过 `app.component(...)` 全局注册，模板中可直接使用（由 `App.vue` 挂载一次）。
- `notify(text, color)` / `notifyApi.success|error|warning|info(text)`：在任意脚本中触发全局 snackbar。
- `confirmDialog(title, message, options?)`：返回 `Promise<boolean>`，删除等破坏性操作用它代替 `window.confirm`。

## 路由与菜单

路由在 `apps/admin/src/router/index.ts`（web 端的 `apps/web/src/router/index.ts` 更精简）。管理端菜单位于 `apps/admin/src/router/menu.ts`：`coreMenu` 加上生成的 `featureMenu`。

`RouteMeta` 支持以下字段：

| 字段 | 作用 |
| --- | --- |
| `public` | 完全跳过认证（登录页、OAuth 回调） |
| `permission` | 所需权限码；不满足则跳转 403 视图 |
| `titleKey` | 用于文档标题的 i18n 键 |
| `pluginId` | 标记由运行时插件贡献的路由 |

管理端路由守卫按此顺序执行：`public` 路由直接放行 → `auth.restore()` → 未登录重定向到登录页 → `mustChangePassword` 用户强制跳到个人资料页 → 首次加载运行时插件 → 检查 `meta.permission`，失败进 forbidden 视图。

两个应用的 `features.gen.ts` 都由 `pnpm run module:add/remove` 生成，**禁止手改**。自己的路由加到 `index.ts`，菜单项加到 `menu.ts`。

## 权限控制

始终使用 `@scaffold/shared` 中生成的常量：

```ts
import { Permission } from '@scaffold/shared'

meta: { titleKey: 'nav.users', permission: Permission.UserList }
```

常量来自 `packages/shared/src/permissions.gen.ts`（由 `contracts/permissions.yaml` 生成，勿手改）。ESLint 在 `apps/admin/src/**/*.ts` 中禁止 `'user:list'` 这类权限字符串字面量（`no-restricted-syntax`），写了会在 `pnpm run lint:web` 报错。组件内用 `useAuthStore().hasPermission(code)` 判断；`is_admin` 用户绕过所有检查。注意 `apps/admin/src/directives/permission.ts` 里定义了 `vPermission` 指令，但目前**未在 `main.ts` 注册**，修复前不要依赖它。

## 状态管理

两个应用都使用 Pinia setup store：

- `apps/admin/src/stores/auth.ts`：会话状态。访问令牌**只存内存**；刷新令牌持久化在 `localStorage` 的 `scaffold.refreshToken` 键。`refresh()` 是单飞（single-flight）实现：并发调用共享同一个进行中的请求。`restore()` 由路由守卫调用，页面刷新后重建会话。
- `stores/plugin.ts`（两个应用）：拉取插件清单，通过 `@scaffold/plugin-runtime` 加载并记录失败。`loaded` 标志保证路由守卫只加载一次。

能用组合式函数或视图内 `ref` 解决的服务端状态，就不要塞进 store。

## API 调用

`packages/api-client/src/client.ts` 基于生成的 `schema.d.ts` 构建类型化的 openapi-fetch 客户端。应用以**空 base URL** 实例化，因此所有请求都是同源的（开发环境走 Vite 代理）。

- 认证调用用 `api()`：自动注入 `Authorization: Bearer <token>`，遇到 401 时执行一次刷新并重试（admin 端用 `x-scaffold-retried` 头防重入）。
- `rawClient` 只用于登录、刷新、登出——它没有任何拦截器。

错误处理：用 `toApiError(error)` 转换后，把 `error.reason` 映射到 i18n 键 `errors.<REASON>`；`apps/*/src/api/errors.ts` 的 `errorMessage()` 已封装此逻辑，翻译缺失时回退到原始 message。注意线上报文是 protojson 格式：字段为 camelCase，64 位整数以字符串传输。

任何 proto 变更后运行 `pnpm run generate` 并提交重新生成的 `packages/api-client` 产物——`packages/api-client/generated/` 与 `packages/api-client/src/generated/` 下的文件一律不得手改。

## 国际化

vue-i18n 11 组合式模式，配置在 `apps/*/src/i18n/index.ts`：

- 语言：`en-US` 与 `zh-CN`；默认 `zh-CN`，回退 `en-US`。
- 持久化在 `localStorage` 的 `vela.locale` 键；用导出的 `setLocale()` 切换。
- 文案按命名空间组织：`app`、`nav`、`common`、`login`、`home`、`user`、`role`、`permission`、`plugin`、`profile`、`oauth`、`errors`。

新增页面时，在 `locales/en-US.ts` 和 `locales/zh-CN.ts` 的同一命名空间下各加一份键。

## 新增页面 5 步清单

- [ ] 创建视图，如 `apps/admin/src/views/FeatureView.vue`，使用 `PageHeader` 和共享 CSS 类。
- [ ] 在 `apps/admin/src/router/index.ts` 注册路由，配上 `titleKey`，需要时加 `permission: Permission.Something`。
- [ ] 在 `apps/admin/src/router/menu.ts` 的 `coreMenu` 加菜单项，`titleKey` 与权限与路由保持一致。
- [ ] 在两个语言文件的对应命名空间下补翻译。
- [ ] 通过 `api()` 发起请求，用 `errorMessage()` 配合 `notifyApi` 展示错误。

完成后用 `pnpm run lint:web` 和 `pnpm run typecheck` 验证。

## 已知不一致

完整清单见[架构负责人指南 · 已知不一致](architecture-lead.zh-CN.md#已知不一致)。与前端直接相关的几条：

- `packages/shared/src/index.ts` 导出 `APP_NAME = "Vela"`，两个应用的语言设置持久化在 `vela.locale`，而主题和刷新令牌的键用的是 `scaffold.*` 前缀。
- `apps/web/src/api/client.ts` 的重试头是 `x-vela-retried`，`apps/admin` 用的是 `x-scaffold-retried`。
- `vPermission` 指令已定义但从未在 `apps/admin/src/main.ts` 注册。
- `PageHeader` 按视图导入，而 `AppSnackbar`、`ConfirmDialog` 全局注册——这是有意的拆分，但容易踩坑。
- 前端没有任何自动化测试；各 web 包的 `test` 脚本都是 `echo no-tests`。

## 相关

- [架构](../architecture.zh-CN.md)——分层、边界、生成产物
- [开发](../development.zh-CN.md)——首次运行与日常开发循环
- [插件与扩展作者指南](plugin-author.zh-CN.md)——运行时插件与编译期模块
- [已知不一致](architecture-lead.zh-CN.md#已知不一致)——影响前端的技术债
