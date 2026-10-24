# Scaffold Documentation Index

[English](README.md) | [简体中文](README.zh-CN.md)

This is the complete documentation for the Scaffold starter. If you are new to
the project, read the documents below in order.

## Directory

| Document | Contents | Audience |
| --- | --- | --- |
| [Architecture](architecture.md) | Layered structure, dependency direction, boundary rules, generated outputs, single-binary dual SPA, UI system | Everyone who wants the full picture — **read this first** |
| [Development](development.md) | First run, the daily three-terminal loop, the full add-a-module chain, code generation, database, tests and lint, production notes | Developers working in this repository daily |
| [Authentication, RBAC and SSO](auth.md) | Local sign-in, forced password change, the RBAC model, the full OAuth flow, provider environment variables and callback URLs | Anyone touching login/permissions, or wiring up GitHub/Google/WeChat/QQ/OIDC |
| [Extensions](extensions.md) | `module.mjs` add/list/remove/update, `module.json` fields, lockfile and safety, bundled templates | Anyone installing features, writing runnable examples, or packaging their own extension |
| [Runtime plugins](../runtime-plugins/README.md) | A different extension model: manifest + browser ESM loaded at startup, Vuetify host components, static/proxy backend, auth, troubleshooting | Anyone adding UI or backend endpoints without touching core code |
| [Renaming](renaming.md) | `rename.mjs` token-type renaming, scope, safety, verification | Anyone starting their own project from this scaffold |
| [Optimization log](scaffold-optimization-log.md) | Changes, acceptance commands and results for tasks T1–T15, plus the forced password-change fix and Vuetify migration | Anyone reviewing the structural history and acceptance criteria of this scaffold |

## Quick links

- **Get running in 5 minutes** → [Development · First run](development.md#first-run)
- **What can and cannot be edited** → [Architecture · Boundaries](architecture.md#boundaries)
- **Add a backend API + page** → [Development · Add a backend module](development.md#add-a-backend-module)
- **Wire up GitHub sign-in** → [Extensions](extensions.md) + [Auth · Provider environment variables](auth.md#provider-environment-variables)
- **Make it your own project** → [Renaming](renaming.md)
- **How the notes example is installed and removed** → [Extensions · Commands](extensions.md#commands)
- **Write a runtime plugin** → [Runtime plugins](../runtime-plugins/README.md)

## Important READMEs elsewhere

- [Root README (English)](../README.md) — project overview, structure, common commands
- [Runtime plugins](../runtime-plugins/README.md) — conventions for the experimental plugin directory
- [Database migrations](../services/internal/data/migrations/README.md) — naming rules and the upgrade path

The Chinese entry point is [README.zh-CN](README.zh-CN.md).
