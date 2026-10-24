# oauth-google

Adds Google OAuth/OIDC sign-in.

```bash
pnpm run module:add oauth-google
```

Configure:

```text
SCAFFOLD_OAUTH_GOOGLE_CLIENT_ID=...
SCAFFOLD_OAUTH_GOOGLE_CLIENT_SECRET=...
```

Register this authorized redirect URI in Google Cloud Console:

```text
http(s)://<your-host>/api/v1/auth/oauth/google/callback
```
