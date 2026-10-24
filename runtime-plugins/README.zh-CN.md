# 运行时插件

[English](README.md) | [简体中文](README.zh-CN.md)

> EXPERIMENTAL —— 不属于 MVP 验收项。

运行时插件是 `runtime-plugins/` 下的一个目录，由服务端在启动时加载。它包含浏览器 ESM、
可选 CSS、可选的 static/proxy 后端，以及一份 `plugin.json` 清单。与 `extensions/` 中的
功能扩展不同，运行时插件不会被编译进 Go 后端或应用 bundle。

仓库内置了一个最小可运行示例 [`hello-panel/`](hello-panel/)：注册路由与侧边栏入口，
使用宿主提供的 Vuetify 组件，并调用一个静态 API。

## 工作原理

1. **启动时发现。** `services/internal/plugin` 扫描 `plugins.dir`（本地开发默认
   `../runtime-plugins`），读取每个 `plugin.json` 并校验：id 必须等于目录名、版本必须是
   semver、路径必须是相对路径、后端模式必须合法。
2. **公开目录。** `GET /api/v1/plugins` 返回可见插件清单，包含 `entryUrl`、`styles`、
   `routes`、`navigations`。
3. **宿主加载。** 管理端在登录后加载插件，公共站在启动时加载。加载器动态 import
   `entryUrl`、注入样式、把路由加到名为 `root` 的路由下，并调用插件的 `activate(ctx)`。
4. **前端资源。** 通过 `/proxy/plugins/<id>/frontend/<file>` 提供，实际读取
   `<plugin>/dist/`。
5. **插件 API。** 请求 `/proxy/plugins/<id>/api/<path>`，由 `<plugin>/server/`（static）
   或上游服务（proxy）处理。

## 目录结构

```text
runtime-plugins/<id>/
├── plugin.json          # 清单；id 必须等于目录名
├── package.json         # 构建脚本与依赖
├── tsconfig.json
├── vite.config.ts       # definePluginConfig()
├── src/
│   ├── index.ts         # components + activate(ctx) / deactivate()
│   ├── context.ts       # 可选：把 ctx 暴露给组件
│   ├── style.css        # 构建为 dist/style.css
│   └── views/           # Vue SFC
└── server/              # 静态后端文件（可选）
```

## 清单

```json
{
  "id": "hello-panel",
  "name": "Hello Panel",
  "version": "0.1.0",
  "description": "A minimal runtime plugin built with Vuetify.",
  "icon": "mdi-hand-wave",
  "frontend": {
    "entry": "index.js",
    "devEntry": "http://localhost:5174/src/index.ts",
    "styles": ["style.css"],
    "routes": [
      {
        "path": "/plugins/hello-panel",
        "name": "plugin-hello-panel",
        "component": "HelloView",
        "permission": "",
        "nav": {
          "label": "Hello plugin",
          "icon": "mdi-hand-wave",
          "order": 60,
          "path": "/plugins/hello-panel",
          "permission": ""
        }
      }
    ]
  },
  "backend": {
    "mode": "static",
    "staticDir": "server",
    "permission": ""
  },
  "permissions": []
}
```

| 字段 | 含义 |
| --- | --- |
| `id` | 小写 kebab-case，必须等于目录名 |
| `version` | semver |
| `frontend.entry` | **相对 `dist/` 的路径**，例如 `index.js` |
| `frontend.devEntry` | `plugins.dev=true` 时使用的绝对 URL |
| `frontend.styles` | **相对 `dist/` 的路径**，例如 `style.css` |
| `frontend.routes[].component` | 模块 `components` 导出中的 key |
| `frontend.routes[].nav` | 可选的侧边栏入口 |
| `backend.mode` | `none`、`static` 或 `proxy` |
| `backend.staticDir` | static 模式的目录，默认 `server` |
| `backend.upstream` | proxy 模式的上游 URL |
| `backend.permission` | 调用插件 API 所需权限 |
| `permissions[]` | 写入数据库的权限码，用于可见性判断 |

**路径约定：** `entry` 与 `styles` 都相对插件的 `dist/` 目录。宿主会拼成
`/proxy/plugins/<id>/frontend/<path>`。例如 `entry: "index.js"` 对外的 URL 是
`/proxy/plugins/hello-panel/frontend/index.js`，磁盘文件是
`runtime-plugins/hello-panel/dist/index.js`。

## 构建

```bash
# 单个插件
pnpm --filter @scaffold/plugin-hello build

# 所有已安装插件
pnpm run build:plugins
```

示例会产出 `dist/index.js`（约 7 KB）与 `dist/style.css`。Vuetify 不会被打包，宿主已经
提供。

## 运行

```bash
pnpm run build:plugins
pnpm run dev:server      # :8000
pnpm run dev:admin       # http://localhost:5175/admin/
pnpm run dev:web         # http://localhost:5173
```

侧边栏会出现 **Hello plugin**。在管理端完成强制改密后可以调用插件 API；公共站会显示
“请登录管理端”的提示。

## 开发模式（HMR）

```bash
# 终端 1
SCAFFOLD_PLUGINS_DEV=true pnpm run dev:server

# 终端 2
pnpm --filter @scaffold/plugin-hello dev   # http://localhost:5174
```

`plugins.dev=true` 时，插件目录返回 `devEntry` 而不是构建后的 `entryUrl`，宿主直接从
Vite 加载。修改清单仍需重启后端，因为注册表在启动时加载。

## 新增自己的插件

1. 复制示例：`cp -r runtime-plugins/hello-panel runtime-plugins/my-plugin`。
2. 修改 `package.json#name`；在 `plugin.json` 中修改 `id`（必须等于目录名）、`name`、
   `version`、`frontend.routes[].path/name` 与 `nav`。
3. 如果 workspace 依赖有变化，运行 `pnpm install`。
4. 构建：`pnpm --filter <包名> build` 或 `pnpm run build:plugins`。
5. 重启后端并刷新应用。
6. 卸载时直接删除目录；没有 lockfile 或代码生成步骤。

## Vuetify

宿主全局安装了 Vuetify，因此插件 SFC 可以直接在模板中使用 `<VCard>`、`<VBtn>`、
`<VAlert>` 等组件，无需 import 或重复打包。示例用 `src/components.d.ts` 为 `vue-tsc`
声明这些全局组件。宿主还通过 `ctx.host.vuetifyLib` 暴露 Vuetify 的组合式 API（如
`useTheme`）。

## 后端模式

- **none** —— 没有 API，纯前端插件。
- **static** —— `server/` 下的文件通过 `/proxy/plugins/<id>/api/*` 提供。需要有效的
  access token，可选 `backend.permission`。
- **proxy** —— 转发到 `backend.upstream`。生产环境要求上游 host 出现在
  `plugins.allowed_upstreams` 中。代理会附加 `X-Scaffold-Plugin-Id`、
  `X-Scaffold-User-Id`、`X-Scaffold-Username`；`forwardAuthorization` 控制是否透传
  调用方的 `Authorization`；`timeout` 默认 30s。

## 鉴权与权限

- `GET /api/v1/plugins` 是公开的。插件 API 需要有效 access token。
- 仍需强制改密的用户调用插件 API 会得到 `MUST_CHANGE_PASSWORD`，与核心接口一致。
- 插件声明 `permissions` 时，非管理员用户至少拥有其中一个才能在目录中看到它。
  `backend.permission` 控制 API 调用。
- 前端权限检查只是展示层；插件 JavaScript 以宿主权限运行，插件 API 才是真正的边界。

## 限制

- EXPERIMENTAL：没有沙箱、签名校验或版本锁定。只安装可信插件。
- 修改清单需要重启后端。
- 一个 id 对应一个插件；id 必须等于目录名。
- 资源同源提供，没有 SSR。
- 多副本部署时每个副本都要有相同插件（打进镜像或挂载共享卷）。

## 常见问题

| 现象 | 检查 |
| --- | --- |
| 插件不出现 | 服务端日志 `skipping invalid plugin`；id 与目录名不一致、版本不是 semver、路径非法 |
| 资源 404 | `entry`/`styles` 相对 `dist/`；构建产物存在；URL 包含 `/frontend/` |
| API 401 | 没有 access token，先登录 |
| API 403 `MUST_CHANGE_PASSWORD` | 先完成强制改密 |
| API 403 `FORBIDDEN` | 未授予 `backend.permission` |
| `devEntry` 不生效 | `plugins.dev=true` / `SCAFFOLD_PLUGINS_DEV=true`；开发服务器在 5174 端口 |
| 组件没有渲染 | 宿主必须安装 Vuetify；插件模板可直接使用全局组件 |
