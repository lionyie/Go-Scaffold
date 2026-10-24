<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'

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
    const payload = (await response.json()) as {
      accessToken?: string
      refreshToken?: string
    }
    if (!payload.accessToken || !payload.refreshToken) throw new Error('missing tokens')
    await auth.applyTokens(payload.accessToken, payload.refreshToken)
    await router.replace('/dashboard')
  } catch (error) {
    failed.value = errorMessage(error, (key) => t(key))
    ElMessage.error(failed.value)
  }
})
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <h1>{{ t('oauth.callbackTitle') }}</h1>
      <p v-if="!failed">{{ t('oauth.exchanging') }}</p>
      <template v-else>
        <el-alert type="error" :closable="false" :title="failed" />
        <el-button style="margin-top: 16px" @click="$router.replace('/login')">
          {{ t('login.title') }}
        </el-button>
      </template>
    </div>
  </div>
</template>
