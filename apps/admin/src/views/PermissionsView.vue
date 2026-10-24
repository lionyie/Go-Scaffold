<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import { PageHeader, notifyApi } from '@scaffold/ui'

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

const headers = computed(() => [
  { title: t('permission.code'), key: 'code' },
  { title: t('common.name'), key: 'name' },
  { title: t('permission.resource'), key: 'resource' },
  { title: t('permission.action'), key: 'action' },
])

async function load(): Promise<void> {
  loading.value = true
  try {
    const { data, error } = await api().GET('/api/v1/permissions')
    if (error) throw error
    rows.value = (data?.items ?? []) as PermissionRow[]
  } catch (error) {
    notifyApi.error(errorMessage(error, (key) => t(key)))
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="page-card">
    <PageHeader :title="t('permission.title')" :description="t('app.name')">
      <template #actions>
        <v-btn variant="tonal" prepend-icon="mdi-refresh" @click="load">{{ t('common.refresh') }}</v-btn>
      </template>
    </PageHeader>
    <v-card class="pa-2">
      <v-data-table :headers="headers" :items="rows" :loading="loading" item-value="code" hover>
        <template #item.code="{ item }">
          <v-chip color="primary" variant="tonal" size="small" prepend-icon="mdi-key-outline">
            {{ item.code }}
          </v-chip>
        </template>
      </v-data-table>
    </v-card>
  </div>
</template>
