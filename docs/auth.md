# Authentication, RBAC and SSO

## Local sign-in

`POST /api/v1/auth/login` verifies the password with bcrypt and returns an access
token (15 minutes) and a rotating refresh token (7 days). Tokens are JWTs signed
with `SCAFFOLD_JWT_SECRET`; the refresh token is stored as a SHA-256 hash.

The seeded administrator is `admin / admin123` with
`must_change_password=true`. Until `POST /api/v1/auth/password` succeeds, every
operation except password change and public endpoints returns
`MUST_CHANGE_PASSWORD`. The authz middleware enforces this, not just the UI.

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
