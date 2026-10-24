import type { Directive } from 'vue'

import { useAuthStore } from '../stores/auth'

/** Removes the element when the current user lacks the given permission. */
export const vPermission: Directive<HTMLElement, string> = {
  mounted(el, binding) {
    if (binding.value && !useAuthStore().hasPermission(binding.value)) {
      el.remove()
    }
  },
}
