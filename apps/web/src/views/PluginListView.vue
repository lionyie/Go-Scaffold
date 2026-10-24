<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'

import { PageHeader } from '@scaffold/ui'

import { usePluginStore } from '../stores/plugin'

const { t } = useI18n()
const router = useRouter()
const plugins = usePluginStore()

const headers = computed(() => [
  { title: t('common.name'), key: 'id' },
  { title: t('common.displayName'), key: 'name' },
  { title: 'Version', key: 'version' },
  { title: t('common.description'), key: 'description' },
])

onMounted(() => {
  plugins.loadAll(router).catch(() => undefined)
})
</script>

<template>
  <div class="page-card">
    <PageHeader :title="t('plugin.title')" :description="t('plugin.installed')">
      <template #actions>
        <v-btn variant="tonal" prepend-icon="mdi-refresh" @click="plugins.loadAll(router)">
          {{ t('common.refresh') }}
        </v-btn>
      </template>
    </PageHeader>
    <v-card class="pa-2">
      <v-data-table :headers="headers" :items="plugins.manifests" :loading="plugins.loading" hover>
        <template #item.id="{ item }">
          <v-chip color="primary" variant="tonal" size="small">{{ item.id }}</v-chip>
        </template>
        <template #no-data>
          <div class="empty-state">
            <v-icon icon="mdi-puzzle-outline" size="42" />
            <p>{{ t('common.empty') }}</p>
          </div>
        </template>
      </v-data-table>
    </v-card>
  </div>
</template>

<style scoped>
.empty-state {
  display: grid;
  place-items: center;
  gap: 8px;
  padding: 40px 0;
  color: rgb(var(--v-theme-on-surface-variant));
}
</style>
