# hello-panel

[English](README.md) | [简体中文](README.zh-CN.md)

The built-in runtime plugin example. It demonstrates:

- registering a frontend route and sidebar entry through `plugin.json`;
- reusing the host's global Vuetify components without bundling Vuetify;
- calling a static backend under `server/` via
  `/proxy/plugins/hello-panel/api/hello.json`;
- loading from the Vite dev server through `devEntry` in development mode.

Build:

```bash
pnpm --filter @scaffold/plugin-hello build
# or build every installed plugin
pnpm run build:plugins
```

Develop (the backend must run with `SCAFFOLD_PLUGINS_DEV=true`):

```bash
pnpm --filter @scaffold/plugin-hello dev
```

Delete the whole directory to remove the example. See the parent
[runtime-plugins/README.md](../README.md) for the full guide.
