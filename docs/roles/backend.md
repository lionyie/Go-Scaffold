# Backend Developer Guide

[English](backend.md) | [简体中文](backend.zh-CN.md)

> Who this is for: developers working on the Go (Kratos v3) backend under `services/`, including protobuf contracts, business logic, persistence, and auth.

## Responsibilities & Boundaries

The backend follows a strict downward dependency chain:

```text
services/internal/server  ->  service  ->  biz  ->  data
```

- `server`: Kratos HTTP/gRPC transport, middleware, authn/authz, OAuth callbacks.
- `service`: thin adapters translating proto messages to biz calls and back.
- `biz`: usecases and repository interfaces (no storage details).
- `data`: GORM repositories, seeding, health checks.

Hard rules:

- Dependencies point downward only; never import `server` or `service` from `biz`/`data`.
- Core code must not import a specific extension; extensions depend on `biz`, `data`, `server/authz`, and `extensions/registry`, never the reverse.
- `contracts/` holds `.proto` files and `permissions.yaml` only, no implementation code.

## Environment & Running

```bash
pnpm run init        # one-time: pinned codegen tools + workspace deps
pnpm run generate    # regenerate all contract outputs
pnpm run dev:server  # HTTP :8000, gRPC :9000
```

- Entry point: `services/cmd/api/main.go` (Wire DI in `services/cmd/api/wire.go`).
- Config: `configs/config.local.yaml`; environment override prefix `SCAFFOLD_`, placeholders use `${NAME:default}`.
- Defaults: HTTP `0.0.0.0:8000`, gRPC `0.0.0.0:9000`, SQLite at `file:../data/scaffold.db` with WAL and foreign keys on.
- Switching to MySQL is a config change (`SCAFFOLD_DB_DRIVER=mysql`, `SCAFFOLD_DB_SOURCE=...`); no code change.
- Seeded admin is `admin / admin123` and is forced to change password on first login.

## Directory Map

| Path | Contents |
| --- | --- |
| `services/cmd/api` | binary entry, Wire setup |
| `services/internal/server` | HTTP/gRPC servers, codec, middleware, `authz/mapping.go` |
| `services/internal/service` | proto <-> biz adapters, `Services` aggregate |
| `services/internal/biz` | usecases, repository interfaces, domain errors |
| `services/internal/data` | GORM repos (`database/`), models (`model/`), migrations hook |
| `services/internal/auth` | JWT/password helpers, OAuth (`oauth/`) |
| `services/internal/plugin` | runtime plugin loader and registry |
| `services/internal/webui` | embedded SPA handler (`/`, `/admin/`) |
| `services/internal/pkg` | shared helpers: `auth`, `rbac`, `id`, `pagination` |
| `contracts/proto/**/v1` | protobuf source of truth |

## Request Lifecycle

1. HTTP middleware runs in this order (`services/internal/server/http.go`):
   `recovery -> logging -> LoginRateLimit -> Authn -> Authz`.
2. Bodies are decoded with protojson (`services/internal/server/codec.go`):
   camelCase field names, enum names (not numbers), 64-bit integers as strings,
   unknown fields discarded, `EmitUnpopulated` on responses.
3. `Authn` validates the JWT access token; `Authz` looks up the generated
   operation name in `authz/mapping.go`. Unknown operations fail closed.
4. `/proxy/plugins/` and `/api/v1/auth/oauth/` are registered with
   `HandlePrefix` and bypass the Kratos middleware chain; they authenticate
   themselves.

## Adding a Backend Module

1. Define the contract in `contracts/proto/<domain>/v1/*.proto` with `google.api.http` annotations.
2. Run `pnpm run generate` to refresh `services/api/**`, OpenAPI, and the TS schema.
3. Implement the usecase in `services/internal/biz/<domain>.go` and declare the repository interface there.
4. Implement the repository in `services/internal/data/database/<domain>.go` against the model in `services/internal/data/model/`.
5. Add the transport adapter in `services/internal/service/<domain>.go`.
6. Register the service in `services/internal/service/service.go` (`Services` struct + `ProviderSet`) and in `services/internal/server/http.go` / `grpc.go`.
7. Add every new operation to `services/internal/server/authz/mapping.go` (`Public`, authenticated-only, or a `rbac.Perm*` code).
8. If new permission codes are needed, add them to `contracts/permissions.yaml` and regenerate.
9. Run `pnpm run verify-authz`; it fails when an operation is unmapped or a code does not exist.
10. Run `go vet ./...` and `go test ./...` under `services/`.

## API Reference (Swagger UI)

`pnpm run generate` also produces the OpenAPI document embedded in the backend.
In non-production environments the server exposes it at `/swagger/`:

- UI: `http://localhost:8000/swagger/`
- Spec: `http://localhost:8000/swagger/openapi.yaml`

New RPCs with `google.api.http` annotations appear there after regeneration. The
endpoint is force-disabled in production and can be toggled with
`SCAFFOLD_SWAGGER_ENABLED`; see
[Development · API reference](../development.md#api-reference-and-swagger-ui).

## Data & Migrations

- GORM `AutoMigrate` runs at startup when `data.database.auto_migrate` is true (see `services/internal/data/database/data.go`); new persistent types are added to the `AutoMigrate` call there.
- SQLite connections are capped at a single writer to avoid busy errors.
- `services/internal/data/migrations/` is intentionally empty; the startup hook reads it and warns when `.sql` files exist but no migrator is enabled. The future `golang-migrate` layout (`NNNN_name.up.sql` / `NNNN_name.down.sql`) is documented in `services/internal/data/migrations/README.md`.
- Seeding (`data.database.seed`) creates the default admin, roles, and permission bindings.

## Auth & RBAC

- JWT access tokens live 15 minutes; refresh tokens live 7 days and are stored hashed. Passwords use bcrypt.
- Login is rate limited: 5 failures per IP+username per minute returns 429.
- The permission catalog is `contracts/permissions.yaml`, generated into `services/internal/pkg/rbac/permissions_gen.go` (Go) and `packages/shared/src/permissions.gen.ts` (TS).
- `is_admin` bypasses all permission checks; every other principal goes through the Authz map, and unknown operations fail closed.
- The seeded admin must change its password before permissioned endpoints unlock; in production, default JWT/OAuth secrets and an unrotated seed password fail startup (`APP_ENV=production`).

## Do-Not-Edit List

Generated files — regenerate instead of editing:

- `services/api/**` (sole exception: `services/api/identity/v1/redact.go`)
- `services/internal/conf/conf.pb.go`
- `services/internal/pkg/rbac/permissions_gen.go`
- `services/internal/auth/oauthproviders_gen/`
- `services/internal/server/authz/extensions_gen.go`
- `services/internal/extensions/registry_gen/`
- `services/cmd/api/wire_gen.go`
- `packages/api-client/generated/**`, `packages/api-client/src/generated/**`
- `packages/shared/src/permissions.gen.ts`, `apps/*/src/router/features.gen.ts`

## Tests & Quality Gates

- Tests exist only on the Go side, e.g. `services/internal/auth/oauth/oauth_test.go`, `services/internal/pkg/rbac/permissions_consistency_test.go`, `services/internal/server/middleware/authz_test.go`, `services/internal/data/database/data_test.go`, `services/internal/webui/fallback_test.go`, `services/internal/plugin/registry_test.go`.
- Local gates: `go vet ./...` and `go test ./...` from `services/`. CI runs `go test -race ./...`; run `-race` locally before pushing concurrency-sensitive changes.
- CI (`.github/workflows/ci.yml`): `codegen` (generate + verify-authz + drift diff), `go` (vet + test -race), `web`, `extensions`, `docker`.

## Common Pitfalls

- Forgetting the authz mapping: a new RPC without an entry in `authz/mapping.go` fails closed and looks like a permission bug.
- protojson wire format: clients send/receive camelCase and int64-as-string; do not hand-roll `encoding/json` DTOs next to proto types.
- Plugin proxy and OAuth routes skip the Kratos middleware chain; never assume `Authn` ran for handlers under `/proxy/plugins/` or `/api/v1/auth/oauth/`.
- Editing generated files: CI drift check fails; fix the `.proto` or `permissions.yaml` and regenerate.
- SQLite write contention: keep writes short; the pool is already pinned to one connection.

## Related

- [Architecture](../architecture.md)
- [Authentication & Authorization](../auth.md)
- [Development Workflow](../development.md)
- [Architecture Lead Guide](architecture-lead.md)
