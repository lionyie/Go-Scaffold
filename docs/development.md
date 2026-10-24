# Development

[English](development.md) | [简体中文](development.zh-CN.md)

## First run

```bash
pnpm run init
pnpm run generate
```

`init` installs pinned codegen tools and workspace dependencies. It does not
require `make`.

## Daily loop

Use three terminals:

```bash
pnpm run dev:server   # :8000 / :9000
pnpm run dev:web      # http://localhost:5173
pnpm run dev:admin    # http://localhost:5175/admin/
```

The seeded administrator is `admin / admin123`. The first login opens a
persistent change-password dialog. Until it is completed, authenticated-only
endpoints (`/auth/me`, logout, change password) work, while permissioned
endpoints return `MUST_CHANGE_PASSWORD`.

## Add a backend module

1. Define the contract in `contracts/proto/<domain>/v1/*.proto` with
   `google.api.http` annotations.
2. Run `pnpm run generate`.
3. Implement the usecase in `services/internal/biz/<domain>.go` and the
   repository in `services/internal/data/database/<domain>.go`.
4. Add the transport adapter in `services/internal/service/<domain>.go` and
   register it in `services/internal/service/service.go` and
   `services/internal/server/http.go` / `grpc.go`.
5. Add the operation-to-permission mapping in
   `services/internal/server/authz/mapping.go`. `pnpm run verify-authz` fails
   when an operation is missing or a permission code does not exist.
6. Build the admin page under `apps/admin/src/views` and add the route/menu
   entry with `Permission` constants from `@scaffold/shared`.

## Code generation

`pnpm run generate`:

1. clears codegen outputs (preserving `services/api/identity/v1/redact.go`),
2. regenerates permission constants,
3. runs the three buf templates,
4. regenerates the TypeScript schema,
5. regenerates wire,
6. runs `go mod tidy`,
7. runs `verify-authz`.

The pipeline must be idempotent: running it twice should leave the tree
unchanged. CI enforces this with `git diff --exit-code`.

## Extensions and examples

```bash
pnpm run module:add -- oauth-github
pnpm run module:add -- example-notes
pnpm run module:list
pnpm run module:remove -- example-notes
```

See [extensions.md](extensions.md). Installing or removing an extension rewrites
the generated registries and runs codegen when the extension declares
`"codegen": true`.

## Database

The local profile uses SQLite at `data/scaffold.db` relative to the repository
root. Schema changes that remove columns require a reset:

```bash
pnpm run db:reset -- --yes
```

`services/internal/data/migrations/` is read before `AutoMigrate`. It is empty
today; see its README for the future `golang-migrate` naming convention.

## Runtime plugin development

```bash
pnpm run build:plugins
pnpm run dev:server
pnpm run dev:admin        # sidebar → Hello plugin
```

For HMR, run the backend with `SCAFFOLD_PLUGINS_DEV=true` and start the plugin
Vite server:

```bash
SCAFFOLD_PLUGINS_DEV=true pnpm run dev:server
pnpm --filter @scaffold/plugin-hello dev
```

See [runtime-plugins/README.md](../runtime-plugins/README.md) for the manifest,
path convention, backend modes, auth and troubleshooting.

## Tests and lint

```bash
pnpm run verify-authz
pnpm run verify-generated
cd services && go vet ./... && go test ./...
pnpm run typecheck
pnpm run lint:web
buf lint
```

`go test -race` requires cgo. On machines without a C compiler, run without
`-race`; CI runs it on Linux.

## Production notes

- Set `SCAFFOLD_JWT_SECRET` and `SCAFFOLD_OAUTH_STATE_SECRET` to non-default
  values. Production startup fails fast otherwise.
- Set `SCAFFOLD_SEED_ADMIN_PASSWORD` before the first production start, or change
  the seeded password interactively. Production refuses to start while the
  seeded password still requires rotation.
- The in-memory OAuth store is single-replica only. Replace it with Redis or a
  database implementation before scaling out.
