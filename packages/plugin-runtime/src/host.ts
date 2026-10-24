import type { PluginHost } from './types'

declare global {
  interface Window {
    __SCAFFOLD_PLUGIN_HOST__?: PluginHost
  }
}

let current: PluginHost | null = null

export function installPluginHost(host: PluginHost): void {
  current = host
  window.__SCAFFOLD_PLUGIN_HOST__ = host
}

export function getPluginHost(): PluginHost {
  if (!current) {
    current = window.__SCAFFOLD_PLUGIN_HOST__ ?? null
  }
  if (!current) {
    throw new Error('plugin host is not installed')
  }
  return current
}
