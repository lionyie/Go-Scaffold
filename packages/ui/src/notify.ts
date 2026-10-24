import { reactive } from 'vue'

export type NotifyColor = 'success' | 'error' | 'warning' | 'info'

export const notifyState = reactive({
  show: false,
  text: '',
  color: 'success' as NotifyColor,
  icon: 'mdi-check-circle',
})

const icons: Record<NotifyColor, string> = {
  success: 'mdi-check-circle',
  error: 'mdi-alert-circle',
  warning: 'mdi-alert',
  info: 'mdi-information',
}

export function notify(text: string, color: NotifyColor = 'success'): void {
  notifyState.text = text
  notifyState.color = color
  notifyState.icon = icons[color]
  notifyState.show = false
  requestAnimationFrame(() => {
    notifyState.show = true
  })
}

export const notifyApi = {
  success: (text: string) => notify(text, 'success'),
  error: (text: string) => notify(text, 'error'),
  warning: (text: string) => notify(text, 'warning'),
  info: (text: string) => notify(text, 'info'),
}
