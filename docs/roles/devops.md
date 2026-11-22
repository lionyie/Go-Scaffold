# DevOps and SRE Guide

[English](devops.md) | [简体中文](devops.zh-CN.md)

> Who this is for: operators building, deploying, and running the scaffold in containers, Compose, or Kubernetes.

## Responsibilities and Delivery Model

The repository ships as **one Go binary** (`bin/scaffold`) that embeds both Vue single-page apps and serves them directly:

- `/` serves the public site from `apps/web/dist`.
- `/admin/` serves the admin console from `apps/admin/dist`.

There is **no nginx and no separate web server**. The `webui` handler in `services/internal/webui/handler.go` reserves the prefixes `api/`, `proxy/`, `healthz`, and `readyz`; everything else falls back to the SPA. Static assets are cached for one year and `index.html` is served with no-cache, so rolling out a new binary is the entire frontend deployment.

## Build Artifacts and Ports

| Port / Artifact | Purpose |
| --- | --- |
| 5173 | `apps/web` Vite dev server (base `/`) |
| 5175 | `apps/admin` Vite dev server (base `/admin/`) |
| 5174 | Runtime plugin HMR dev server (`pnpm --filter @scaffold/plugin-hello dev`) |
| 8000 | Backend HTTP (also serves both SPAs in production) |
| 9000 | Backend gRPC |
| 80 | Kubernetes Service port for HTTP (maps to 8000) |
| 3306 | MySQL exposed by the Compose stack |

Outputs:

- `apps/web/dist/` and `apps/admin/dist/` from `pnpm run build:web` / `pnpm run build:admin`.
- `bin/scaffold.exe` on Windows or `bin/scaffold` elsewhere, from `pnpm run build:server` (`scripts/build-server.mjs`).
- `pnpm run build` runs the full pipeline via `scripts/build.mjs`.

## Configuration System

Three config files live in `configs/`:

- `configs/config.local.yaml` — local development; defaults to SQLite at `file:../data/scaffold.db` with WAL enabled.
- `configs/docker.yaml` — used inside the container; defaults to MySQL, JSON logs.
- `configs/k8s.yaml` — used by the Kubernetes manifests; same shape as the Docker config.

Rules that matter operationally:

- Environment variables use the **`SCAFFOLD_` prefix** and override the YAML file (Kratos env source).
- YAML values support **`${NAME:default}` placeholders**, e.g. `"${JWT_SECRET:dev-only-change-me}"` and `"${DB_SOURCE:...}"`. So `SCAFFOLD_DB_SOURCE` overrides `data.database.source`, `SCAFFOLD_LOG_FORMAT` overrides `log.format`, and so on.
- The binary is started with `-conf <path>`; the container entrypoint defaults to `/app/configs/docker.yaml`, the k8s Deployment overrides it to `/app/configs/k8s.yaml`.

## Production Secrets and Startup Gates

Set these before enabling production mode:

- `SCAFFOLD_JWT_SECRET` — JWT signing secret.
- `SCAFFOLD_OAUTH_STATE_SECRET` — OAuth state cookie signing secret.
- `SCAFFOLD_SEED_ADMIN_PASSWORD` — first-start password for the seeded `admin` account (default `admin123`).
- `APP_ENV=production` — turns on the gates.

With `APP_ENV=production`, `services/cmd/api/main.go` **panics at startup** if the JWT secret or OAuth state secret is empty or still the dev default, and `services/internal/data/database/seed.go` refuses to start with the unrotated seeded admin password. A crash-looping pod with one of these panic messages is a configuration problem, not an application bug.

## Local Full-Stack Compose

`deploy/compose/docker-compose.yml` brings up **MySQL 8.4 plus the server**:

```bash
docker compose -f deploy/compose/docker-compose.yml up --build
```

- MySQL is exposed on `3306` with database/user/password `scaffold` and persists to the `mysql-data` volume.
- The server waits for MySQL via `depends_on` with `condition: service_healthy` (mysqladmin ping healthcheck).
- The server sets `SCAFFOLD_DB_SOURCE` to MySQL and **placeholder secrets** (`change-me-in-production`) — do not reuse this stack as-is beyond local evaluation.
- Tear down with `docker compose -f deploy/compose/docker-compose.yml down` (add `-v` to drop the database volume).

## Image Build

`deploy/docker/Dockerfile` is a three-stage build:

1. **node:24-alpine** — installs the pnpm workspace, builds `apps/web`, `apps/admin`, and the runtime plugins, then copies web dist into the Go embed path.
2. **golang:1.27-alpine** — `CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.Version=${VERSION}"`; pass `--build-arg VERSION=<tag>` to stamp the version.
3. **gcr.io/distroless/static-debian12:nonroot** — final image contains only `/app/scaffold`, `/app/configs`, and `/app/runtime-plugins`; runs as the `nonroot` user and exposes 8000/9000.

Build locally with:

```bash
docker build -f deploy/docker/Dockerfile -t scaffold:local .
```

## Kubernetes

Manifests live in `deploy/k8s/`:

- `namespace.yaml` — `scaffold-system`.
- `deployment.yaml` — single replica, image `scaffold:latest` with `imagePullPolicy: IfNotPresent`, args `-conf /app/configs/k8s.yaml`.
- `service.yaml` — port 80 -> 8000 (http) and 9000 -> 9000 (grpc).
- `secret.example.yaml` — template for `scaffold-secrets` (`jwt-secret`, `oauth-state-secret`); replace the placeholder values before applying.

Probes: readiness on `/readyz` (initial delay 5s, period 10s), liveness on `/healthz` (initial delay 10s, period 20s).

Current gaps to plan around:

- [ ] No MySQL manifest — `SCAFFOLD_DB_SOURCE` points at a `mysql:3306` host you must provide yourself.
- [ ] No Ingress or TLS — front the Service with your own ingress/load balancer.
- [ ] No HPA, and `replicas: 1` is effectively mandatory (see scaling limits below).
- [ ] `scaffold:latest` + `IfNotPresent` means node-cached images can go stale; tag images per release in real deployments.

## Database

- **Local:** SQLite (`file:../data/scaffold.db`, WAL mode). Back up by copying the file while the server is stopped.
- **Production:** MySQL via `SCAFFOLD_DB_SOURCE` (DSN in the Compose and k8s examples).
- **Schema:** GORM `AutoMigrate` runs at startup when `data.database.auto_migrate` is true (default in all shipped configs). The `migrations/` directory is empty — there are no hand-written migrations yet, so destructive schema changes need care: back up (`mysqldump` or a volume snapshot) before upgrades and prefer additive model changes.
- Seeding (`data.database.seed`) creates the initial admin user, roles, and permissions on an empty database.

## Health Checks, Logs, and Observability

- `GET /healthz` — liveness; `GET /readyz` — readiness. Both are served on port 8000 outside the SPA fallback.
- Logs go to stdout; `log.format` is `text` locally and `json` in `configs/docker.yaml` and `configs/k8s.yaml` (`SCAFFOLD_LOG_FORMAT` / `SCAFFOLD_LOG_LEVEL` to override).
- There is **no metrics or tracing wired** — no Prometheus endpoint and no OpenTelemetry exporter. Treat log aggregation as your only signal until that changes.

## API Reference (Swagger UI)

Non-production deployments can expose the generated OpenAPI document at `/swagger/`.
It is controlled by `swagger.enabled` (`SCAFFOLD_SWAGGER_ENABLED`):

- `configs/config.local.yaml` defaults it on.
- `configs/docker.yaml` and `configs/k8s.yaml` default it off.
- `APP_ENV=production` force-disables it even when the flag is true.
- When off, `/swagger/` returns 404 and is not part of the SPA fallback.

For staging, add `SCAFFOLD_SWAGGER_ENABLED: "true"` to the container environment
and do not set `APP_ENV=production`. Never expose `/swagger/` through a public
Ingress or load balancer in production.

## Scaling Limits

Horizontal scaling is currently constrained:

- The OAuth state/ticket store is **in-memory**, so OAuth logins break if requests hop between replicas. Keep `replicas: 1` or use sticky sessions until a shared store exists.
- Refresh tokens live in the database, but the **login rate limiter is in-memory** (per replica), and plugin assets are read from the container filesystem (`/app/runtime-plugins`), so **every replica must carry the same plugin set baked into the image** — plugins cannot be added to one replica at runtime.
- Plan for a shared state store (Redis or DB-backed) before raising the replica count.

## CI and Release

`.github/workflows/ci.yml` runs five jobs:

| Job | What it does |
| --- | --- |
| codegen | `pnpm run generate`, `pnpm run verify-authz`, then `git diff --exit-code` to catch codegen drift |
| go | `go vet ./...` and `go test -race ./...` in `services/` |
| web | `pnpm run lint:web`, `typecheck`, `build:web`, `build:admin` |
| extensions | adds/removes the `oauth-github` and `example-notes` modules and verifies the tree is clean |
| docker | builds `deploy/docker/Dockerfile` tagged `scaffold:ci`, **push: false**, main branch only |

Operational gap: nothing publishes an image. Add a registry login and push step (or a release workflow) to the docker job when you adopt a registry, and tag by git SHA or version instead of `latest`.

## Troubleshooting

| Symptom | Likely cause | Check |
| --- | --- | --- |
| Container crash-loops at startup with a panic about secrets | `APP_ENV=production` with default JWT/OAuth secret | Pod logs; `SCAFFOLD_JWT_SECRET`, `SCAFFOLD_OAUTH_STATE_SECRET` in the Secret |
| Crash-loop about seeded admin password | Unrotated `admin123` in production | Set `SCAFFOLD_SEED_ADMIN_PASSWORD` before first start |
| 502 / connection refused from the browser | Backend not listening on 8000 | `docker compose ps`, `kubectl get pods`, port-forward 8000 |
| Readiness probe failing | Server up but `/readyz` not OK | `kubectl describe pod`; curl `/readyz` on port 8000 |
| OAuth login fails after redirect | State secret unset, or multiple replicas with in-memory state | `SCAFFOLD_OAUTH_STATE_SECRET`; replica count / sticky sessions |
| Admin UI loads but API calls fail | Requests not reaching `/api` on the backend | Vite proxy targets in dev; Service port 80 -> 8000 in k8s |
| Plugin page 404 or stale assets | Plugin missing from image or dev mode off | `/app/runtime-plugins/<id>` in the image; `SCAFFOLD_PLUGINS_DEV` |
| Dev servers unreachable | Expected ports not free | 5173 (web), 5175 (admin), 5174 (plugin), 8000/9000 (backend) |
| gRPC clients cannot connect | Only port 80 exposed | Service also maps 9000 -> 9000; check your LB/Ingress passes 9000 |

## Related

- [Architecture](../architecture.md)
- [Authentication and Authorization](../auth.md)
- [Development Guide](../development.md)
- [Testing and QA](testing-qa.md)
