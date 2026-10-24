# Extensions

Extensions are compile-time modules copied into the tree by
`scripts/module.mjs`. They are distinct from the experimental runtime plugins in
`runtime-plugins/`.

## Commands

```bash
node scripts/module.mjs list
node scripts/module.mjs add oauth-github
node scripts/module.mjs remove oauth-github
node scripts/module.mjs update oauth-github
node scripts/module.mjs check-upstream

# pnpm aliases
pnpm run module:add -- oauth-github
pnpm run module:remove -- oauth-github
pnpm run module:update -- oauth-github
pnpm run module:list
pnpm run module:check-upstream
```

`Makefile` aliases:

```bash
make module-add NAME=oauth-github
make example-add NAME=notes      # module:add example-notes
make example-list
make example-remove NAME=notes
```

## Manifest

`extensions/templates/<id>/module.json`:

```json
{
  "schemaVersion": 1,
  "id": "oauth-github",
  "kind": "feature",
  "version": "1.0.0",
  "requires": ["auth-core"],
  "copy": [
    { "from": "provider.go", "to": "services/internal/auth/oauth/providers/github.go" }
  ],
  "registries": ["oauth-providers"],
  "env": ["SCAFFOLD_OAUTH_GITHUB_CLIENT_ID"],
  "codegen": false,
  "authzImport": "...",
  "authzAlias": "notesv1",
  "authz": [{ "operation": "OperationNotesServiceListNotes", "public": true }],
  "generatedDirs": ["services/api/notes/v1"],
  "targets": {
    "admin": { "routes": [], "menu": [] },
    "web": { "routes": [], "menu": [] }
  }
}
```

- `copy` paths are relative to the template and the repository root.
- Go files copied under `services/internal/auth/oauth/providers` are
  blank-imported by `oauthproviders_gen`.
- Go files copied under `services/internal/extensions` are blank-imported by
  `registry_gen` and may register HTTP routes through
  `internal/extensions/registry`.
- `authz` entries are written to
  `services/internal/server/authz/extensions_gen.go`, so
  `verify-authz` sees them.
- `targets` are written to `apps/{web,admin}/src/router/features.gen.ts`.
- `codegen: true` runs the full generate pipeline after install/remove.
- `generatedDirs` are removed on uninstall.

## Lockfile and safety

`extensions/installed.lock.json` records:

- template version,
- `baseHash` of the template at install time,
- per-file SHA-256 hashes,
- generated files/directories,
- targets.

`add` validates dependencies and conflicts, stages files in `.scaffold-tmp/`,
then copies them and writes the lockfile. Failures roll back created files.
`remove` refuses to delete modified files unless `--force` is passed, and refuses
to remove an extension required by another one. Concurrent commands are rejected
through `extensions/.lock`. Every command supports `--dry-run` and `--json`.

`update` overwrites files that are unchanged since install. Files you modified
are written next to the target as `<file>.new`, with a diff summary.

## Bundled templates

- `oauth-github`, `oauth-google`, `oauth-wechat`, `oauth-qq`, `oauth-oidc`
- `example-notes`: a full slice with a proto, in-memory Go service, and pages in
  both apps. It demonstrates the extension registries and is safe to remove.

## Upstream tracking

`extensions/upstream.json` pins the go-admin reference commit. `check-upstream`
uses `git ls-remote`; when the network is unavailable it reports `offline` and
still exits 0. `--strict` exits non-zero for offline or outdated references.
`extensions/upstream-report.md` is the human-readable report. Upstream code is
not merged automatically because go-admin is Gin-based and this scaffold is
Kratos-based; the report is a migration checklist.
