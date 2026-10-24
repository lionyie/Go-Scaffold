# Renaming the scaffold

[English](renaming.md) | [简体中文](renaming.zh-CN.md)

`scripts/rename.mjs` changes the project identity by token type. It never runs a
global string replace.

## Token sources

All current values come from `scaffold.json`:

```json
{
  "name": "scaffold",
  "module": "github.com/example/scaffold/services",
  "envPrefix": "SCAFFOLD",
  "npmScope": "@scaffold",
  "binary": "scaffold",
  "dbName": "scaffold",
  "dockerImage": "scaffold",
  "namespace": "scaffold-system",
  "displayName": "Scaffold",
  "description": "..."
}
```

| Token | Scope |
| --- | --- |
| Go module | `services/go.mod`, Go imports, `buf.gen.*.yaml`, `*.proto` `go_package`, template Go files |
| npm scope | `package.json`, `apps/`, `packages/`, `scripts/`, extension template JSON |
| environment prefix | all text files (`SCAFFOLD_` and `env.NewSource("SCAFFOLD")`) |
| binary | root `package.json`, `Makefile`, `scripts/`, `services/cmd/`, `deploy/`, `configs/`, docs |
| database name | `configs/`, `scripts/`, `deploy/`, docs |
| Docker image | `deploy/`, `.github/`, `Makefile`, README |
| Kubernetes namespace | `deploy/`, `configs/`, docs, README |
| display name | README, docs, root `package.json` |

The file `scaffold.json` keeps its name even after renaming; references to it are
protected.

## Usage

```bash
# dry run (default) — prints file, token type and match count
node scripts/rename.mjs --name myapp

# apply
node scripts/rename.mjs --name myapp --module github.com/me/myapp/services --yes

# pnpm alias
pnpm run rename -- --name myapp --module github.com/me/myapp/services --yes
```

Optional flags: `--display`, `--force`, `--skip-install`, `--skip-generate`.

After applying, the script:

1. rewrites `scaffold.json`,
2. deletes and reinstalls `pnpm-lock.yaml` unless `--skip-install`,
3. runs `scripts/generate.mjs`,
4. runs `go -C services mod tidy`.

## Safety

- If the repository has commits, `git status --porcelain` must be empty.
- If `extensions/installed.lock.json` contains extensions, renaming is refused
  because file hashes would become stale. Remove extensions or pass `--force`.
- `pnpm-lock.yaml` and `go.sum` are never text-replaced.

## Verification

```bash
git clone . /tmp/rename-test
cd /tmp/rename-test
node scripts/rename.mjs --name demo --module github.com/me/demo/services --yes
(cd services && go build ./...)
pnpm typecheck
```

The original repository is untouched by this test.
