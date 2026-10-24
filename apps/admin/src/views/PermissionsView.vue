<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'

import { errorMessage } from '../api/errors'
import { api } from '../api/client'

interface PermissionRow {
  code?: string
  name?: string
  resource?: string
  action?: string
}

const { t } = useI18n()
const rows = ref<PermissionRow[]>([])
const loading = ref(false)

async function load(): Promise<void> {
  loading.value = true
  try {
    const { data, error } = await api().GET('/api/v1/permissions')
    if (error) throw error
    rows.value = (data?.items ?? []) as PermissionRow[]
  } catch (error) {
    ElMessage.error(errorMessage(error, (key) => t(key)))
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="page-card">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>{{ t('permission.title') }}</span>
          <el-button @click="load">{{ t('common.refresh') }}</el-button>
        </div>
      </template>
      <el-table v-loading="loading" :data="rows" border>
        <el-table-column prop="code" :label="t('permission.code')" min-width="200" />
        <el-table-column prop="name" :label="t('common.name')" min-width="200" />
        <el-table-column prop="resource" :label="t('permission.resource')" min-width="140" />
        <el-table-column prop="action" :label="t('permission.action')" min-width="140" />
      </el-table>
    </el-card>
  </div>
</template>

<style scoped>
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
</style>
