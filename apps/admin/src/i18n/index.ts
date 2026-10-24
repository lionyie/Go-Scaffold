import { createI18n } from 'vue-i18n'

import enUS from './locales/en-US'
import zhCN from './locales/zh-CN'

const saved = localStorage.getItem('vela.locale')
const locale = saved === 'en-US' || saved === 'zh-CN' ? saved : 'zh-CN'

export const i18n = createI18n({
  legacy: false,
  locale,
  fallbackLocale: 'en-US',
  messages: {
    'en-US': enUS,
    'zh-CN': zhCN,
  },
})

export function setLocale(next: 'en-US' | 'zh-CN'): void {
  i18n.global.locale.value = next
  localStorage.setItem('vela.locale', next)
}

export function currentLocale(): 'en-US' | 'zh-CN' {
  return i18n.global.locale.value as 'en-US' | 'zh-CN'
}
