import type { PluginContext } from '@scaffold/plugin-runtime'

let current: PluginContext | null = null

export function setPluginContext(ctx: PluginContext): void {
  current = ctx
}

export function clearPluginContext(): void {
  current = null
}

export function usePluginContext(): PluginContext {
  if (!current) {
    throw new Error('hello-panel context is not ready')
  }
  return current
}
