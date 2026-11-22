# Authentication, RBAC and SSO

[English](auth.md) | [简体中文](auth.zh-CN.md)

## Local sign-in

`POST /api/v1/auth/login` verifies the password with bcrypt and returns an access
token (15 minutes) and a rotating refresh token (7 days). Tokens are JWTs signed
with `SCAFFOLD_JWT_SECRET`; the refresh token is stored as a SHA-256 hash.

The seeded administrator is `admin / admin123` with
`must_change_password=true`. Until `POST /api/v1/auth/password` succeeds,
authenticated-only operations such as `GET /api/v1/auth/me` and logout remain
available, while every permissioned operation returns `MUST_CHANGE_PASSWORD`.
This lets the admin app render a persistent change-password dialog. The authz
middleware enforces the restriction, not just the UI. After a successful change
the SPA rotates its access token before loading the dashboard.

Login failures are rate limited in memory: five failures per IP + username in a
minute return HTTP 429.

## RBAC model

```text
users(id, username, email, password_hash, is_admin, must_change_password, status, ...)
roles(id, name, display_name, description, is_system)
permissions(code, name, resource, action)
user_roles(user_id, role_id)
role_permissions(role_id, permission_code)
oauth_accounts(id, user_id, provider, provider_user_id, email, email_verified, ...)
refresh_tokens(id, user_id, token_hash, session_id, expires_at, revoked_at)
```

- The permission catalog lives in `contracts/permissions.yaml` and is generated
  into Go (`rbac.Catalog`, `rbac.All`) and TypeScript (`Permission`).
- `services/internal/server/authz/mapping.go` maps every generated operation to
  either `Public`, an empty code (authenticated only), or a permission code.
- Unknown operations fail closed with 403.
- `is_admin=true` bypasses permission checks.
- `pnpm run verify-authz` fails if a generated operation is missing from the
  mapping, a mapping entry is stale, or a permission code is unknown.

### Permission sources

| Source | Origin | Delete | Edit code |
| --- | --- | --- | --- |
| `builtin` | `contracts/permissions.yaml` | no | no |
| `plugin` | plugin manifest | no | no |
| `custom` | `POST /api/v1/permissions` (API) | yes | no |

Seeding upserts the builtin catalog and prunes only stale `builtin` rows, so
custom and plugin permissions survive restarts. Deleting a custom permission
also removes its `role_permissions` rows in the same transaction.

A custom permission is only a catalog entry: it protects no endpoint until an
operation is mapped to it, either in `Requirements` or through a runtime binding
(see below).

### Runtime operation bindings

The compiled mapping in `mapping.go` is the default. The backend
`PUT /api/v1/operations/permission` can override the permission of an operation
at runtime. There is no admin UI for this anymore; it is an operator API.

- `services/internal/pkg/operations/operations_gen.go` is generated from the
  OpenAPI spec, so the catalog always matches the real routes.
- Overrides live in `operation_permissions(operation, permission_code)`.
- Public and authenticated-only operations are locked and cannot be rebound.
- Any other operation can be bound to any existing permission code, including
  custom ones. This is how a custom permission actually starts guarding an
  endpoint.
- Bindings are cached for 30 seconds and invalidated on write.
- Deleting a custom permission also deletes the operation bindings that
  referenced it, and those operations fall back to the compiled mapping.
- Changing bindings requires `permission:assign`; listing requires
  `permission:list`.

### Session roles

Each session token carries the caller's roles (`roles` claim). Authorization
checks read these roles directly; there is no per-request database lookup.

- Local password login resolves roles from the local RBAC tables.
- OIDC login uses the provider's `realm_access.roles`. For Keycloak, a realm
  role whose name equals a permission code (for example `user:list`) grants that
  permission directly.
- A role named `admin` implies `is_admin` (full bypass).
- `/api/v1/auth/me` returns the roles carried by the session.

Local password login remains available as a fallback. The admin permission
management UI was removed: Keycloak manages roles for SSO users, and the local
role page manages roles for password users.

Default roles: `admin` (all permissions) and `viewer` (read-only). The admin
console uses the generated `Permission` constants for route and menu guards, and
ESLint rejects permission string literals under `apps/admin/src`.

## Third-party sign-in

OAuth is provider-agnostic. The core endpoints always exist; providers are
extensions that register themselves from `init()`.

| Endpoint | Purpose |
| --- | --- |
| `GET /api/v1/auth/oauth/providers` | Public list of enabled providers |
| `POST /api/v1/auth/oauth/{provider}/start` | Returns `{authUrl}` |
| `GET /api/v1/auth/oauth/{provider}/callback` | Provider redirect; exchanges the code and redirects to the admin app |
| `POST /api/v1/auth/oauth/exchange` | Exchanges a one-time ticket for tokens |
| `GET /api/v1/auth/oauth/accounts` | Lists the caller's linked accounts |
| `DELETE /api/v1/auth/oauth/{provider}` | Unlinks a provider |

Flow:

1. The admin login page loads enabled providers.
2. `start` creates a random state, signs it with
   `SCAFFOLD_OAUTH_STATE_SECRET`, creates a PKCE verifier/challenge and a nonce,
   stores them in `OAuthStore`, and returns the provider URL.
3. The callback verifies the state signature, consumes it (single use), exchanges
   the code, and normalizes the profile.
4. Existing links log in. Verified email is used for automatic linking only when
   `email_verified=true`. Otherwise the account is created only when
   `auth.oauth.auto_register=true`.
5. Tokens are placed in a one-time, 60-second ticket. The callback redirects to
   `SCAFFOLD_OAUTH_ADMIN_REDIRECT_URL` with `?ticket=...`; the SPA exchanges it.
   Tokens never appear in the URL.
6. The callback sets `Referrer-Policy: no-referrer`.

### Provider environment variables

```text
SCAFFOLD_OAUTH_GITHUB_CLIENT_ID / _CLIENT_SECRET
SCAFFOLD_OAUTH_GOOGLE_CLIENT_ID / _CLIENT_SECRET
SCAFFOLD_OAUTH_WECHAT_APP_ID / _APP_SECRET
SCAFFOLD_OAUTH_QQ_APP_ID / _APP_KEY
SCAFFOLD_OAUTH_OIDC_ISSUER / _CLIENT_ID / _CLIENT_SECRET / _SCOPES
```

Register the callback URL in each provider console:

```text
http(s)://<host>/api/v1/auth/oauth/<provider>/callback
```

WeChat (Open Platform QR) and QQ Connect do not return an email address; they
never participate in email-based linking. OIDC verifies the ID token signature
against the issuer JWKS and checks `iss`, `aud`, `exp`, and the nonce.

### Keycloak (OIDC)

`pnpm run module:add oauth-oidc` installs the generic OpenID Connect provider at
`services/internal/auth/oauth/providers/oidc.go`. A development realm ships in
`deploy/keycloak/scaffold-realm.json`.

Start Keycloak (opt-in override):

```bash
docker compose -f deploy/compose/docker-compose.yml \
               -f deploy/compose/docker-compose.keycloak.yml up -d keycloak
```

It listens on `http://localhost:8080`, imports realm `scaffold`, and creates:

- client `scaffold-admin` (confidential, secret `dev-scaffold-admin-secret`,
  PKCE S256)
- redirect URIs for `http://localhost:8000` and `http://127.0.0.1:8000`
- dev user `admin@example.com / admin123` with realm role `admin`
- dev user `viewer@example.com / viewer123` with the composite role `viewer`
- one realm role per permission code (for example `user:list`)

Realm roles are the permission codes. The ID token's `realm_access.roles` is
copied into the session, so assigning `user:list` in Keycloak grants the
`user:list` permission. The `admin` realm role implies `is_admin`.

Configure the API:

```text
SCAFFOLD_OAUTH_STATE_SECRET=<random>
SCAFFOLD_OAUTH_OIDC_ISSUER=http://localhost:8080/realms/scaffold
SCAFFOLD_OAUTH_OIDC_CLIENT_ID=scaffold-admin
SCAFFOLD_OAUTH_OIDC_CLIENT_SECRET=dev-scaffold-admin-secret
SCAFFOLD_OAUTH_OIDC_SCOPES=openid email profile
SCAFFOLD_OAUTH_OIDC_DISPLAY_NAME=Keycloak
SCAFFOLD_OAUTH_AUTO_REGISTER=true
```

Account linking:

- A verified email links the Keycloak identity to an existing local user with
  the same address.
- The seeded admin uses `admin@example.com`; override it with
  `SCAFFOLD_SEED_ADMIN_EMAIL`.
- With `auto_register=true`, an unknown identity creates a local user without
  administrator rights.
- Local password login remains enabled alongside SSO.

Security notes:

- PKCE S256 and the OIDC nonce are both verified.
- Always set `SCAFFOLD_OAUTH_STATE_SECRET`.
- In the full compose stack the server container uses
  `extra_hosts: localhost:host-gateway` so it can reach the host-published
  Keycloak port.

Keycloak is an opt-in provider: the local RBAC and password login remain the
default, and the scaffold does not require Keycloak to run.

### Multi-replica deployments

`OAuthStore` ships with an in-memory implementation for a single replica. State
and tickets are lost on restart and are not shared across pods. Replace it with a
Redis or database implementation before scaling horizontally.

## go-admin reference

The RBAC model and admin UX are inspired by
[go-admin](https://github.com/go-admin-team/go-admin), but this scaffold keeps a
simple role-permission table instead of Casbin and does not copy go-admin code.
`extensions/upstream.json` pins the reference commit, and
`pnpm run module:check-upstream` reports whether it moved.

## Related

- Guides: [Architecture](architecture.md) · [Extensions](extensions.md) · [Development](development.md)
- Roles: [Backend developer](roles/backend.md) · [DevOps and SRE](roles/devops.md) · [Testing and QA](roles/testing-qa.md)
- All roles: [Role guides](roles/README.md)
