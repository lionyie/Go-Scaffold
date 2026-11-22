# Plugin and Extension Author Guide

[English](plugin-author.md) | [简体中文](plugin-author.zh-CN.md)

> Who this is for: developers adding new functionality to the scaffold as a compile-time extension or an experimental runtime plugin.

## Two Models Compared

| Aspect | Compile-time extension | Runtime plugin (experimental) |
| --- | --- | --- |
| Lives in | `extensions/templates/<id>/` | `runtime-plugins/<id>/` |
| Installed by | `pnpm run module:add -- <id>` (copies files into the tree) | Dropping the directory in; server loads it at startup |
| Manifest | `module.json` | `plugin.json` |
| Can change core/generated code | Yes — `copy` targets anywhere in the repo, plus codegen | No — ships browser ESM + CSS + optional static/proxy backend |
| Backend logic | Real Go code compiled into the binary | Static files under `server/`, or a proxied upstream |
| Frontend | Source merged into the apps, routes/menu generated into `features.gen.ts` | Loaded dynamically from `/proxy/plugins/<id>/frontend/` |
| Install/remove without rebuild | No | Yes for frontend/assets; manifest changes need a server restart |
| Maturity | Supported workflow with lockfile and rollback | EXPERIMENTAL: no sandbox, no signature verification |

## Choosing Between Them

- Need to modify or add **core Go services, protos, or generated code**? → compile-time extension.
- Want **independent HMR** while developing, without rebuilding the host apps? → runtime plugin (`SCAFFOLD_PLUGINS_DEV=true`).
- Need to **install or remove a feature without rebuilding** the Go binary and app bundles? → runtime plugin (restart only reloads the manifest).
- Need a **real permission boundary** backed by the RBAC catalog and `verify-authz`? → compile-time extension; runtime-plugin permissions are enforced only at the `/proxy/plugins/<id>/api/` boundary.
- Just adding a page, a sidebar entry, or a simple read-only API? → start with a runtime plugin; promote to an extension when you outgrow it.

## Runtime Plugins

Each plugin is a directory under `runtime-plugins/` whose name must equal the manifest `id`. The example `runtime-plugins/hello-panel/` is runnable as-is.

### plugin.json

Key fields (full table in the [runtime plugins README](../../runtime-plugins/README.md)):

| Field | Meaning |
| --- | --- |
| `id` | Lowercase kebab-case; must equal the directory name |
| `version` | Semver |
| `frontend.entry` | Path inside `dist/`, e.g. `index.js` |
| `frontend.devEntry` | Absolute URL used when `plugins.dev=true` |
| `frontend.styles` | Paths inside `dist/` |
| `frontend.routes[]` | Route records; `component` is a key in the module's `components` export; optional `nav` adds a sidebar entry |
| `backend.mode` | `none`, `static`, or `proxy` |
| `backend.staticDir` | Directory for `static` mode; defaults to `server` |
| `backend.upstream` | Upstream URL for `proxy` mode |
| `backend.permission` | Permission required to call the plugin API |
| `permissions[]` | Codes seeded into the DB; gate catalog visibility |

### Serving and API

- Frontend assets: `GET /proxy/plugins/<id>/frontend/<file>` is served from `<plugin>/dist/`. `entry: "index.js"` maps to `/proxy/plugins/hello-panel/frontend/index.js`.
- Plugin API: `GET /proxy/plugins/<id>/api/<path>` is served from `server/` (static mode) or forwarded to `backend.upstream` (proxy mode).
- These routes bypass the Kratos middleware chain and authenticate themselves: a valid access token is required, users who must change their password get `MUST_CHANGE_PASSWORD`, and `backend.permission` gates calls with `FORBIDDEN`.
- The proxy adds `X-Scaffold-Plugin-Id`, `X-Scaffold-User-Id`, and `X-Scaffold-Username`; the default timeout is 30s. In production the upstream host must be listed in `plugins.allowed_upstreams`.

### Build and develop

```bash
pnpm --filter @scaffold/plugin-hello build   # one plugin
pnpm run build:plugins                       # every installed plugin

# HMR: two terminals
SCAFFOLD_PLUGINS_DEV=true pnpm run dev:server
pnpm --filter @scaffold/plugin-hello dev     # http://localhost:5174
```

With `plugins.dev=true` the catalog returns `devEntry` instead of the built `entryUrl`, so the host imports straight from Vite. Manifest changes still require a backend restart because the registry is loaded at startup.

### Host integration

- The host installs Vuetify globally; plugin SFCs use `<VCard>`, `<VBtn>`, etc. directly without bundling Vuetify.
- The host exposes globals through `window.__SCAFFOLD_PLUGIN_HOST__` (see `packages/plugin-runtime/src/host.ts`), and `ctx.host.vuetifyLib` gives composables such as `useTheme`.
- To uninstall, delete the directory. There is no lockfile or codegen step.

## Compile-Time Extensions

Extensions are templates in `extensions/templates/` copied into the tree by `scripts/module.mjs` (pnpm aliases `module:add|module:remove|module:update|module:list|module:check-upstream`).

### module.json

See `extensions/templates/example-notes/module.json` for a complete example. Important fields:

- `copy`: `{ from, to }` pairs relative to the template and the repo root.
- `registries`: Go files copied under `services/internal/extensions` are blank-imported by `registry_gen`; OAuth providers under `services/internal/auth/oauth/providers` are blank-imported by `oauthproviders_gen`.
- `authz` / `authzImport` / `authzAlias`: entries are written to `services/internal/server/authz/extensions_gen.go` so `pnpm run verify-authz` sees them.
- `targets`: routes and menu items written to `apps/{web,admin}/src/router/features.gen.ts` (never hand-edit these files).
- `codegen: true`: runs the full `pnpm run generate` pipeline after install/remove.
- `generatedDirs`: removed on uninstall.

### Safety and lifecycle

- `extensions/installed.lock.json` records the template version, `baseHash`, per-file SHA-256 hashes, generated files, and targets.
- `add` validates dependencies and conflicts, stages files in `.scaffold-tmp/`, and rolls back created files on failure.
- `remove` refuses to delete files you modified unless `--force`, and refuses to remove an extension another one requires.
- `update` overwrites unchanged files; files you modified are written next to the target as `<file>.new` with a diff summary.
- Concurrent commands are rejected through `extensions/.lock`; every command supports `--dry-run` and `--json`.
- `check-upstream` compares against the pinned go-admin reference in `extensions/upstream.json` and writes a (gitignored) report to `extensions/upstream-report.md`; treat it as a manual migration checklist.

## Permissions

- A runtime plugin's `permissions[]` codes are seeded into the DB and only control catalog visibility; `backend.permission` is the real API gate. Frontend permission checks in a plugin are cosmetic — plugin JavaScript runs with the host's privileges.
- If your feature needs a **new** permission code (typical for extensions):
  - [ ] Add it to `contracts/permissions.yaml`.
  - [ ] Regenerate so it lands in `services/internal/pkg/rbac/permissions_gen.go` and `packages/shared/src/permissions.gen.ts`.
  - [ ] Map operations in `services/internal/server/authz/mapping.go` (or the extension's `authz` block).
  - [ ] Run `pnpm run verify-authz` — unknown operations fail closed.

## Testing and Verification

- [ ] Build: `pnpm run build:plugins` and confirm `dist/index.js` + `dist/style.css` exist.
- [ ] Catalog: `GET /api/v1/plugins` returns your manifest with `entryUrl` and routes.
- [ ] Asset: `GET /proxy/plugins/<id>/frontend/index.js` returns 200.
- [ ] API: unauthenticated call → 401; authenticated without `backend.permission` → 403 `FORBIDDEN`; granted → 200.
- [ ] Extension install: `pnpm run module:add -- example-notes`, then `go build ./...` in `services/` and `pnpm run typecheck`.
- [ ] `pnpm run module:list` shows the extension; `remove` restores a clean tree.
- [ ] CI runs the same checks: codegen + `verify-authz` + `git diff`, `go vet` + `go test -race`, and an extension add/remove cycle.

## Limitations

- Runtime plugins are EXPERIMENTAL: no sandbox, no signature verification, no version pinning — install only plugins you trust.
- Host-provided Vue, Vuetify, Pinia, and vue-i18n must **not** be bundled into the plugin; they are externalized via `packages/config-vite/plugin.ts`.
- Manifest changes require a backend restart; one plugin per id.
- Assets are served same-origin; there is no SSR.
- Multi-replica deployments must ship the same plugins on every replica (bake into the image or mount a shared volume). Single-replica assumptions also apply to the plugin registry, which is loaded once at startup.

## Related

- [Extensions](../extensions.md) — module.mjs commands, lockfile, bundled templates
- [Runtime plugins README](../../runtime-plugins/README.md) — manifest reference and troubleshooting
- [Authentication, RBAC and SSO](../auth.md) — permission model behind `backend.permission`
- [Architecture](../architecture.md) — generated outputs and boundary rules
- [Development](../development.md) — codegen pipeline and daily workflow
