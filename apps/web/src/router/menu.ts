import { featureMenu } from './features.gen'

export interface MenuItem {
  titleKey?: string
  label?: string
  path: string
}

export const coreMenu: MenuItem[] = [
  { titleKey: 'nav.home', path: '/' },
  { titleKey: 'nav.plugins', path: '/plugins' },
]

export const menu: MenuItem[] = [...coreMenu, ...featureMenu]
