# Frontend and UI Developer Guide

[English](frontend-ui.md) | [简体中文](frontend-ui.zh-CN.md)

> Who this is for: developers building or changing the two Vue 3 SPAs, the shared Vuetify design system, and the frontend API layer.

## Scope and the two SPAs

The repository ships two independent Vue 3.5 + Vite 8 + Vuetify 4 single-page apps, both served by the Go binary in production:

| App | Role | Dev port | Base | Entry |
| --- | --- | --- | --- | --- |
| `apps/web` | Public-facing site | 5173 | `/` | `apps/web/src/main.ts` |
| `apps/admin` | Admin console | 5175 | `/admin/` | `apps/admin/src/main.ts` |

Change `apps/web` for marketing/public pages and anonymous plugin views; change `apps/admin` for anything behind login (users, roles, permissions, plugins, profile). Only `apps/admin` has an auth store (`apps/admin/src/stores/auth.ts`); `apps/web` configures the API client with `getToken: () => null`.

## Quick start and proxying

```bash
pnpm install
pnpm run dev:server   # Go backend on http://127.0.0.1:8000
pnpm run dev:web      # apps/web on :5173
pnpm run dev:admin    # apps/admin on :5175
```

Each app has an `app.json` (for example `apps/admin/app.json`) declaring `name`, `base`, `devPort`, `embed`, `proxy`, and `fallback`. The Vite dev server proxies the prefixes `/api`, `/proxy`, `/healthz`, and `/readyz` to `http://127.0.0.1:8000`, so the frontend always calls the backend same-origin and never deals with CORS.

## Directory layout and shared packages

Each app follows the same layout under `apps/*/src/`:

| Directory | Contents |
| --- | --- |
| `api/` | `client.ts` (openapi-fetch wiring) and `errors.ts` |
| `components/` | App-local components (for example `ChangePasswordDialog.vue` in admin) |
| `layouts/` | `DefaultLayout.vue` shell |
| `router/` | `index.ts`, `menu.ts`, generated `features.gen.ts` |
| `stores/` | Pinia setup stores (`auth.ts` admin-only, `plugin.ts` both apps) |
| `styles/` | `index.css` global utility classes |
| `views/` | Route-level views |
| `i18n/` | `index.ts` plus `locales/en-US.ts` and `locales/zh-CN.ts` |

Shared packages you will import daily:

| Package | Provides |
| --- | --- |
| `@scaffold/ui` (`packages/ui`) | `createScaffoldVuetify()`, `PageHeader`, `AppSnackbar`, `ConfirmDialog`, `notify`/`notifyApi`, `confirmDialog` |
| `@scaffold/shared` (`packages/shared`) | Generated `Permission` constants, `APP_NAME`, small helpers like `formatBytes` |
| `@scaffold/api-client` (`packages/api-client`) | `createApiClient()` over openapi-fetch, generated `schema.d.ts` types, `ApiError`/`toApiError` |
| `@scaffold/plugin-runtime` (`packages/plugin-runtime`) | `installPluginHost()`, `loadPlugins()`, plugin manifest types |

## Vuetify design system

Never call `createVuetify()` directly in an app. Both `main.ts` files call `createScaffoldVuetify()` from `@scaffold/ui` (`packages/ui/src/theme.ts`), which:

- registers all Vuetify components and directives plus the MDI icon set (`@mdi/font`),
- defines the two themes `scaffoldLight` and `scaffoldDark` with the shared color tokens,
- sets component defaults (pill buttons, `xl` cards, outlined comfortable inputs, and so on).

The active theme is persisted in `localStorage` under the key `scaffold.theme`; `main.ts` reads it at startup and passes `scaffoldDark` or `scaffoldLight` to `createScaffoldVuetify()`. Reference colors in CSS via Vuetify variables such as `rgb(var(--v-theme-primary))`.

Global utility classes live in `apps/*/src/styles/index.css` and are theme-aware: `.glass`, `.gradient-text`, `.hover-lift`, `.metric-card`, `.page-card`, and `.blob`. Prefer these over one-off styles so both themes keep working.

## Shared components

`packages/ui` exports three components and two imperative helpers:

- `PageHeader` — page title, optional description, and an `actions` slot. It is **not** globally registered; import it per view: `import { PageHeader } from '@scaffold/ui'`.
- `AppSnackbar` and `ConfirmDialog` — globally registered in both `main.ts` files via `app.component(...)`, so every template can use `<AppSnackbar />` and `<ConfirmDialog />` without imports (App.vue mounts them once).
- `notify(text, color)` / `notifyApi.success|error|warning|info(text)` — drive the snackbar from any script.
- `confirmDialog(title, message, options?)` — returns `Promise<boolean>`; use it for destructive actions instead of `window.confirm`.

## Routing and menu

Routes live in `apps/admin/src/router/index.ts` (and the slimmer `apps/web/src/router/index.ts`). The admin menu is `apps/admin/src/router/menu.ts`: `coreMenu` plus the generated `featureMenu`.

`RouteMeta` supports these fields:

| Field | Effect |
| --- | --- |
| `public` | Skip authentication entirely (login, OAuth callback) |
| `permission` | Permission code required; failing users go to the 403 view |
| `titleKey` | i18n key used for the document title |
| `pluginId` | Marks routes contributed by a runtime plugin |

The admin guard runs in this order: allow `public` routes, `auth.restore()`, redirect unauthenticated users to login, force `mustChangePassword` users to the profile page, load runtime plugins once, then check `meta.permission` and fall back to the forbidden view.

`features.gen.ts` (in both apps) is generated by `pnpm run module:add/remove` — never hand-edit it. Add your own routes to `index.ts` and menu entries to `menu.ts`.

## Permissions in the UI

Always reference the generated constants from `@scaffold/shared`:

```ts
import { Permission } from '@scaffold/shared'

meta: { titleKey: 'nav.users', permission: Permission.UserList }
```

The constants come from `packages/shared/src/permissions.gen.ts` (generated from `contracts/permissions.yaml` — do not edit). ESLint forbids literal permission strings like `'user:list'` in `apps/admin/src/**/*.ts` (`no-restricted-syntax`), so string literals will fail `pnpm run lint:web`. Inside components, check access with `useAuthStore().hasPermission(code)`; `is_admin` users bypass all checks. A `vPermission` directive exists at `apps/admin/src/directives/permission.ts` but is currently **not registered** in `main.ts` — do not rely on it until that is fixed.

## State management

Both apps use Pinia setup stores:

- `apps/admin/src/stores/auth.ts` — session state. The access token is kept **in memory only**; the refresh token is persisted in `localStorage` under `scaffold.refreshToken`. `refresh()` is single-flight: concurrent calls share one pending request. `restore()` runs from the router guard and rebuilds the session after a page reload.
- `stores/plugin.ts` (both apps) — fetches plugin manifests, loads them through `@scaffold/plugin-runtime`, and tracks failures. `loaded` ensures the router guard loads plugins only once.

Keep server state out of stores where a plain composable or a view-level `ref` is enough.

## Calling the API

`packages/api-client/src/client.ts` builds an openapi-fetch client typed by the generated `schema.d.ts`. Apps instantiate it with an **empty base URL**, so all requests are same-origin (and hit the Vite proxy in dev).

- Use `api()` for authenticated calls: it injects `Authorization: Bearer <token>` and, on a 401, runs a single refresh-and-retry (guarded by an `x-scaffold-retried` header in admin).
- Use `rawClient` only for login, refresh, and logout — it has no interceptors.

Errors: convert with `toApiError(error)`, then map `error.reason` to the i18n key `errors.<REASON>`; `errorMessage()` in `apps/*/src/api/errors.ts` does this and falls back to the raw message. Remember the wire format is protojson: camelCase fields and 64-bit integers as strings.

After any proto change, run `pnpm run generate` and commit the regenerated `packages/api-client` outputs — never hand-edit files under `packages/api-client/generated/` or `packages/api-client/src/generated/`.

## Internationalization

vue-i18n 11 in composition mode, configured in `apps/*/src/i18n/index.ts`:

- Locales: `en-US` and `zh-CN`; default is `zh-CN`, fallback `en-US`.
- Persisted under the `localStorage` key `vela.locale`; switch with the exported `setLocale()`.
- Messages are namespaced objects: `app`, `nav`, `common`, `login`, `home`, `user`, `role`, `permission`, `plugin`, `profile`, `oauth`, `errors`.

When you add a page, add its keys to both `locales/en-US.ts` and `locales/zh-CN.ts` in the same namespace.

## Checklist: add a new page

- [ ] Create the view, for example `apps/admin/src/views/FeatureView.vue`, using `PageHeader` and the shared CSS classes.
- [ ] Register the route in `apps/admin/src/router/index.ts` with `titleKey` and, if needed, `permission: Permission.Something`.
- [ ] Add a menu entry to `coreMenu` in `apps/admin/src/router/menu.ts` with the same `titleKey` and permission.
- [ ] Add translations under the right namespace in both locale files.
- [ ] Wire the API call through `api()` and surface errors via `errorMessage()` plus `notifyApi`.

Verify with `pnpm run lint:web` and `pnpm run typecheck`.

## Known inconsistencies

The full list is maintained in [Architecture lead · Known inconsistencies](architecture-lead.md#known-inconsistencies). The ones that affect frontend work:

- `packages/shared/src/index.ts` exports `APP_NAME = "Vela"` and both apps persist the locale under `vela.locale`, while theme and refresh-token keys use the `scaffold.*` prefix.
- `apps/web/src/api/client.ts` uses the retry header `x-vela-retried`; `apps/admin` uses `x-scaffold-retried`.
- `vPermission` is defined but never registered in `apps/admin/src/main.ts`.
- `PageHeader` is imported per view, while `AppSnackbar` and `ConfirmDialog` are global — an intentional split, but easy to trip over.
- There are no frontend tests; every web package's `test` script is `echo no-tests`.

## Related

- [Architecture](../architecture.md) — layers, boundaries, generated outputs
- [Development](../development.md) — first run and the daily dev loop
- [Plugin and extension author guide](plugin-author.md) — runtime plugins and compile-time modules
- [Known inconsistencies](architecture-lead.md#known-inconsistencies) — tech debt that affects the frontend
