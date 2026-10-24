# hello-panel

[English](README.md) | [简体中文](README.zh-CN.md)

内置的运行时插件示例。它演示：

- 通过 `plugin.json` 注册前端路由与侧边栏入口；
- 复用宿主提供的 Vuetify 全局组件（不重复打包）；
- 通过 `/proxy/plugins/hello-panel/api/hello.json` 调用 `server/` 下的静态后端；
- 开发模式下通过 `devEntry` 从 Vite 开发服务器加载。

构建：

```bash
pnpm --filter @scaffold/plugin-hello build
# 或构建全部已安装插件
pnpm run build:plugins
```

开发（需要后端以 `SCAFFOLD_PLUGINS_DEV=true` 启动）：

```bash
pnpm --filter @scaffold/plugin-hello dev
```

删除整个目录即可移除示例。完整说明见上级的
[runtime-plugins/README.zh-CN.md](../README.zh-CN.md)。
