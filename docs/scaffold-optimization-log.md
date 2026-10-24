# Scaffold optimization log

[English](scaffold-optimization-log.md) | [简体中文](scaffold-optimization-log.zh-CN.md)

Format: `任务ID | 变更文件 | 验收命令 | 结果 | 偏差说明`

| Task | Changed | Acceptance command | Result | Deviations |
| --- | --- | --- | --- | --- |
| Phase 0 | `apps/{web,admin}`, `services/`, `contracts/`, `extensions/`, `runtime-plugins/`, `deploy/`, `configs/`, `scripts/` | `pnpm run generate`; `go -C services build ./...`; `pnpm run typecheck` | pass | Repo has no commits; see end note |
| T1 | `extensions/` (was modules), `runtime-plugins/` (was plugins), registries, docs, CI, ignores | `grep -rn "modules/templates\|internal/modules\|installed\.json\|plugins/example" ...` | no matches (exit 1) | `.turbo` cache removed because it still referenced the old plugin path |
| T3 | `package.json`, `Makefile`, `scripts/{init,tool-versions,generate,verify-*}.mjs` | `pnpm run init && pnpm run generate && pnpm run build` (no make) | pass | Added `scripts/build.mjs` to orchestrate web + admin + plugins + server |
| T2 | `scripts/module.mjs`, `extensions/installed.lock.json`, templates | `add`; duplicate `add`; `remove`; `update`; `remove --force` | pass: duplicate refused, update wrote `.new`, force remove cleaned it | Lock uses a simple pid lock, not `proper-lockfile` |
| T4 | `scripts/generated-dirs.mjs`, `scripts/generate.mjs`, `scripts/verify-generated.mjs` | hash generated outputs → `generate` → hash again | identical (`idempotent=yes`) | `services/api` clear preserves `services/api/identity/v1/redact.go`; `services/internal/conf` clears only `conf.pb.go` |
| T5 | `cmd/genpermissions`, `rbac` constants, `scripts/verify-authz.mjs`, admin `Permission` usage, ESLint | `pnpm run generate`; `pnpm run verify-authz`; temporary missing mapping test | pass: `authz ok: 19 operations, 19 requirements, 11 permissions` | Missing-mapping test was done during development; final tree passes |
| T6 | `internal/auth/oauth`, `OAuthStore`, OAuth handler, provider templates, tests | `go -C services test ./internal/auth/oauth` | pass (5 security cases) | In-memory store only; Redis/DB documented |
| T7 | `seed.go`, `main.go`, login rate limiter, password change flow | smoke: login → 403 → change → relogin → 200; 5 bad logins then 429; `APP_ENV=production` runs | pass: `401×5,429`; default JWT and default admin fail fast | `go test -race` not run locally (no C compiler) |
| T8 | `scripts/db-reset.mjs`, `services/internal/data/migrations/`, `data.go`, `data_test.go` | `pnpm run db:reset -- --yes`; `go -C services test ./internal/data/database` | pass | Empty directory + AutoMigrate; golang-migrate remains a documented upgrade path |
| T9 | `scripts/rename.mjs` | temp copy: `rename --name demo --yes`; `pnpm typecheck`; `go -C services build ./...` | pass (10/10 typecheck tasks) | `scaffold.json` keeps its filename; lockfiles are recreated, not text-replaced |
| T10 | `extensions/upstream.json`, `check-upstream` | `check-upstream --json`; `--strict` | pass: offline status, strict exit 1; pinned commit recorded from an earlier online check | Network to GitHub is intermittent in this environment |
| T11 | `module.json#targets`, `features.gen.ts` | add `example-notes`; grep both generated files | pass: `/notes` route and menu in web and admin | Example removed after the check |
| T12 | `apps/*/app.json`, Vite configs, `copy-web-dist`, `webui` | `pnpm run build`; smoke `/`, `/admin/`, `/admin/users` | pass: 200 + SPA HTML; `bin/scaffold.exe` produced | `build-server` does not need app manifests; embeds are static |
| T13 | `services/internal/webui/fallback_test.go` | `go -C services test ./internal/webui` | pass (fallback, assets, reserved API, redirect) | none |
| T14 | `runtime-plugins/README.md`, architecture docs | manual review | pass | Plugin runtime remains experimental |
| T15 | OAuth tests, RBAC matrix, permission consistency, DB migration test | `go -C services test ./...`; `pnpm run verify-authz` | pass | Local run without `-race`; CI runs `-race` |

## Final regression

```text
pnpm run generate            pass
pnpm run verify-authz        pass (19 operations / 19 requirements / 11 permissions)
go -C services vet ./...     pass
go -C services test ./...    pass
pnpm -r --if-present typecheck pass
pnpm run lint:web            pass
buf lint                     pass
pnpm run build               pass (web + admin + server)
API/SPA smoke                pass
```

## Follow-up: forced password change and Vuetify UI

| Item | Change | Acceptance | Result |
| --- | --- | --- | --- |
| Forced password change bug | `/auth/me` was blocked by the must-change guard, so the SPA could not load the user or render a dialog. Authz now allows authenticated-only operations (`/auth/me`, logout) and blocks permissioned operations. The admin app shows a persistent change-password dialog and rotates the access token after a successful change. | Headless Chrome: login as `admin/admin123` → dialog appears on `/admin/profile` → submit new password → `/admin/dashboard` with 0 browser console errors. | pass |
| Element Plus → Vuetify 4 | Both apps and shared UI moved to Vuetify 4 + MDI; shared light/dark theme, glass surfaces, route/card transitions, snackbar/confirm components, favicon, and localized hero copy. Element Plus removed from all dependencies and templates. | `pnpm -r typecheck`, `pnpm run lint:web`, `pnpm run build:web`, `pnpm run build:admin`, wide screenshots of login/home/dashboard. | pass |
| Runtime plugin example | Added `runtime-plugins/hello-panel` (Vuetify page + static API), changed manifest `entry`/`styles` to be `dist`-relative, exposed Vuetify globals to plugins, and documented the full workflow. | `pnpm --filter @scaffold/plugin-hello build`; `GET /api/v1/plugins`; asset 200; API 401/403/200; headless admin + web pages with 0 console errors; `SCAFFOLD_PLUGINS_DEV=true` returns `devEntry`. | pass |

## End notes

- The repository still has no commits. `verify-generated` therefore reports
  “no commits yet; skipping diff check”. CI drift verification becomes active
  after the first commit.
- `go test -race` requires a C compiler on Windows; this machine has none. The
  same tests run in CI on Linux with `-race`.
- The go-admin reference is tracked, not merged. go-admin is Gin-based; this
  scaffold is Kratos-based.
