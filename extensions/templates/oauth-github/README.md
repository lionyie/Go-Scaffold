# oauth-github

Adds GitHub OAuth sign-in.

```bash
pnpm run module:add oauth-github
```

Configure:

```text
SCAFFOLD_OAUTH_GITHUB_CLIENT_ID=...
SCAFFOLD_OAUTH_GITHUB_CLIENT_SECRET=...
```

Register the callback URL in the GitHub OAuth app:

```text
http(s)://<your-host>/api/v1/auth/oauth/github/callback
```
