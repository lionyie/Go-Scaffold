<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'

import { PageHeader } from '@scaffold/ui'

import { usePluginStore } from '../stores/plugin'

const { t } = useI18n()
const router = useRouter()
const plugins = usePluginStore()

async function refresh(): Promise<void> {
  await plugins.loadAll(router)
}
</script>

<template>
  <div class="page-card">
    <PageHeader :title="t('plugin.title')" :description="t('plugin.installed')">
      <template #actions>
        <el-button @click="refresh">{{ t('common.refresh') }}</el-button>
      </template>
    </PageHeader>

    <el-table :data="plugins.manifests" border>
      <el-table-column prop="id" :label="t('common.name')" min-width="160" />
      <el-table-column prop="name" :label="t('common.displayName')" min-width="160" />
      <el-table-column prop="version" label="Version" width="120" />
      <el-table-column prop="description" :label="t('common.description')" min-width="240" />
      <el-table-column prop="entryUrl" :label="t('plugin.entry')" min-width="260" />
      <template #empty>
        <div class="empty-state">{{ t('common.empty') }}</div>
      </template>
    </el-table>

    <el-alert
      v-if="plugins.failures.length > 0"
      type="error"
      :closable="false"
      style="margin-top: 12px"
      :title="plugins.failures.map((item) => `${item.id}: ${item.error}`).join('; ')"
    />
  </div>
</template>
