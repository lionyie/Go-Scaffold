<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'

import { errorMessage } from '../api/errors'
import { useAuthStore } from '../stores/auth'

interface ProviderInfo {
  name: string
  displayName: string
}

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const { t } = useI18n()

const form = reactive({ username: 'admin', password: 'admin123' })
const loading = ref(false)
const providers = ref<ProviderInfo[]>([])

async function submit(): Promise<void> {
  loading.value = true
  try {
    await auth.login(form.username, form.password)
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/dashboard'
    await router.replace(redirect)
  } catch (error) {
    ElMessage.error(errorMessage(error, (key) => t(key)))
  } finally {
    loading.value = false
  }
}

async function startOAuth(provider: string): Promise<void> {
  try {
    const response = await fetch(`/api/v1/auth/oauth/${encodeURIComponent(provider)}/start`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: '{}',
    })
    if (!response.ok) throw new Error(await response.text())
    const payload = (await response.json()) as { authUrl?: string }
    if (!payload.authUrl) throw new Error('missing authUrl')
    window.location.href = payload.authUrl
  } catch (error) {
    ElMessage.error(errorMessage(error, (key) => t(key)))
  }
}

onMounted(async () => {
  try {
    const response = await fetch('/api/v1/auth/oauth/providers')
    if (!response.ok) return
    const payload = (await response.json()) as { providers?: ProviderInfo[] }
    providers.value = payload.providers ?? []
  } catch {
    providers.value = []
  }
})
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <h1>{{ t('login.title') }}</h1>
      <p>{{ t('login.subtitle') }}</p>
      <el-form label-position="top" @submit.prevent="submit">
        <el-form-item :label="t('login.username')">
          <el-input v-model="form.username" autocomplete="username" />
        </el-form-item>
        <el-form-item :label="t('login.password')">
          <el-input v-model="form.password" type="password" autocomplete="current-password" />
        </el-form-item>
        <el-button type="primary" native-type="submit" :loading="loading" style="width: 100%">
          {{ t('login.submit') }}
        </el-button>
      </el-form>

      <div v-if="providers.length > 0" class="oauth">
        <el-divider>{{ t('login.oauth') }}</el-divider>
        <el-button
          v-for="provider in providers"
          :key="provider.name"
          style="width: 100%; margin-bottom: 8px"
          @click="startOAuth(provider.name)"
        >
          {{ provider.displayName }}
        </el-button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.oauth {
  margin-top: 8px;
}
</style>
