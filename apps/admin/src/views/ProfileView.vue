<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'

import { errorMessage } from '../api/errors'
import { useAuthStore } from '../stores/auth'

interface ProviderInfo {
  name: string
  displayName: string
}

const { t } = useI18n()
const auth = useAuthStore()
const providers = ref<ProviderInfo[]>([])
const form = reactive({ oldPassword: '', newPassword: '' })
const saving = ref(false)

async function loadProviders(): Promise<void> {
  try {
    const response = await fetch('/api/v1/auth/oauth/providers')
    if (!response.ok) return
    const payload = (await response.json()) as { providers?: ProviderInfo[] }
    providers.value = payload.providers ?? []
  } catch {
    providers.value = []
  }
}

async function changePassword(): Promise<void> {
  saving.value = true
  try {
    await auth.changePassword(form.oldPassword, form.newPassword)
    form.oldPassword = ''
    form.newPassword = ''
    ElMessage.success(t('profile.changed'))
  } catch (error) {
    ElMessage.error(errorMessage(error, (key) => t(key)))
  } finally {
    saving.value = false
  }
}

onMounted(loadProviders)
</script>

<template>
  <div class="page-card">
    <el-alert
      v-if="auth.user?.mustChangePassword"
      type="warning"
      :closable="false"
      :title="t('profile.mustChange')"
      style="margin-bottom: 16px"
    />

    <el-card shadow="never">
      <template #header>{{ t('profile.title') }}</template>
      <el-descriptions :column="2" border>
        <el-descriptions-item :label="t('common.name')">{{ auth.user?.username }}</el-descriptions-item>
        <el-descriptions-item :label="t('common.displayName')">
          {{ auth.user?.displayName || '—' }}
        </el-descriptions-item>
        <el-descriptions-item label="Email">{{ auth.user?.email || '—' }}</el-descriptions-item>
        <el-descriptions-item :label="t('user.admin')">
          {{ auth.user?.isAdmin ? 'yes' : 'no' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('user.roles')" :span="2">
          <el-tag v-for="role in auth.user?.roles ?? []" :key="role" style="margin-right: 6px">
            {{ role }}
          </el-tag>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <el-card shadow="never" style="margin-top: 16px">
      <template #header>{{ t('profile.changePassword') }}</template>
      <el-form label-position="top" style="max-width: 420px">
        <el-form-item :label="t('profile.oldPassword')">
          <el-input v-model="form.oldPassword" type="password" autocomplete="current-password" />
        </el-form-item>
        <el-form-item :label="t('profile.newPassword')">
          <el-input v-model="form.newPassword" type="password" autocomplete="new-password" />
        </el-form-item>
        <el-button type="primary" :loading="saving" @click="changePassword">
          {{ t('common.save') }}
        </el-button>
      </el-form>
    </el-card>

    <el-card v-if="providers.length > 0" shadow="never" style="margin-top: 16px">
      <template #header>{{ t('oauth.linked') }}</template>
      <el-tag v-for="provider in providers" :key="provider.name" style="margin-right: 8px">
        {{ provider.displayName }}
      </el-tag>
    </el-card>
  </div>
</template>
