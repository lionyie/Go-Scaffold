# Scaffold

[English](README.md) | [简体中文](README.zh-CN.md)

A multi-app starter for Go + TypeScript projects. It ships with an admin console, role-based access control (RBAC), on-demand sign-in extensions, and a script-based rename flow so it can become a real project without rewriting the tooling.

## What is in the box

- **Apps**: `apps/web` (public site) and `apps/admin` (management console), both Vue 3 + Vite + Vuetify 4 with a custom light/dark theme, glass surfaces, and page transitions.
- **UI system**: shared Vuetify theme, MDI icons, snackbar/confirm dialogs, and a forced password-change modal on first admin login.
- **Services**: one Go module under `services/` with Kratos HTTP/gRPC, protobuf contracts, GORM, JWT auth, RBAC, plugin hosting, and an embedded web UI.
- **Contracts**: `.proto` sources under `contracts/proto`, generated Go and TypeScript clients.
- **Extensions**: compile-time feature modules and runnable examples under `extensions/`, managed by `scripts/module.mjs`.
- **Runtime plugins**: an experimental loader under `runtime-plugins/` (not part of MVP acceptance).

## Structure

```text
apps/
  web/                  public site (port 5173)
  admin/                admin console (port 5175, base /admin/)
services/
  cmd/api/              Go entrypoint + wire
  api/                  generated Go protobuf code
  internal/
    biz/ data/ service/ identity usecases, repositories, transport adapters
    server/             HTTP/gRPC, authn/authz, OAuth callbacks
    auth/               JWT, password, OAuth provider registry
    extensions/         extension registries
    plugin/             runtime plugin loader (experimental)
    webui/              embedded web + admin builds
packages/               shared TypeScript libraries
contracts/              proto sources + permission catalog
extensions/             feature/example templates and installed lockfile
runtime-plugins/        experimental runtime plugins
deploy/                 Docker, Compose, Kubernetes
configs/                local/docker/k8s runtime config
scripts/                codegen, extensions, rename, build, clean
docs/                   architecture, development, auth, extensions, renaming
```

## Quick start

Prerequisites: Go 1.27+, Node 22.12+, pnpm 10+. `make` is optional.

```bash
pnpm run init          # install pinned codegen tools + workspace dependencies
pnpm run generate      # permissions, protobuf, OpenAPI/TS, wire

# terminal 1
pnpm run dev:server    # HTTP :8000, gRPC :9000

# terminal 2
pnpm run dev:web       # http://localhost:5173

# terminal 3
pnpm run dev:admin     # http://localhost:5175/admin/
```

Sign in with `admin / admin123`. The first login forces a password change; after that, the account can manage users, roles, and permissions.

## Common commands

`pnpm` scripts are the source of truth. `make <target>` is a thin alias.

| Goal | Command |
| --- | --- |
| Install tools and deps | `pnpm run init` |
| Regenerate contracts | `pnpm run generate` |
| Verify generated code | `pnpm run verify-generated` |
| Verify authz mapping | `pnpm run verify-authz` |
| Run backend | `pnpm run dev:server` |
| Run public web | `pnpm run dev:web` |
| Run admin | `pnpm run dev:admin` |
| Build everything | `pnpm run build` |
| Typecheck | `pnpm run typecheck` |
| Lint | `pnpm run lint:web` + `buf lint` |
| Reset local DB | `pnpm run db:reset -- --yes` |
| Install a feature | `pnpm run module:add -- oauth-github` |
| Install an example | `pnpm run module:add -- example-notes` |
| Remove an extension | `pnpm run module:remove -- <id>` |
| Check go-admin upstream | `pnpm run module:check-upstream` |
| Rename the project | `pnpm run rename -- --name myapp --yes` |

## Third-party sign-in

OAuth providers are feature extensions. Install only the ones you need:

```bash
pnpm run module:add -- oauth-github
pnpm run module:add -- oauth-google
pnpm run module:add -- oauth-wechat
pnpm run module:add -- oauth-qq
pnpm run module:add -- oauth-oidc
```

Each provider reads environment variables (for example `SCAFFOLD_OAUTH_GITHUB_CLIENT_ID` and `_CLIENT_SECRET`) and appears on the admin login page when configured. See [docs/auth.md](docs/auth.md).

## Runtime plugins (experimental)

A runnable example lives in `runtime-plugins/hello-panel/`. It registers a route
and sidebar entry, uses the host's Vuetify components without bundling Vuetify,
and calls a static API endpoint through the authenticated plugin proxy.

```bash
pnpm run build:plugins                       # build every installed plugin
pnpm run dev:server                          # :8000
pnpm run dev:admin                           # sidebar → Hello plugin

# HMR for plugin source
SCAFFOLD_PLUGINS_DEV=true pnpm run dev:server
pnpm --filter @scaffold/plugin-hello dev      # :5174
```

Delete the plugin directory to uninstall it. See
[runtime-plugins/README.md](runtime-plugins/README.md) for the manifest, path
conventions, backend modes, auth and troubleshooting.

## Deployment

```bash
pnpm run build
node scripts/db-reset.mjs --yes
docker compose -f deploy/compose/docker-compose.yml up --build
kubectl apply -f deploy/k8s/
```

The single binary embeds both SPAs: `/` serves the public app, `/admin/` serves the admin app.

## Documentation

The full index lives in [docs/README.md](docs/README.md).

- [Architecture](docs/architecture.md)
- [Development](docs/development.md)
- [Authentication and SSO](docs/auth.md)
- [Extensions](docs/extensions.md)
- [Renaming](docs/renaming.md)
- [Optimization log](docs/scaffold-optimization-log.md)
