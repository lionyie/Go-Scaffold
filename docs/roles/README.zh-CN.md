# 按角色阅读的文档

[English](README.md) | [简体中文](README.zh-CN.md)

> 适合谁：想按岗位快速切入仓库、而不是从头按主题读完所有文档的人。

本目录为每类角色提供一个简短、面向任务的入口。`docs/` 下的主题文档仍然是
参考资料；角色页负责告诉你哪些部分与你的工作相关、应该按什么顺序阅读。

## 按角色开始

| 角色 | 从这里开始 | 接着读 | 主要接触的目录 |
| --- | --- | --- | --- |
| 新成员 | [新人全栈入门](onboarding.zh-CN.md) | [架构](../architecture.zh-CN.md)、[开发指南](../development.zh-CN.md) | 初期全部只读 |
| 后端开发 | [后端开发指南](backend.zh-CN.md) | [架构](../architecture.zh-CN.md)、[认证与 SSO](../auth.zh-CN.md)、[开发指南](../development.zh-CN.md) | `services/`、`contracts/`、`configs/` |
| 前端 / UI 开发 | [前端与 UI 开发指南](frontend-ui.zh-CN.md) | [架构](../architecture.zh-CN.md)、[开发指南](../development.zh-CN.md) | `apps/`、`packages/ui/`、`packages/api-client/` |
| 运维 / SRE | [运维与 SRE 指南](devops.zh-CN.md) | [架构](../architecture.zh-CN.md)、[认证与 SSO](../auth.zh-CN.md) | `deploy/`、`configs/`、`services/internal/webui/` |
| 插件 / 扩展作者 | [插件与扩展开发指南](plugin-author.zh-CN.md) | [扩展机制](../extensions.zh-CN.md)、[运行时插件](../../runtime-plugins/README.zh-CN.md) | `extensions/`、`runtime-plugins/` |
| 测试 / QA | [测试与质量保障指南](testing-qa.zh-CN.md) | [开发指南 · 测试与 lint](../development.zh-CN.md#测试与-lint)、[认证与 SSO](../auth.zh-CN.md) | `services/**/*_test.go`、CI |
| 架构 / 技术负责人 | [架构与技术负责人指南](architecture-lead.zh-CN.md) | [架构](../architecture.zh-CN.md)、[认证与 SSO](../auth.zh-CN.md)、[扩展机制](../extensions.zh-CN.md) | 边界、生成策略、技术债 |

## 所有人都该读

1. [架构](../architecture.zh-CN.md) — 分层、依赖方向、边界规则、生成物。
2. [开发指南](../development.zh-CN.md) — 首次运行、日常流程、新增模块完整链路。
3. [认证与 SSO](../auth.zh-CN.md) — 访问控制如何端到端生效。
4. [API 参考](../development.zh-CN.md#api-参考与-swagger-ui) — Swagger UI 与生成的 OpenAPI 文档。
5. [文档索引](../README.zh-CN.md) — 完整的主题文档索引。

## 职责一览

| 领域 | 主要负责角色 | 参考文档 |
| --- | --- | --- |
| Protobuf 契约、`services/internal/**` | 后端 | [后端开发指南](backend.zh-CN.md)、[开发指南](../development.zh-CN.md) |
| `apps/web`、`apps/admin`、`packages/ui` | 前端 / UI | [前端与 UI 开发指南](frontend-ui.zh-CN.md)、[架构](../architecture.zh-CN.md) |
| `deploy/`、`configs/`、CI 流水线 | 运维 / SRE | [运维与 SRE 指南](devops.zh-CN.md) |
| `extensions/`、`runtime-plugins/` | 插件 / 扩展作者 | [插件与扩展开发指南](plugin-author.zh-CN.md)、[扩展机制](../extensions.zh-CN.md) |
| 测试套件与 CI 门禁 | 测试 / QA | [测试与质量保障指南](testing-qa.zh-CN.md) |
| 边界、生成物策略、技术债 | 架构 / 技术负责人 | [架构与技术负责人指南](architecture-lead.zh-CN.md) |

## 所有人都要遵守的规则

- 不要手改生成文件。完整清单见[架构 · 边界规则](../architecture.zh-CN.md#边界规则)。
- 先改 proto，再执行 `pnpm run generate`；不要手补生成的 TypeScript 或 Go。
- 新增权限码需要同时修改 `contracts/permissions.yaml`、authz 映射，并运行 `pnpm run verify-authz`。
- 英文与中文文档必须成对维护（`x.md` 与 `x.zh-CN.md`）。
- 目前没有前端自动化测试；在声称 UI 已覆盖之前，请先看[测试与质量保障指南](testing-qa.zh-CN.md)。

## 相关

- [文档索引](../README.zh-CN.md)
- [根 README](../../README.zh-CN.md)
- [架构](../architecture.zh-CN.md)
- [开发指南](../development.zh-CN.md)
- [已知不一致](architecture-lead.zh-CN.md#已知不一致)
