# 扩展机制

[English](extensions.md) | [简体中文](extensions.zh-CN.md)

扩展是由 `scripts/module.mjs` 编译期复制进代码树的模块。它们与
`runtime-plugins/` 中的实验性运行时插件是两回事。

## 命令

```bash
node scripts/module.mjs list
node scripts/module.mjs add oauth-github
node scripts/module.mjs remove oauth-github
node scripts/module.mjs update oauth-github
node scripts/module.mjs check-upstream

# pnpm 别名
pnpm run module:add -- oauth-github
pnpm run module:remove -- oauth-github
pnpm run module:update -- oauth-github
pnpm run module:list
pnpm run module:check-upstream
```

`Makefile` 别名：

```bash
make module-add NAME=oauth-github
make example-add NAME=notes      # 即 module:add example-notes
make example-list
make example-remove NAME=notes
```

## 清单

`extensions/templates/<id>/module.json`：

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

- `copy` 路径分别相对模板目录与仓库根目录。
- 复制到 `services/internal/auth/oauth/providers` 下的 Go 文件会被
  `oauthproviders_gen` blank-import。
- 复制到 `services/internal/extensions` 下的 Go 文件会被 `registry_gen`
  blank-import，并可通过 `internal/extensions/registry` 注册 HTTP 路由。
- `authz` 条目会写入
  `services/internal/server/authz/extensions_gen.go`，因此
  `verify-authz` 能看到它们。
- `targets` 会写入 `apps/{web,admin}/src/router/features.gen.ts`。
- `codegen: true` 表示安装/删除后执行完整的生成流水线。
- `generatedDirs` 在卸载时删除。

## 锁文件与安全

`extensions/installed.lock.json` 记录：

- 模板版本，
- 安装时模板的 `baseHash`，
- 每个文件的 SHA-256 哈希，
- 生成的文件/目录，
- targets。

`add` 会校验依赖与冲突，先把文件暂存到 `.scaffold-tmp/`，再复制并写锁文件。
失败时回滚已创建的文件。`remove` 拒绝删除被修改过的文件（除非传 `--force`），
也拒绝删除仍被其他扩展依赖的扩展。并发命令通过 `extensions/.lock` 拒绝。
所有命令都支持 `--dry-run` 与 `--json`。

`update` 直接覆盖自安装以来未改动的文件；你修改过的文件会写成旁边的
`<file>.new`，并输出 diff 摘要。

## 内置模板

- `oauth-github`、`oauth-google`、`oauth-wechat`、`oauth-qq`、`oauth-oidc`
- `example-notes`：完整切片，包含 proto、内存版 Go 服务，以及两个应用中的
  页面。它演示了扩展注册表机制，随时可以安全删除。

## 上游跟踪

`extensions/upstream.json` 固定 go-admin 的参考 commit。`check-upstream` 使用
`git ls-remote`；网络不可用时报告 `offline` 并仍以 0 退出。`--strict` 会在
离线或过期时以非零退出。`extensions/upstream-report.md` 是人类可读的报告。
上游代码不会被自动合并——go-admin 基于 Gin，本脚手架基于 Kratos；报告只作为
迁移清单。
