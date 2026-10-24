# oauth-wechat

Adds WeChat Open Platform (微信开放平台) QR sign-in. WeChat returns no email
address, so the account is always linked by provider identity, never by email.

```bash
pnpm run module:add oauth-wechat
```

Configure:

```text
SCAFFOLD_OAUTH_WECHAT_APP_ID=...
SCAFFOLD_OAUTH_WECHAT_APP_SECRET=...
```

Register this callback domain in the WeChat Open Platform console:

```text
http(s)://<your-host>/api/v1/auth/oauth/wechat/callback
```
