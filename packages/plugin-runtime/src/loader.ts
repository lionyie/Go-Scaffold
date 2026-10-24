import { defineComponent, h } from 'vue'
import type { RouteRecordRaw, Router } from 'vue-router'

import { getPluginHost } from './host'
import { createPluginHttp } from './http'
import type {
  PluginContext,
  PluginLoadResult,
  PluginManifest,
  PluginNotify,
  VelaPluginModule,
} from './types'

export interface LoadPluginsOptions {
  router: Router
  onNavigation?(manifest: PluginManifest): void
  onError?(manifest: PluginManifest, error: unknown): void
}

const MissingComponent = defineComponent({
  name: 'PluginComponentMissing',
  render() {
    return h('div', { class: 'plugin-missing' }, 'Plugin component failed to load')
  },
})

function injectStyles(manifest: PluginManifest): void {
  for (const href of manifest.styles) {
    if (document.querySelector(`link[data-vela-plugin="${manifest.id}"][href="${href}"]`)) continue
    const link = document.createElement('link')
    link.rel = 'stylesheet'
    link.href = href
    link.dataset.velaPlugin = manifest.id
    document.head.appendChild(link)
  }
}

export async function loadPlugins(
  manifests: PluginManifest[],
  options: LoadPluginsOptions,
): Promise<PluginLoadResult[]> {
  const host = getPluginHost()
  const notify: PluginNotify = host.notify
  const results: PluginLoadResult[] = []

  for (const manifest of manifests) {
    try {
      injectStyles(manifest)
      const separator = manifest.entryUrl.includes('?') ? '&' : '?'
      const url = `${manifest.entryUrl}${separator}v=${encodeURIComponent(manifest.version)}`
      const module = (await import(/* @vite-ignore */ url)) as VelaPluginModule
      const components = module.components ?? {}
      const removers: Array<() => void> = []

      for (const route of manifest.routes) {
        const record: RouteRecordRaw = {
          path: route.path,
          name: route.name,
          component: components[route.component] ?? MissingComponent,
          meta: { permission: route.permission, pluginId: manifest.id },
        }
        removers.push(options.router.addRoute('root', record))
      }

      const context: PluginContext = {
        manifest,
        host,
        http: createPluginHttp(`/proxy/plugins/${manifest.id}/api`, () => host.getAccessToken()),
        t: (key, ...args) => host.i18n.global.t(key, ...args),
        hasPermission: (code) => host.hasPermission(code),
        registerRoute: (record) => options.router.addRoute('root', record),
        notify,
      }

      await module.activate?.(context)
      options.onNavigation?.(manifest)
      results.push({ manifest, status: 'loaded' })
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error)
      host.logger.error(`failed to load plugin ${manifest.id}`, error)
      notify.error(`Plugin ${manifest.name} failed to load`)
      options.onError?.(manifest, error)
      results.push({ manifest, status: 'failed', error: message })
    }
  }

  return results
}
