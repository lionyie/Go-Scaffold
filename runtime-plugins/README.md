# Runtime plugins

[English](README.md) | [简体中文](README.zh-CN.md)

> EXPERIMENTAL — not part of MVP acceptance.

A runtime plugin is a directory under `runtime-plugins/` that the server loads at
startup. It ships browser ESM, optional CSS, an optional static or proxy backend,
and a `plugin.json` manifest. Unlike the feature extensions in `extensions/`,
runtime plugins are not compiled into the Go backend or the app bundles.

A minimal, runnable example is included in
[`hello-panel/`](hello-panel/). It registers a route and sidebar entry, uses the
host's Vuetify components, and calls a static API endpoint.

## How it works

1. **Startup discovery.** `services/internal/plugin` scans `plugins.dir`
   (`../runtime-plugins` in local dev), reads each `plugin.json`, and validates it:
   the id must equal the directory name, the version must be semver, paths must be
   relative, and the backend mode must be valid.
2. **Public catalog.** `GET /api/v1/plugins` returns the visible manifests with
   `entryUrl`, `styles`, `routes`, and `navigations`.
3. **Host loading.** The admin app loads plugins after sign-in; the public web app
   loads them on startup. The loader dynamic-imports `entryUrl`, injects styles,
   adds routes to the `root` route, and calls the plugin's `activate(ctx)`.
4. **Frontend assets.** They are served at
   `/proxy/plugins/<id>/frontend/<file>` from `<plugin>/dist/`.
5. **Plugin API.** Calls go to `/proxy/plugins/<id>/api/<path>` and are served
   from `<plugin>/server/` (static) or proxied to an upstream.

## Directory layout

```text
runtime-plugins/<id>/
├── plugin.json          # manifest; id must equal the directory name
├── package.json         # build scripts and dependencies
├── tsconfig.json
├── vite.config.ts       # definePluginConfig()
├── src/
│   ├── index.ts         # components + activate(ctx) / deactivate()
│   ├── context.ts       # optional: expose ctx to components
│   ├── style.css        # extracted to dist/style.css
│   └── views/           # Vue SFCs
└── server/              # static backend files (optional)
```

## Manifest

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

| Field | Meaning |
| --- | --- |
| `id` | Lowercase kebab-case, must equal the directory name |
| `version` | Semver |
| `frontend.entry` | Path **inside `dist/`**, e.g. `index.js` |
| `frontend.devEntry` | Absolute URL used when `plugins.dev=true` |
| `frontend.styles` | Paths **inside `dist/`**, e.g. `style.css` |
| `frontend.routes[].component` | Key in the module's `components` export |
| `frontend.routes[].nav` | Optional sidebar entry |
| `backend.mode` | `none`, `static`, or `proxy` |
| `backend.staticDir` | Directory for `static` mode; defaults to `server` |
| `backend.upstream` | Upstream URL for `proxy` mode |
| `backend.permission` | Permission required to call the plugin API |
| `permissions[]` | Codes seeded into the DB; used for visibility |

**Path convention:** `entry` and `styles` are relative to the plugin's `dist/`
directory. The host builds the public URL as
`/proxy/plugins/<id>/frontend/<path>`. For example, `entry: "index.js"` maps to
`/proxy/plugins/hello-panel/frontend/index.js` on the wire and
`runtime-plugins/hello-panel/dist/index.js` on disk.

## Build

```bash
# one plugin
pnpm --filter @scaffold/plugin-hello build

# every installed plugin
pnpm run build:plugins
```

The example produces `dist/index.js` (about 7 KB) and `dist/style.css`. Vuetify is
not bundled: the host already provides it.

## Run

```bash
pnpm run build:plugins
pnpm run dev:server      # :8000
pnpm run dev:admin       # http://localhost:5175/admin/
pnpm run dev:web         # http://localhost:5173
```

The sidebar shows **Hello plugin**. In the admin console the API call works after
the forced password change; the public site shows a “sign in to the admin console”
hint instead.

## Develop with HMR

```bash
# terminal 1
SCAFFOLD_PLUGINS_DEV=true pnpm run dev:server

# terminal 2
pnpm --filter @scaffold/plugin-hello dev   # http://localhost:5174
```

With `plugins.dev=true` the catalog returns `devEntry` instead of the built
`entryUrl`, so the host imports directly from Vite. Manifest changes still require
a server restart because the registry is loaded at startup.

## Add your own plugin

1. Copy the example: `cp -r runtime-plugins/hello-panel runtime-plugins/my-plugin`.
2. Change `package.json#name`, and in `plugin.json` change `id` (must equal the
   directory name), `name`, `version`, `frontend.routes[].path/name`, and `nav`.
3. Run `pnpm install` if the workspace dependency graph changed.
4. Build: `pnpm --filter <package-name> build` or `pnpm run build:plugins`.
5. Restart the backend and reload the app.
6. To uninstall, delete the directory. There is no lockfile or codegen step.

## Vuetify

The host installs Vuetify globally, so plugin SFCs can use `<VCard>`, `<VBtn>`,
`<VAlert>` and friends directly in templates without importing or bundling them.
The example declares the global components in `src/components.d.ts` for
`vue-tsc`. The host also exposes `ctx.host.vuetifyLib` for composables such as
`useTheme`.

## Backend modes

- **none** — no API; the plugin is frontend-only.
- **static** — files under `server/` are served at
  `/proxy/plugins/<id>/api/*`. Requires a valid access token and an optional
  `backend.permission`.
- **proxy** — requests are forwarded to `backend.upstream`. In production the
  upstream host must be listed in `plugins.allowed_upstreams`. The proxy adds
  `X-Scaffold-Plugin-Id`, `X-Scaffold-User-Id`, and `X-Scaffold-Username`;
  `forwardAuthorization` controls whether the caller's `Authorization` header is
  forwarded. `timeout` defaults to 30s.

## Authentication and permissions

- `GET /api/v1/plugins` is public. Plugin API calls require a valid access token.
- A user who still must change their password gets `MUST_CHANGE_PASSWORD` from the
  plugin API, the same as core endpoints.
- If a plugin declares `permissions`, non-admin users need at least one of them to
  see the plugin in the catalog. `backend.permission` gates API calls.
- Frontend permission checks are cosmetic; plugin JavaScript runs with the host's
  privileges. The plugin API is the real boundary.

## Limitations

- EXPERIMENTAL: no sandbox, no signature verification, no version pinning. Install
  only plugins you trust.
- Manifest changes require a backend restart.
- One plugin per id; the id must match the directory.
- Assets are served same-origin; there is no SSR.
- Multi-replica deployments must include the same plugins on every replica
  (bake them into the image or mount a shared volume).

## Troubleshooting

| Symptom | Check |
| --- | --- |
| Plugin is missing | Server log: `skipping invalid plugin`; id/dir mismatch, bad semver, or invalid relative path |
| Asset 404 | `entry`/`styles` are relative to `dist/`; build output exists; URL includes `/frontend/` |
| API 401 | No access token; sign in first |
| API 403 `MUST_CHANGE_PASSWORD` | Complete the forced password change |
| API 403 `FORBIDDEN` | `backend.permission` is not granted |
| `devEntry` not used | `plugins.dev=true` / `SCAFFOLD_PLUGINS_DEV=true`; dev server on port 5174 |
| Component is not rendered | The host app must have Vuetify installed; plugin templates can use global components |
