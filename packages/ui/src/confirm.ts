import { reactive } from 'vue'

export const confirmState = reactive({
  show: false,
  title: '',
  message: '',
  confirmText: 'Confirm',
  cancelText: 'Cancel',
  resolve: null as ((value: boolean) => void) | null,
})

export interface ConfirmOptions {
  confirmText?: string
  cancelText?: string
}

export function confirmDialog(title: string, message: string, options: ConfirmOptions = {}): Promise<boolean> {
  return new Promise((resolve) => {
    confirmState.title = title
    confirmState.message = message
    confirmState.confirmText = options.confirmText ?? 'Confirm'
    confirmState.cancelText = options.cancelText ?? 'Cancel'
    confirmState.resolve = resolve
    confirmState.show = true
  })
}

export function settleConfirm(value: boolean): void {
  confirmState.show = false
  const resolve = confirmState.resolve
  confirmState.resolve = null
  resolve?.(value)
}
