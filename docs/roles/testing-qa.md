# Testing and QA Guide

[English](testing-qa.md) | [简体中文](testing-qa.zh-CN.md)

> Who this is for: engineers and QA staff who run the test suites, watch CI gates, and report defects for this repository.

## Current state

Automated coverage exists **only in Go**. There are exactly six test files under
`services/`, and **zero frontend tests**: every `test` script in `apps/*` and
`packages/*` is `"echo no-tests"`, and `pnpm run test:web` runs those stubs via
Turbo. CI passes without any UI test ever executing. Treat "all green" as
"Go tests + lint + typecheck passed", not "the product works".

## Test map

| File | What it covers |
| --- | --- |
| `services/internal/auth/oauth/oauth_test.go` | OAuth state expiry, one-time ticket replay, state signature, PKCE, account linking by email |
| `services/internal/pkg/rbac/permissions_consistency_test.go` | Generated `permissions_gen.go` catalog matches `contracts/permissions.yaml` |
| `services/internal/webui/fallback_test.go` | SPA fallback for `/` and `/admin/`, reserved API prefixes, `/admin` redirect |
| `services/internal/data/database/data_test.go` | SQLite schema migration creates `users` / `oauth_accounts`; migration runner accepts an empty directory |
| `services/internal/server/middleware/authz_test.go` | Authz middleware: public routes, unknown operation fails closed, missing principal, allow/deny, admin bypass, forced password change |
| `services/internal/plugin/registry_test.go` | Plugin registry: asset URLs under `/proxy/plugins/<id>/frontend/`, dev-mode entry |

## How to run

```bash
# All Go tests (from repo root)
go -C services test ./...

# Race detector (requires cgo; on Windows you need gcc, e.g. via MSYS2/MinGW)
go -C services test -race ./...

# Verify permission catalog, authz mapping, and frontend codes stay in sync
pnpm run verify-authz

# Verify generated code matches the source contracts (proto, permissions, modules)
pnpm run verify-generated
```

CI runs `go test -race ./...` on Linux, so a local run without `-race` is fine
only as a smoke check.

## CI gates

Defined in `.github/workflows/ci.yml`. Every push to `main` and every PR runs:

| Job | Runs | What a failure means |
| --- | --- | --- |
| `codegen` | `pnpm run generate`, `pnpm run verify-authz`, then `git diff --exit-code` | You edited a generated file by hand, or forgot to rerun codegen after changing a proto / permission / module |
| `go` | `go vet ./...` and `go test -race ./...` in `services/` | Compile error, vet violation, failing test, or a data race |
| `web` | `lint:web`, `typecheck`, `build:web`, `build:admin` | ESLint error, TS type error, or a broken Vite build |
| `extensions` | `module.mjs add/remove` for `oauth-github` and `example-notes`, plus Go builds | The module lifecycle is broken or leaves uncommitted drift |
| `docker` | Builds `deploy/docker/Dockerfile` (no push, main only) | The production image no longer builds |

`git diff --exit-code` is the key trick: after codegen, the working tree must
be clean, so stale generated files fail CI.

## The frontend test gap

There is no test runner configured anywhere in the frontend: no Vitest, no
Playwright, no Cypress. The stub scripts (`"test": "echo no-tests"`) exist so
`turbo run test` exits 0. Adding real tests means choosing and wiring a runner
first:

- Unit/component tests: add Vitest to the workspace, a `test` script per
  package, and a Turbo task so `pnpm run test:web` executes them.
- End-to-end tests: add Playwright with the dev servers (5173/5175) or the Go
  binary (8000) as the target, and a new CI job.

Until that exists, do not claim UI coverage — use the manual checklist below.

## Manual acceptance checklist

Run this against a fresh local instance (`pnpm run dev:server`, web on 5173,
admin on 5175, or the built binary on 8000):

- [ ] Log in as the seeded admin (`admin` / `admin123`) — you are forced to change the password before any other page loads
- [ ] After the password change, admin pages load and the user list renders
- [ ] Create a non-admin user with no roles — visiting a permission-guarded admin route redirects/denies, and direct API calls return 403
- [ ] Grant a role with one permission — exactly that screen unlocks, nothing more
- [ ] OAuth: start the GitHub/WeChat flow, complete the callback, verify the ticket exchange succeeds once and a second exchange of the same ticket fails
- [ ] Install a runtime plugin — its menu entry appears, its panel loads from `/proxy/plugins/<id>/frontend/`
- [ ] Reload on a deep route such as `/admin/users` — the SPA fallback serves `index.html`, not a 404
- [ ] `curl http://127.0.0.1:8000/healthz` and `/readyz` both return 200
- [ ] Enter a wrong password 5 times in a row for the same account — the 6th attempt returns 429 (login rate limit)

## API Reference (Swagger UI) smoke check

- [ ] In development, `GET /swagger/` returns 200 and renders Swagger UI.
- [ ] `GET /swagger/openapi.yaml` returns 200 and contains `openapi: 3.0.3`.
- [ ] With `APP_ENV=production`, `/swagger/` and `/swagger/openapi.yaml` return 404.
- [ ] Disabling `SCAFFOLD_SWAGGER_ENABLED` in a non-production environment also returns 404.

## Adding Go tests

Follow the patterns already in the repo:

- Prefer table-driven tests; `webui/fallback_test.go` is the reference shape
  (case struct + loop + `httptest`).
- For middleware, fake the seams instead of the network: `authz_test.go` uses a
  `fakeChecker` for RBAC and a `fakeTransport` whose `Operation()` returns the
  route, then asserts the Kratos error code and reason via `kerrors.FromError`.
- For data-layer tests, use SQLite in memory via `gormlite.Open(":memory:")` and
  `db.AutoMigrate(...)`; see `data_test.go`. Never touch the default
  `file:../data/scaffold.db`.
- Keep tests deterministic: no wall-clock sleeps, no random ports, no network.
  Inject `time.Time` / TTLs as the OAuth tests do with negative durations.
- Put tests in the same package (`package foo`, not `foo_test`) when you need
  unexported behavior, as all six existing files do.

## Reporting defects

Include these fields in every report (there is no template file — paste this):

| Field | Example |
| --- | --- |
| Environment | local dev / Docker / k8s; OS; Go and Node versions |
| Commit | `git rev-parse HEAD` output |
| Steps | numbered, minimal, reproducible from a clean state |
| Expected vs actual | one line each |
| Logs | backend stderr, browser console, relevant `curl -v` |
| Screenshots | only when the bug is visual |

Always state whether you reproduced it against the dev servers or the built
binary, and whether `pnpm run verify-generated` was clean — drift there causes
bugs that are not real.

## Related

- [Development](../development.md) — the daily loop, tests and lint section
- [DevOps and SRE](devops.md) — CI, Docker, and deployment gates
- [Backend](backend.md) — the Go layers your tests should target
- [Architecture lead · Known inconsistencies](architecture-lead.md#known-inconsistencies) — existing quirks you may mistake for bugs
