export default {
  app: { name: 'Scaffold' },
  nav: { home: 'Home', plugins: 'Plugins' },
  home: {
    title: 'Welcome to Scaffold',
    subtitle: 'A multi-app starter with an admin console, RBAC and on-demand extensions.',
    apps: 'Apps',
    appsHint: 'Add web, admin, mobile or desktop apps under apps/.',
    extensions: 'Extensions',
    extensionsHint: 'Install features and runnable examples with node scripts/module.mjs.',
    admin: 'Admin',
    adminLink: 'Open the admin console',
  },
  plugin: {
    title: 'Plugins',
    installed: 'Installed runtime plugins (experimental)',
    entry: 'Entry',
  },
  common: {
    refresh: 'Refresh',
    empty: 'No data',
    loading: 'Loading',
  },
  errors: {
    UNAUTHORIZED: 'Please sign in first',
    FORBIDDEN: 'You do not have permission for this action',
    INTERNAL_ERROR: 'Internal error',
    INVALID_ARGUMENT: 'Invalid input',
    MUST_CHANGE_PASSWORD: 'You must change your password first',
  },
}
