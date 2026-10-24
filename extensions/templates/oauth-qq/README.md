# oauth-qq

Adds QQ Connect sign-in. QQ returns no email address, so accounts are linked by
provider identity only.

```bash
pnpm run module:add oauth-qq
```

Configure:

```text
SCAFFOLD_OAUTH_QQ_APP_ID=...
SCAFFOLD_OAUTH_QQ_APP_KEY=...
```

Register the callback URL in the QQ Connect console:

```text
http(s)://<your-host>/api/v1/auth/oauth/qq/callback
```
