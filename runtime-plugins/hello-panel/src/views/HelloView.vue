<script setup lang="ts">
import { computed, ref } from 'vue'

import { usePluginContext } from '../context'

const ctx = usePluginContext()
const loading = ref(false)
const result = ref<Record<string, unknown> | null>(null)
const error = ref('')
const canCallApi = computed(() => Boolean(ctx.host.getAccessToken()))

async function callApi(): Promise<void> {
  loading.value = true
  error.value = ''
  try {
    result.value = await ctx.http.get<Record<string, unknown>>('/hello.json')
    ctx.notify.success('Plugin API responded')
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err)
    ctx.notify.error('Plugin API call failed')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="hello-panel">
    <VCard class="hello-card">
      <div class="hello-card__glow" />
      <VCardTitle class="text-h5 font-weight-bold pt-7 px-7">
        <VIcon icon="mdi-hand-wave" color="primary" class="mr-2" />
        {{ ctx.manifest.name }}
      </VCardTitle>
      <VCardText class="px-7 pb-7">
        <p class="text-medium-emphasis">{{ ctx.manifest.description }}</p>

        <div class="d-flex flex-wrap ga-2 my-5">
          <VChip color="primary" variant="tonal" prepend-icon="mdi-puzzle-outline">
            {{ ctx.manifest.id }}
          </VChip>
          <VChip color="secondary" variant="tonal" prepend-icon="mdi-tag-outline">
            v{{ ctx.manifest.version }}
          </VChip>
          <VChip color="accent" variant="tonal" prepend-icon="mdi-flash-outline">
            runtime plugin
          </VChip>
        </div>

        <VDivider class="my-5" />

        <div class="text-subtitle-1 font-weight-bold mb-2">Static backend</div>
        <p class="text-body-2 text-medium-emphasis">
          This page calls
          <code>/proxy/plugins/{{ ctx.manifest.id }}/api/hello.json</code>
          through the authenticated plugin proxy.
        </p>

        <VAlert
          v-if="!canCallApi"
          type="info"
          variant="tonal"
          class="my-5"
          icon="mdi-lock-outline"
        >
          The public site has no access token. Sign in to the admin console to call the plugin API.
        </VAlert>

        <div class="d-flex flex-wrap ga-3 my-5">
          <VBtn
            color="primary"
            prepend-icon="mdi-api"
            :loading="loading"
            :disabled="!canCallApi"
            @click="callApi"
          >
            Call plugin API
          </VBtn>
          <VBtn
            variant="tonal"
            prepend-icon="mdi-refresh"
            :disabled="!canCallApi || loading"
            @click="callApi"
          >
            Refresh
          </VBtn>
        </div>

        <VAlert v-if="error" type="error" variant="tonal" class="mb-5">
          {{ error }}
        </VAlert>

        <VCard v-if="result" variant="tonal" class="result-card">
          <VCardText>
            <pre class="hello-result">{{ JSON.stringify(result, null, 2) }}</pre>
          </VCardText>
        </VCard>
      </VCardText>
    </VCard>
  </div>
</template>
