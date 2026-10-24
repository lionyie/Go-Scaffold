import { defineStore } from 'pinia'
import { ref } from 'vue'

import type { PluginManifest, PluginNav } from '@scaffold/plugin-runtime'
import { loadPlugins } from '@scaffold/plugin-runtime'

import type { Router } from 'vue-router'

import { toApiError } from '@scaffold/api-client'

import { api } from '../api/client'

export interface PluginNavItem extends PluginNav {
  pluginId: string
}

export interface PluginFailure {
  id: string
  error: string
}

export const usePluginStore = defineStore('plugin', () => {
  const manifests = ref<PluginManifest[]>([])
  const navigations = ref<PluginNavItem[]>([])
  const failures = ref<PluginFailure[]>([])
  const loaded = ref(false)
  const loading = ref(false)

  async function loadAll(router: Router): Promise<void> {
    if (loaded.value || loading.value) return
    loading.value = true
    try {
      const { data, error } = await api().GET('/api/v1/plugins')
      if (error || !data) throw toApiError(error)

      manifests.value = (data.items ?? []) as unknown as PluginManifest[]
      const results = await loadPlugins(manifests.value, {
        router,
        onError: (manifest, loadError) => {
          failures.value.push({ id: manifest.id, error: String(loadError) })
        },
      })

      const failedIDs = new Set(
        results.filter((result) => result.status === 'failed').map((result) => result.manifest.id),
      )
      const items: PluginNavItem[] = []
      for (const manifest of manifests.value) {
        if (failedIDs.has(manifest.id)) continue
        for (const nav of manifest.navigations ?? []) {
          items.push({ ...nav, pluginId: manifest.id })
        }
      }
      navigations.value = items.sort((a, b) => a.order - b.order)
      loaded.value = true
    } finally {
      loading.value = false
    }
  }

  return { manifests, navigations, failures, loaded, loading, loadAll }
})
