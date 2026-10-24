<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import { PageHeader, confirmDialog, notifyApi } from '@scaffold/ui'

import { errorMessage } from '../api/errors'
import { api } from '../api/client'

interface RoleRow {
  id?: string
  name: string
  displayName?: string
  description?: string
  isSystem?: boolean
  permissions?: string[]
}

interface PermissionRow {
  code?: string
  name?: string
  resource?: string
  action?: string
}

const { t } = useI18n()
const rows = ref<RoleRow[]>([])
const permissions = ref<PermissionRow[]>([])
const loading = ref(false)
const dialog = ref(false)
const editing = ref<RoleRow | null>(null)
const form = reactive({
  name: '',
  displayName: '',
  description: '',
  permissions: [] as string[],
})

const headers = computed(() => [
  { title: t('common.name'), key: 'name', sortable: true },
  { title: t('common.displayName'), key: 'displayName' },
  { title: t('common.description'), key: 'description' },
  { title: t('role.permissions'), key: 'permissions', sortable: false },
  { title: t('role.system'), key: 'isSystem', align: 'center' as const },
  { title: t('common.actions'), key: 'actions', sortable: false, align: 'end' as const },
])

const permissionItems = computed(() =>
  permissions.value
    .filter((permission) => permission.code)
    .map((permission) => ({ title: permission.code!, value: permission.code! })),
)

async function load(): Promise<void> {
  loading.value = true
  try {
    const roles: any = await api().GET('/api/v1/roles', { params: { query: { pageSize: 100 } } })
    const perms: any = await api().GET('/api/v1/permissions')
    if (roles.error) throw roles.error
    if (perms.error) throw perms.error
    rows.value = (roles.data?.items ?? []) as RoleRow[]
    permissions.value = (perms.data?.items ?? []) as PermissionRow[]
  } catch (error) {
    notifyApi.error(errorMessage(error, (key) => t(key)))
  } finally {
    loading.value = false
  }
}

function openCreate(): void {
  editing.value = null
  Object.assign(form, { name: '', displayName: '', description: '', permissions: [] })
  dialog.value = true
}

function openEdit(row: RoleRow): void {
  editing.value = row
  Object.assign(form, {
    name: row.name,
    displayName: row.displayName ?? '',
    description: row.description ?? '',
    permissions: [...(row.permissions ?? [])],
  })
  dialog.value = true
}

async function save(): Promise<void> {
  const role = {
    name: form.name,
    displayName: form.displayName,
    description: form.description,
    permissions: form.permissions,
  }
  try {
    if (editing.value?.id) {
      const { error } = await api().PUT('/api/v1/roles/{id}', {
        params: { path: { id: editing.value.id } },
        body: { id: editing.value.id, role },
      })
      if (error) throw error
    } else {
      const { error } = await api().POST('/api/v1/roles', { body: { role } })
      if (error) throw error
    }
    dialog.value = false
    notifyApi.success(t('common.save'))
    await load()
  } catch (error) {
    notifyApi.error(errorMessage(error, (key) => t(key)))
  }
}

async function remove(row: RoleRow): Promise<void> {
  const confirmed = await confirmDialog(t('common.confirm'), t('role.deleteConfirm', { name: row.name }), {
    confirmText: t('common.delete'),
    cancelText: t('common.cancel'),
  })
  if (!confirmed) return
  const { error } = await api().DELETE('/api/v1/roles/{id}', {
    params: { path: { id: row.id! } },
  })
  if (error) {
    notifyApi.error(errorMessage(error, (key) => t(key)))
    return
  }
  notifyApi.success(t('common.delete'))
  await load()
}

onMounted(load)
</script>

<template>
  <div class="page-card">
    <PageHeader :title="t('role.title')" :description="t('app.name')">
      <template #actions>
        <v-btn variant="tonal" prepend-icon="mdi-refresh" @click="load">{{ t('common.refresh') }}</v-btn>
        <v-btn color="primary" prepend-icon="mdi-shield-plus" @click="openCreate">
          {{ t('role.create') }}
        </v-btn>
      </template>
    </PageHeader>

    <v-card class="pa-2">
      <v-data-table :headers="headers" :items="rows" :loading="loading" item-value="id" hover>
        <template #item.permissions="{ item }">
          <v-chip size="small" color="secondary" variant="tonal">
            {{ (item.permissions ?? []).length }}
          </v-chip>
        </template>
        <template #item.isSystem="{ item }">
          <v-icon
            :icon="item.isSystem ? 'mdi-lock-check' : 'mdi-lock-open-variant-outline'"
            :color="item.isSystem ? 'warning' : 'medium-emphasis'"
          />
        </template>
        <template #item.actions="{ item }">
          <v-btn icon="mdi-pencil-outline" size="small" variant="text" @click="openEdit(item)" />
          <v-btn
            icon="mdi-trash-can-outline"
            size="small"
            variant="text"
            color="error"
            :disabled="item.isSystem"
            @click="remove(item)"
          />
        </template>
        <template #no-data>
          <div class="empty-state">
            <v-icon icon="mdi-shield-off-outline" size="42" />
            <p>{{ t('common.empty') }}</p>
          </div>
        </template>
      </v-data-table>
    </v-card>

    <v-dialog v-model="dialog" max-width="640" transition="dialog-bottom-transition">
      <v-card>
        <v-card-title class="text-h6 font-weight-bold px-6 pt-5">
          {{ editing ? t('common.edit') : t('role.create') }}
        </v-card-title>
        <v-card-text class="px-6">
          <v-form @submit.prevent="save">
            <v-text-field v-model="form.name" :label="t('common.name')" :disabled="Boolean(editing)" class="mb-3" />
            <v-text-field v-model="form.displayName" :label="t('common.displayName')" class="mb-3" />
            <v-textarea v-model="form.description" :label="t('common.description')" rows="2" class="mb-3" />
            <v-select
              v-model="form.permissions"
              :items="permissionItems"
              :label="t('role.permissions')"
              multiple
              chips
              closable-chips
              density="comfortable"
            />
          </v-form>
        </v-card-text>
        <v-card-actions class="px-6 pb-5">
          <v-spacer />
          <v-btn variant="text" @click="dialog = false">{{ t('common.cancel') }}</v-btn>
          <v-btn color="primary" @click="save">{{ t('common.save') }}</v-btn>
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
</style>
