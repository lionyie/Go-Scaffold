// Generated outputs and the files/directories that own them. `generate.mjs`
// clears codegenDirs only; extension registries are owned by scripts/module.mjs
// and must survive a plain generate.
export const codegenDirs = [
  { path: 'services/api', keep: ['services/api/identity/v1/redact.go'] },
  { path: 'services/internal/conf/conf.pb.go' },
  { path: 'services/internal/pkg/rbac/permissions_gen.go' },
  { path: 'packages/api-client/generated' },
  { path: 'packages/api-client/src/generated' },
  { path: 'packages/shared/src/permissions.gen.ts' },
]

export const extensionDirs = [
  'services/internal/extensions/registry_gen',
  'services/internal/auth/oauthproviders_gen',
  'services/internal/server/authz/extensions_gen.go',
  'apps/admin/src/router/features.gen.ts',
  'apps/web/src/router/features.gen.ts',
]

export const generatedDirs = [...codegenDirs.map((item) => item.path), ...extensionDirs]
