# oauth-oidc

Adds a generic OpenID Connect provider for enterprise SSO (Keycloak, Authentik,
Auth0, Okta, etc.). The ID token is verified against the issuer's JWKS and its
`iss`, `aud`, and `exp` claims.

```bash
pnpm run module:add oauth-oidc
```

Configure:

```text
SCAFFOLD_OAUTH_OIDC_ISSUER=https://sso.example.com/realms/main
SCAFFOLD_OAUTH_OIDC_CLIENT_ID=scaffold
SCAFFOLD_OAUTH_OIDC_CLIENT_SECRET=...
SCAFFOLD_OAUTH_OIDC_SCOPES="openid email profile"   # optional
```

Register the callback URL:

```text
http(s)://<your-host>/api/v1/auth/oauth/oidc/callback
```
