# 项目改名

[English](renaming.md) | [简体中文](renaming.zh-CN.md)

`scripts/rename.mjs` 按 token 类型修改项目身份，从不做全局字符串替换。

## Token 来源

所有当前值来自 `scaffold.json`：

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

| Token | 作用范围 |
| --- | --- |
| Go module | `services/go.mod`、Go import、`buf.gen.*.yaml`、`*.proto` 的 `go_package`、模板 Go 文件 |
| npm scope | `package.json`、`apps/`、`packages/`、`scripts/`、扩展模板 JSON |
| 环境变量前缀 | 所有文本文件（`SCAFFOLD_` 与 `env.NewSource("SCAFFOLD")`） |
| 二进制名 | 根 `package.json`、`Makefile`、`scripts/`、`services/cmd/`、`deploy/`、`configs/`、文档 |
| 数据库名 | `configs/`、`scripts/`、`deploy/`、文档 |
| Docker 镜像名 | `deploy/`、`.github/`、`Makefile`、README |
| Kubernetes namespace | `deploy/`、`configs/`、文档、README |
| 显示名 | README、docs、根 `package.json` |

`scaffold.json` 这个文件名在改名后保持不变；对它的引用受保护。

## 用法

```bash
# dry run（默认）—— 打印文件、token 类型和命中次数
node scripts/rename.mjs --name myapp

# 执行
node scripts/rename.mjs --name myapp --module github.com/me/myapp/services --yes

# pnpm 别名
pnpm run rename -- --name myapp --module github.com/me/myapp/services --yes
```

可选参数：`--display`、`--force`、`--skip-install`、`--skip-generate`。

执行后脚本会：

1. 重写 `scaffold.json`，
2. 除非 `--skip-install`，否则删除并重装 `pnpm-lock.yaml`，
3. 运行 `scripts/generate.mjs`，
4. 运行 `go -C services mod tidy`。

## 安全措施

- 若仓库已有 commit，`git status --porcelain` 必须为空。
- 若 `extensions/installed.lock.json` 中有已安装扩展，改名会被拒绝（文件哈希
  会失效）。请先删除扩展，或传 `--force`。
- `pnpm-lock.yaml` 与 `go.sum` 永远不做文本替换。

## 验证

```bash
git clone . /tmp/rename-test
cd /tmp/rename-test
node scripts/rename.mjs --name demo --module github.com/me/demo/services --yes
(cd services && go build ./...)
pnpm typecheck
```

此测试不会改动原仓库。
