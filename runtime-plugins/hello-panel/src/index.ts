import type { PluginContext } from '@scaffold/plugin-runtime'

import HelloView from './views/HelloView.vue'
import { clearPluginContext, setPluginContext } from './context'
import './style.css'

export const components = { HelloView }

export function activate(ctx: PluginContext): void {
  setPluginContext(ctx)
  ctx.notify.success(`Plugin ${ctx.manifest.name} activated`)
}

export function deactivate(): void {
  clearPluginContext()
}
