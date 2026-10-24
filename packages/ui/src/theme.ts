import { createVuetify } from 'vuetify'
import type { ThemeDefinition } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'

const scaffoldLight: ThemeDefinition = {
  dark: false,
  colors: {
    background: '#F3F5FF',
    surface: '#FFFFFF',
    'surface-bright': '#FFFFFF',
    'surface-light': '#EEF2FF',
    'surface-variant': '#E0E7FF',
    'on-surface-variant': '#475569',
    'on-background': '#0F172A',
    'on-surface': '#0F172A',
    primary: '#6366F1',
    'primary-darken-1': '#4F46E5',
    secondary: '#06B6D4',
    'secondary-darken-1': '#0891B2',
    accent: '#F472B6',
    success: '#10B981',
    warning: '#F59E0B',
    error: '#EF4444',
    info: '#3B82F6',
  },
}

const scaffoldDark: ThemeDefinition = {
  dark: true,
  colors: {
    background: '#070B18',
    surface: '#0F172A',
    'surface-bright': '#16213B',
    'surface-light': '#1E293B',
    'surface-variant': '#25324D',
    'on-surface-variant': '#94A3B8',
    'on-background': '#E2E8F0',
    'on-surface': '#E2E8F0',
    primary: '#818CF8',
    'primary-darken-1': '#6366F1',
    secondary: '#22D3EE',
    'secondary-darken-1': '#06B6D4',
    accent: '#F9A8D4',
    success: '#34D399',
    warning: '#FBBF24',
    error: '#F87171',
    info: '#60A5FA',
  },
}

const defaults = {
  VBtn: { rounded: 'pill', elevation: 0, class: 'font-weight-semibold' },
  VCard: { rounded: 'xl', elevation: 0 },
  VTextField: {
    variant: 'outlined',
    density: 'comfortable',
    rounded: 'lg',
    hideDetails: 'auto',
    color: 'primary',
  },
  VSelect: {
    variant: 'outlined',
    density: 'comfortable',
    rounded: 'lg',
    hideDetails: 'auto',
    color: 'primary',
  },
  VTextarea: { variant: 'outlined', density: 'comfortable', rounded: 'lg', hideDetails: 'auto' },
  VDialog: { maxWidth: 560 },
  VChip: { rounded: 'pill' },
  VDataTable: { density: 'comfortable' },
  VNavigationDrawer: { elevation: 0 },
  VAppBar: { elevation: 0 },
}

export function createScaffoldVuetify(defaultTheme: 'scaffoldLight' | 'scaffoldDark' = 'scaffoldLight') {
  return createVuetify({
    components,
    directives,
    icons: { defaultSet: 'mdi' },
    theme: {
      defaultTheme,
      themes: { scaffoldLight, scaffoldDark },
    },
    defaults,
  })
}

export type ScaffoldTheme = 'scaffoldLight' | 'scaffoldDark'
