# Scaffold Documentation Index

[English](README.md) | [简体中文](README.zh-CN.md)

This is the complete documentation for the Scaffold starter. If you are new to
the project, read the documents below in order, or pick your role first.

## Start by role

| Role | Start here | Then read |
| --- | --- | --- |
| New member | [Onboarding](roles/onboarding.md) | [Architecture](architecture.md), [Development](development.md) |
| Backend developer | [Backend](roles/backend.md) | [Architecture](architecture.md), [Auth](auth.md), [Development](development.md) |
| Frontend / UI developer | [Frontend and UI](roles/frontend-ui.md) | [Architecture · UI system](architecture.md#ui-system), [Development](development.md) |
| DevOps / SRE | [DevOps and SRE](roles/devops.md) | [Architecture · Single binary](architecture.md#single-binary-two-spas), [Auth](auth.md) |
| Plugin / extension author | [Plugin and extension](roles/plugin-author.md) | [Extensions](extensions.md), [Runtime plugins](../runtime-plugins/README.md) |
| Testing / QA | [Testing and QA](roles/testing-qa.md) | [Development · Tests and lint](development.md#tests-and-lint), [Auth](auth.md) |
| Architecture / tech lead | [Architecture lead](roles/architecture-lead.md) | [Architecture](architecture.md), [Auth](auth.md), [Extensions](extensions.md) |

The detailed role index is [docs/roles/README.md](roles/README.md).

## Directory

| Document | Contents | Audience |
| --- | --- | --- |
| [Architecture](architecture.md) | Layered structure, dependency direction, boundary rules, generated outputs, single-binary dual SPA, UI system | Everyone who wants the full picture — **read this first** |
| [Development](development.md) | First run, the daily three-terminal loop, the full add-a-module chain, code generation, database, tests and lint, production notes | Developers working in this repository daily |
| [API reference](development.md#api-reference-and-swagger-ui) | Swagger UI and the OpenAPI document generated from `contracts/proto`, with environment gating | Anyone browsing or testing the HTTP API |
| [Authentication, RBAC and SSO](auth.md) | Local sign-in, forced password change, the RBAC model, the full OAuth flow, provider environment variables and callback URLs | Anyone touching login/permissions, or wiring up GitHub/Google/WeChat/QQ/OIDC |
| [Extensions](extensions.md) | `module.mjs` add/list/remove/update, `module.json` fields, lockfile and safety, bundled templates | Anyone installing features, writing runnable examples, or packaging their own extension |
| [Runtime plugins](../runtime-plugins/README.md) | A different extension model: manifest + browser ESM loaded at startup, Vuetify host components, static/proxy backend, auth, troubleshooting | Anyone adding UI or backend endpoints without touching core code |
| [Renaming](renaming.md) | `rename.mjs` token-type renaming, scope, safety, verification | Anyone starting their own project from this scaffold |
| [Optimization log](scaffold-optimization-log.md) | Changes, acceptance commands and results for tasks T1–T15, plus the forced password-change fix and Vuetify migration | Anyone reviewing the structural history and acceptance criteria of this scaffold |
| [Role guides](roles/README.md) | Seven role-based routes into the repository, with ownership boundaries and shared reading | Anyone who wants a role-focused route through the docs |

## Quick links

- **Get running in 5 minutes** → [Development · First run](development.md#first-run)
- **Browse or test the HTTP API** → [Development · API reference](development.md#api-reference-and-swagger-ui) (non-production only)
- **Choose a role-based reading path** → [Role guides](roles/README.md)
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
