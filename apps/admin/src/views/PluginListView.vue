<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'

import { Permission } from '@scaffold/shared'
import { PageHeader, confirmDialog, notifyApi } from '@scaffold/ui'

import { errorMessage } from '../api/errors'
import { api } from '../api/client'
import { usePluginStore } from '../stores/plugin'

const { t } = useI18n()
const router = useRouter()
const plugins = usePluginStore()

const dialog = ref(false)
const installTab = ref<'git' | 'upload'>('git')
const saving = ref(false)
const gitUrl = ref('')
const gitRef = ref('')
const replace = ref(false)
const archiveBase64 = ref('')
const archiveName = ref('')

const headers = computed(() => [
  { title: t('common.name'), key: 'id' },
  { title: t('common.displayName'), key: 'name' },
  { title: 'Version', key: 'version' },
  { title: t('plugin.source'), key: 'source', align: 'center' as const },
  { title: t('common.description'), key: 'description' },
  { title: t('plugin.entry'), key: 'entryUrl' },
  { title: t('common.actions'), key: 'actions', sortable: false, align: 'end' as const },
])

function sourceLabel(source?: string): string {
  return source === 'installed' ? t('plugin.sourceInstalled') : t('plugin.sourceBundled')
}

function openInstall(): void {
  installTab.value = 'git'
  gitUrl.value = ''
  gitRef.value = ''
  replace.value = false
  archiveBase64.value = ''
  archiveName.value = ''
  dialog.value = true
}

async function onFileChange(value: File | File[] | null): Promise<void> {
  const file = Array.isArray(value) ? value[0] : value
  if (!file) {
    archiveBase64.value = ''
    archiveName.value = ''
    return
  }
  if (file.size > 10 * 1024 * 1024) {
    notifyApi.error(t('plugin.fileTooLarge'))
    archiveBase64.value = ''
    archiveName.value = ''
    return
  }
  archiveBase64.value = await fileToBase64(file)
  archiveName.value = file.name
}

function fileToBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      const result = String(reader.result ?? '')
      resolve(result.slice(result.indexOf(',') + 1))
    }
    reader.onerror = () => reject(reader.error)
    reader.readAsDataURL(file)
  })
}

async function install(): Promise<void> {
  if (installTab.value === 'git' && !gitUrl.value.trim()) return
  if (installTab.value === 'upload' && !archiveBase64.value) return
  saving.value = true
  try {
    const body =
      installTab.value === 'git'
        ? { gitUrl: gitUrl.value.trim(), gitRef: gitRef.value.trim() || undefined, replace: replace.value }
        : { archiveBase64: archiveBase64.value, replace: replace.value }
    const { error } = await api().POST('/api/v1/plugins/install', { body })
    if (error) throw error
    notifyApi.success(t('plugin.installSuccess'))
    window.location.reload()
  } catch (error) {
    notifyApi.error(errorMessage(error, (key) => t(key)))
  } finally {
    saving.value = false
  }
}

async function uninstall(id: string): Promise<void> {
  const confirmed = await confirmDialog(t('common.confirm'), t('plugin.uninstallConfirm', { id }), {
    confirmText: t('plugin.uninstall'),
    cancelText: t('common.cancel'),
  })
  if (!confirmed) return
  const { error } = await api().DELETE('/api/v1/plugins/{id}', { params: { path: { id } } })
  if (error) {
    notifyApi.error(errorMessage(error, (key) => t(key)))
    return
  }
  notifyApi.success(t('plugin.uninstallSuccess'))
  window.location.reload()
}

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
        <v-btn
          v-permission="Permission.PluginManage"
          color="primary"
          prepend-icon="mdi-puzzle-plus"
          @click="openInstall"
        >
          {{ t('plugin.install') }}
        </v-btn>
      </template>
    </PageHeader>

    <v-card class="pa-2">
      <v-data-table :headers="headers" :items="plugins.manifests" :loading="plugins.loading" hover>
        <template #item.id="{ item }">
          <v-chip color="primary" variant="tonal" size="small">{{ item.id }}</v-chip>
        </template>
        <template #item.source="{ item }">
          <v-chip
            :color="item.source === 'installed' ? 'success' : 'secondary'"
            variant="tonal"
            size="small"
          >
            {{ sourceLabel(item.source) }}
          </v-chip>
        </template>
        <template #item.actions="{ item }">
          <v-btn
            v-if="item.source === 'installed'"
            v-permission="Permission.PluginManage"
            icon="mdi-trash-can-outline"
            size="small"
            variant="text"
            color="error"
            @click="uninstall(item.id)"
          />
          <span v-else class="muted">—</span>
        </template>
        <template #no-data>
          <div class="empty-state">
            <v-icon icon="mdi-puzzle-outline" size="42" />
            <p>{{ t('common.empty') }}</p>
          </div>
        </template>
      </v-data-table>

      <v-alert
        v-if="plugins.failures.length > 0"
        type="error"
        variant="tonal"
        class="ma-4"
        :text="plugins.failures.map((item) => `${item.id}: ${item.error}`).join('; ')"
      />
    </v-card>

    <v-dialog v-model="dialog" max-width="560" transition="dialog-bottom-transition">
      <v-card>
        <v-card-title class="text-h6 font-weight-bold px-6 pt-5">
          {{ t('plugin.install') }}
        </v-card-title>
        <v-card-text class="px-6">
          <v-tabs v-model="installTab" color="primary" class="mb-4">
            <v-tab value="git">{{ t('plugin.installGit') }}</v-tab>
            <v-tab value="upload">{{ t('plugin.installUpload') }}</v-tab>
          </v-tabs>

          <v-window v-model="installTab">
            <v-window-item value="git">
              <v-text-field
                v-model="gitUrl"
                :label="t('plugin.gitUrl')"
                :hint="t('plugin.gitUrlHint')"
                persistent-hint
                class="mb-4"
              />
              <v-text-field v-model="gitRef" :label="t('plugin.gitRef')" class="mb-3" />
            </v-window-item>
            <v-window-item value="upload">
              <v-file-input
                :label="t('plugin.uploadFile')"
                accept=".zip,application/zip"
                prepend-icon="mdi-zip-box-outline"
                show-size
                class="mb-3"
                @update:model-value="onFileChange"
              />
              <p v-if="archiveName" class="muted">{{ archiveName }}</p>
            </v-window-item>
          </v-window>

          <v-switch v-model="replace" :label="t('plugin.replace')" color="primary" inset hide-details />
          <v-alert type="info" variant="tonal" density="compact" class="mt-4" :text="t('plugin.installHint')" />
        </v-card-text>
        <v-card-actions class="px-6 pb-5">
          <v-spacer />
          <v-btn variant="text" @click="dialog = false">{{ t('common.cancel') }}</v-btn>
          <v-btn color="primary" :loading="saving" @click="install">{{ t('common.save') }}</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
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

.muted {
  color: rgb(var(--v-theme-on-surface-variant));
  font-size: 0.85rem;
}
</style>
