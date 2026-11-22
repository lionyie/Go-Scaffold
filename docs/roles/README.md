# Role-Based Documentation

[English](README.md) | [简体中文](README.zh-CN.md)

> Who this is for: anyone who wants a role-focused route through the repository instead of reading every topic document in order.

This directory gives each role a short, task-oriented entry point. The topic
documents in [`docs/`](../README.md) remain the reference material; the role
pages tell you which parts matter for your job and in what order to read them.

## Start by role

| Role | Start here | Then read | You will mostly touch |
| --- | --- | --- | --- |
| New member | [Onboarding](onboarding.md) | [Architecture](../architecture.md), [Development](../development.md) | Everything, read-only at first |
| Backend developer | [Backend](backend.md) | [Architecture](../architecture.md), [Auth](../auth.md), [Development](../development.md) | `services/`, `contracts/`, `configs/` |
| Frontend / UI developer | [Frontend and UI](frontend-ui.md) | [Architecture · UI system](../architecture.md#ui-system), [Development](../development.md) | `apps/`, `packages/ui/`, `packages/api-client/` |
| DevOps / SRE | [DevOps and SRE](devops.md) | [Architecture · Single binary](../architecture.md#single-binary-two-spas), [Auth](../auth.md) | `deploy/`, `configs/`, `services/internal/webui/` |
| Plugin / extension author | [Plugin and extension](plugin-author.md) | [Extensions](../extensions.md), [Runtime plugins](../../runtime-plugins/README.md) | `extensions/`, `runtime-plugins/` |
| Testing / QA | [Testing and QA](testing-qa.md) | [Development · Tests and lint](../development.md#tests-and-lint), [Auth](../auth.md) | `services/**/*_test.go`, CI |
| Architecture / tech lead | [Architecture lead](architecture-lead.md) | [Architecture](../architecture.md), [Auth](../auth.md), [Extensions](../extensions.md) | Boundaries, codegen policy, debt |

## Shared reading (everyone)

1. [Architecture](../architecture.md) — layers, dependency direction, boundary rules, generated outputs.
2. [Development](../development.md) — first run, the daily loop, and the full add-a-module chain.
3. [Authentication, RBAC and SSO](../auth.md) — how access is enforced end to end.
4. [API reference](../development.md#api-reference-and-swagger-ui) — the Swagger UI and the generated OpenAPI document.
5. [Documentation index](../README.md) — the complete topic-based index.

## Ownership at a glance

| Area | Primary role | Reference |
| --- | --- | --- |
| Protobuf contracts, `services/internal/**` | Backend | [Backend guide](backend.md), [Development](../development.md) |
| `apps/web`, `apps/admin`, `packages/ui` | Frontend / UI | [Frontend guide](frontend-ui.md), [Architecture · UI system](../architecture.md#ui-system) |
| `deploy/`, `configs/`, CI pipelines | DevOps / SRE | [DevOps guide](devops.md) |
| `extensions/`, `runtime-plugins/` | Plugin / extension author | [Plugin guide](plugin-author.md), [Extensions](../extensions.md) |
| Test suites and CI gates | Testing / QA | [QA guide](testing-qa.md) |
| Boundaries, generated-code policy, tech debt | Architecture / tech lead | [Architecture lead guide](architecture-lead.md) |

## Rules that apply to everyone

- Never hand-edit generated files. The full list is in [Architecture · Boundaries](../architecture.md#boundaries).
- Change the proto first, then run `pnpm run generate`; do not patch generated TypeScript or Go by hand.
- A new permission code needs `contracts/permissions.yaml`, the authz mapping, and `pnpm run verify-authz`.
- Keep English and Chinese documentation in pairs (`x.md` plus `x.zh-CN.md`).
- There are no automated frontend tests today; see [Testing and QA](testing-qa.md) before claiming UI coverage.

## Related

- [Documentation index](../README.md)
- [Root README](../../README.md)
- [Architecture](../architecture.md)
- [Development](../development.md)
- [Known inconsistencies](architecture-lead.md#known-inconsistencies)
