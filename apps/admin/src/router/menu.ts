import { Permission } from '@scaffold/shared'

import { featureMenu } from './features.gen'

export interface MenuItem {
  titleKey?: string
  label?: string
  path: string
  permission?: string
}

export const coreMenu: MenuItem[] = [
  { titleKey: 'nav.dashboard', path: '/dashboard' },
  { titleKey: 'nav.users', path: '/users', permission: Permission.UserList },
  { titleKey: 'nav.roles', path: '/roles', permission: Permission.RoleList },
  { titleKey: 'nav.permissions', path: '/permissions', permission: Permission.RolePermissionList },
  { titleKey: 'nav.plugins', path: '/plugins', permission: Permission.PluginList },
  { titleKey: 'nav.profile', path: '/profile' },
]

export const menu: MenuItem[] = [...coreMenu, ...featureMenu]
