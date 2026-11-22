# Onboarding: Full-stack Quick Start

[English](onboarding.md) | [简体中文](onboarding.zh-CN.md)

> Who this is for: a new contributor who needs to clone this monorepo, get the Go backend plus both Vue SPAs running locally, and land a first safe change on day one.

## Day-one roadmap

- [ ] Install the prerequisites (Go, Node, pnpm) and clone the repository.
- [ ] Run `pnpm run init` to install workspace dependencies and pinned codegen tools.
- [ ] Run `pnpm run generate` to regenerate contracts, permission constants, and the API client.
- [ ] Start three terminals: `pnpm run dev:server`, `pnpm run dev:web`, `pnpm run dev:admin`.
- [ ] Log in to the admin console as `admin / admin123` and complete the forced password change.
- [ ] Browse `http://localhost:5173` (public site) and `http://localhost:5175/admin/` (console).
- [ ] Make one small frontend change and one small backend change (see below).
- [ ] Run the quality gates locally before opening your first pull request.

## Prerequisites and versions

| Tool | Required version | Notes |
| --- | --- | --- |
| Go | 1.27+ | `services/go.mod` declares `go 1.27.0`. |
| Node | 22.12+ | `package.json` `engines`; the repo `.nvmrc` pins 24. |
| pnpm | 10 | Pinned via `packageManager: pnpm@10.33.2`; enable with `corepack enable`. |
| make | optional | Not required; all workflows run through pnpm scripts and `scripts/*.mjs`. |

## Five-minute startup

```bash
pnpm run init        # install deps + pinned codegen tools
pnpm run generate    # proto -> Go/TS code, permission constants, OpenAPI schema
```

Then open three terminals:

```bash
pnpm run dev:server   # Go backend: HTTP :8000, gRPC :9000
pnpm run dev:web      # public SPA:   http://localhost:5173
pnpm run dev:admin    # admin SPA:    http://localhost:5175/admin/
```

Each Vite dev server proxies `/api`, `/proxy`, `/healthz`, and `/readyz` to `http://127.0.0.1:8000`, so frontend and backend stay same-origin with no CORS setup. Backend defaults come from `configs/config.local.yaml`: SQLite at `data/scaffold.db` (created automatically), HTTP `0.0.0.0:8000`, gRPC `:9000`.

## Login and forced password change

The seed creates an administrator `admin / admin123` (`services/internal/data/database/seed.go`). On first login:

1. The admin guard sees the account is flagged for rotation and opens a persistent change-password dialog.
2. Until you finish it, only authenticated self-service endpoints (`/auth/me`, logout, change password) work; permissioned endpoints return the error reason `MUST_CHANGE_PASSWORD`.
3. After changing the password the full console unlocks. In production, startup refuses to run while the seeded password is unrotated; set `SCAFFOLD_SEED_ADMIN_PASSWORD` before first boot.

Login is rate limited to 5 failures per IP+username per minute (HTTP 429), so avoid scripted retries against the dev server.

## Repository map

| Path | What lives here |
| --- | --- |
| `apps/` | Two Vue 3.5 + Vite 8 + Vuetify 4 SPAs: `apps/web` (base `/`) and `apps/admin` (base `/admin/`). |
| `services/` | Go (Kratos v3) backend; entry `services/cmd/api/main.go`; layers `server -> service -> biz -> data` under `services/internal/`. |
| `packages/` | Shared TS: `api-client`, `ui`, `shared`, `config-vite`, `plugin-runtime`. |
| `contracts/` | Source of truth: `proto/**/v1/*.proto` plus `permissions.yaml`. |
| `extensions/` | Compile-time extension templates and `installed.lock.json`. |
| `runtime-plugins/` | Runtime plugins (e.g. `hello-panel`) served under `/proxy/plugins/`. |
| `deploy/` | `docker/Dockerfile`, `compose/docker-compose.yml`, `k8s/*.yaml`. |
| `configs/` | Backend profiles: `config.local.yaml`, `docker.yaml`, `k8s.yaml`. |
| `scripts/` | Repo automation: `init.mjs`, `generate.mjs`, `module.mjs`, `dev-server.mjs`, etc. |
| `docs/` | Documentation, including this role series under `docs/roles/`. |

## First safe changes

**Smallest safe frontend change:** edit a string in `apps/web/src/i18n/locales/zh-CN.ts` (and the matching key in `en-US.ts`), then reload `http://localhost:5173`. Vite HMR applies it instantly, and nothing generated is involved.

**Smallest safe backend change:** add a log line or tweak an error message in an existing usecase such as `services/internal/biz/user.go`, then restart `pnpm run dev:server`. No codegen is needed because you are not touching proto contracts.

Things that *do* require `pnpm run generate`: editing `contracts/proto/**/v1/*.proto` or `contracts/permissions.yaml`. Never hand-edit generated outputs (`services/api/**`, `services/internal/conf/conf.pb.go`, `permissions_gen.go`, `packages/api-client/src/generated/**`, `apps/*/src/router/features.gen.ts`).

## Quality gates

Run these before pushing; CI enforces all of them:

```bash
pnpm run typecheck      # turbo typecheck across apps/packages
pnpm run lint:web       # eslint over apps, packages, extensions, scripts
pnpm run verify-authz   # authz mapping vs permission catalog
go -C services test ./...   # Go unit tests (add -race when cgo is available)
```

`pnpm run generate` must also be idempotent; CI runs it and fails on any `git diff`.

## Common pitfalls

- **Port conflicts:** 8000 (backend HTTP), 9000 (gRPC), 5173 (web), 5175 (admin). Kill stale processes before restarting; the plugin dev server uses 5174 when HMR is enabled.
- **Editing generated files:** changes under the generated paths listed above are wiped by `pnpm run generate`. Edit the contract, then regenerate.
- **`go test -race`:** requires cgo. On machines without a C compiler, run plain `go test ./...`; CI runs the race build on Linux.
- **First login "stuck":** the forced password dialog blocks permissioned pages by design; finish the change instead of trying to bypass it.
- **64-bit integers in JSON:** responses use protojson, so int64 fields arrive as strings and enums as names — do not `parseInt` them blindly on the frontend.

## Glossary

| Term | Meaning here |
| --- | --- |
| SPA | Single-page application; this repo ships two (`apps/web`, `apps/admin`). |
| Proto contract | `contracts/proto/**/v1/*.proto` files that define RPCs and HTTP bindings. |
| Generated | Any file produced by `pnpm run generate`; treat as read-only. |
| Extension | Compile-time module under `extensions/templates/`, installed via `pnpm run module:add`. |
| Runtime plugin | Drop-in plugin under `runtime-plugins/`, loaded at runtime via `/proxy/plugins/`. |
| RBAC | Role-based access control; catalog in `contracts/permissions.yaml`. |
| Authz mapping | Operation-to-permission map in `services/internal/server/authz/mapping.go`. |
| Single binary | The built Go server embeds both SPAs: `/` serves web, `/admin/` serves admin. |

## Related

- [backend.md](backend.md)
- [frontend-ui.md](frontend-ui.md)
- [devops.md](devops.md)
- [Development guide](../development.md)
