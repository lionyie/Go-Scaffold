<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import { api } from '../api/client'
import { useAuthStore } from '../stores/auth'

const { t } = useI18n()
const auth = useAuthStore()
const userCount = ref<number | null>(null)
const roleCount = ref<number | null>(null)
const pluginCount = ref<number | null>(null)

onMounted(async () => {
  const [users, roles, plugins] = await Promise.all([
    api().GET('/api/v1/users', { params: { query: { pageSize: 1 } } }),
    api().GET('/api/v1/roles', { params: { query: { pageSize: 1 } } }),
    api().GET('/api/v1/plugins'),
  ])
  userCount.value = users.data?.totalSize ?? 0
  roleCount.value = roles.data?.totalSize ?? 0
  pluginCount.value = plugins.data?.items?.length ?? 0
})
</script>

<template>
  <div>
    <el-row :gutter="16">
      <el-col :span="6">
        <el-card shadow="never">
          <div class="metric-label">{{ t('nav.users') }}</div>
          <div class="metric-value">{{ userCount ?? '—' }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <div class="metric-label">{{ t('nav.roles') }}</div>
          <div class="metric-value">{{ roleCount ?? '—' }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <div class="metric-label">{{ t('nav.plugins') }}</div>
          <div class="metric-value">{{ pluginCount ?? '—' }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <div class="metric-label">{{ t('nav.profile') }}</div>
          <div class="metric-value">{{ auth.user?.username ?? '—' }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="never" style="margin-top: 16px">
      <template #header>{{ t('profile.title') }}</template>
      <el-descriptions :column="2" border>
        <el-descriptions-item :label="t('common.displayName')">
          {{ auth.user?.displayName || auth.user?.username }}
        </el-descriptions-item>
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
  </div>
</template>

<style scoped>
.metric-label {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
.metric-value {
  margin-top: 8px;
  font-size: 28px;
  font-weight: 600;
}
</style>
