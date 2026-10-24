<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'

import { notifyApi } from '@scaffold/ui'

import { errorMessage } from '../api/errors'
import { useAuthStore } from '../stores/auth'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const failed = ref('')

onMounted(async () => {
  const ticket = typeof route.query.ticket === 'string' ? route.query.ticket : ''
  if (!ticket) {
    failed.value = t('oauth.failed')
    return
  }
  try {
    const response = await fetch('/api/v1/auth/oauth/exchange', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ticket }),
    })
    if (!response.ok) throw new Error(await response.text())
    const payload = (await response.json()) as { accessToken?: string; refreshToken?: string }
    if (!payload.accessToken || !payload.refreshToken) throw new Error('missing tokens')
    await auth.applyTokens(payload.accessToken, payload.refreshToken)
    await router.replace('/dashboard')
  } catch (error) {
    failed.value = errorMessage(error, (key) => t(key))
    notifyApi.error(failed.value)
  }
})
</script>

<template>
  <div class="callback-shell">
    <div class="blob blob--one" />
    <v-card class="callback-card glass" width="420">
      <v-card-text class="text-center py-10">
        <template v-if="!failed">
          <v-progress-circular indeterminate color="primary" size="56" width="5" class="mb-6" />
          <h2 class="text-h6 font-weight-bold mb-2">{{ t('oauth.callbackTitle') }}</h2>
          <p class="text-medium-emphasis">{{ t('oauth.exchanging') }}</p>
        </template>
        <template v-else>
          <v-icon icon="mdi-alert-circle-outline" size="56" color="error" class="mb-4" />
          <h2 class="text-h6 font-weight-bold mb-2">{{ t('oauth.failed') }}</h2>
          <v-alert type="error" variant="tonal" :text="failed" class="mb-4" />
          <v-btn color="primary" @click="$router.replace('/login')">{{ t('login.title') }}</v-btn>
        </template>
      </v-card-text>
    </v-card>
  </div>
</template>

<style scoped>
.callback-shell {
  position: relative;
  min-height: 100vh;
  display: grid;
  place-items: center;
  overflow: hidden;
  background:
    radial-gradient(700px 420px at 80% 0%, rgba(var(--v-theme-secondary), 0.18), transparent 60%),
    rgb(var(--v-theme-background));
}
.callback-card {
  position: relative;
  z-index: 2;
  border-radius: 26px;
}
.blob--one {
  top: -120px;
  right: -80px;
  width: 420px;
  height: 420px;
  background: rgba(99, 102, 241, 0.32);
}
</style>
