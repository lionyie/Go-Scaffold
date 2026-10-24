# Scaffold 文档索引

[English](README.md) | [简体中文](README.zh-CN.md)

这里是 Scaffold 脚手架的完整文档。如果你是第一次接触这个项目，建议按下面的顺序阅读。

## 文档目录

| 文档 | 内容 | 适合谁 |
| --- | --- | --- |
| [架构（architecture）](architecture.zh-CN.md) | 分层结构、依赖方向、五条边界规则、生成物清单、单二进制双 SPA、UI 体系 | 想理解项目全貌的所有人，**建议先读** |
| [开发指南（development）](development.zh-CN.md) | 首次运行、日常三终端流程、新增后端模块的完整链路、代码生成、数据库、测试与 lint、生产注意事项 | 日常在这个仓库里写代码的开发者 |
| [认证与 SSO（auth）](auth.zh-CN.md) | 本地登录、强制改密、RBAC 模型、第三方登录全流程、各 provider 的环境变量与回调地址 | 做登录/权限相关功能，或要接入 GitHub/Google/微信/QQ/OIDC 的人 |
| [扩展机制（extensions）](extensions.zh-CN.md) | `module.mjs` 的 add/list/remove/update、`module.json` 清单字段、锁文件与安全机制、内置模板 | 想安装功能模块、写可运行示例，或自己封装扩展的人 |
| [项目改名（renaming）](renaming.zh-CN.md) | `rename.mjs` 按 token 类型改名、作用范围、安全措施、验证方式 | 用这个脚手架启动自己的新项目的人 |
| [优化执行日志（scaffold-optimization-log）](scaffold-optimization-log.zh-CN.md) | T1–T15 任务的变更、验收命令与结果，以及强制改密修复与 Vuetify 迁移的后记 | 想了解脚手架经历了哪些结构性改动、验收标准是什么的人 |

## 快速跳转

- **5 分钟跑起来** → [开发指南 · 首次运行](development.zh-CN.md#首次运行)
- **这个目录里什么能改、什么不能改** → [架构 · 边界规则](architecture.zh-CN.md#边界规则)
- **怎么加一个后端接口 + 页面** → [开发指南 · 新增后端模块](development.zh-CN.md#新增后端模块)
- **怎么接入 GitHub 登录** → [扩展机制](extensions.zh-CN.md) + [认证与 SSO · Provider 环境变量](auth.zh-CN.md#provider-环境变量)
- **怎么把它变成我自己的项目** → [项目改名](renaming.zh-CN.md)
- **示例（notes）是怎么装的、怎么删的** → [扩展机制 · 命令](extensions.zh-CN.md#命令)

## 其他位置的重要 README

- [根 README（中文）](../README.zh-CN.md) — 项目概览、目录结构、常用命令
- [运行时插件（中文）](../runtime-plugins/README.zh-CN.md) — 实验性插件目录的约定（EXPERIMENTAL）
- [数据库迁移（中文）](../services/internal/data/migrations/README.zh-CN.md) — 迁移目录的命名规范与升级路径

英文版入口见 [README](README.md)。
