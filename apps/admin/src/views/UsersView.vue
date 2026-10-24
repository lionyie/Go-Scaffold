<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import { PageHeader, confirmDialog, notifyApi } from '@scaffold/ui'

import { errorMessage } from '../api/errors'
import { api } from '../api/client'

interface UserRow {
  id?: string
  username: string
  displayName?: string
  email?: string
  isAdmin?: boolean
  status?: string
  roles?: string[]
}

const { t } = useI18n()
const rows = ref<UserRow[]>([])
const loading = ref(false)
const search = ref('')
const dialog = ref(false)
const editing = ref<UserRow | null>(null)
const form = reactive({
  username: '',
  displayName: '',
  email: '',
  password: '',
  roles: '',
  isAdmin: false,
  status: 'active',
})

const headers = computed(() => [
  { title: t('common.name'), key: 'username', sortable: true },
  { title: t('common.displayName'), key: 'displayName' },
  { title: 'Email', key: 'email' },
  { title: t('user.roles'), key: 'roles', sortable: false },
  { title: t('user.admin'), key: 'isAdmin', align: 'center' as const },
  { title: t('common.actions'), key: 'actions', sortable: false, align: 'end' as const },
])

async function load(): Promise<void> {
  loading.value = true
  try {
    const { data, error } = await api().GET('/api/v1/users', {
      params: { query: { pageSize: 100 } },
    })
    if (error) throw error
    rows.value = (data?.items ?? []) as UserRow[]
  } catch (error) {
    notifyApi.error(errorMessage(error, (key) => t(key)))
  } finally {
    loading.value = false
  }
}

function openCreate(): void {
  editing.value = null
  Object.assign(form, {
    username: '',
    displayName: '',
    email: '',
    password: '',
    roles: '',
    isAdmin: false,
    status: 'active',
  })
  dialog.value = true
}

function openEdit(row: UserRow): void {
  editing.value = row
  Object.assign(form, {
    username: row.username,
    displayName: row.displayName ?? '',
    email: row.email ?? '',
    password: '',
    roles: (row.roles ?? []).join(','),
    isAdmin: Boolean(row.isAdmin),
    status: row.status ?? 'active',
  })
  dialog.value = true
}

async function save(): Promise<void> {
  const user = {
    username: form.username,
    displayName: form.displayName,
    email: form.email,
    isAdmin: form.isAdmin,
    status: form.status,
    roles: form.roles
      .split(',')
      .map((item) => item.trim())
      .filter(Boolean),
  }
  try {
    if (editing.value?.id) {
      const { error } = await api().PUT('/api/v1/users/{id}', {
        params: { path: { id: editing.value.id } },
        body: { id: editing.value.id, user, password: form.password || undefined },
      })
      if (error) throw error
    } else {
      const { error } = await api().POST('/api/v1/users', {
        body: { user, password: form.password },
      })
      if (error) throw error
    }
    dialog.value = false
    notifyApi.success(t('common.save'))
    await load()
  } catch (error) {
    notifyApi.error(errorMessage(error, (key) => t(key)))
  }
}

async function remove(row: UserRow): Promise<void> {
  const confirmed = await confirmDialog(t('common.confirm'), t('user.deleteConfirm', { name: row.username }), {
    confirmText: t('common.delete'),
    cancelText: t('common.cancel'),
  })
  if (!confirmed) return
  const { error } = await api().DELETE('/api/v1/users/{id}', {
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
    <PageHeader :title="t('user.title')" :description="t('app.name')">
      <template #actions>
        <v-btn variant="tonal" prepend-icon="mdi-refresh" @click="load">{{ t('common.refresh') }}</v-btn>
        <v-btn color="primary" prepend-icon="mdi-account-plus" @click="openCreate">
          {{ t('user.create') }}
        </v-btn>
      </template>
    </PageHeader>

    <v-card class="pa-2">
      <div class="table-toolbar">
        <v-text-field
          v-model="search"
          :label="t('common.search')"
          prepend-inner-icon="mdi-magnify"
          density="compact"
          hide-details
          variant="solo-filled"
          flat
          rounded="pill"
          class="search-field"
        />
      </div>

      <v-data-table
        :headers="headers"
        :items="rows"
        :loading="loading"
        :search="search"
        item-value="id"
        hover
        class="users-table"
      >
        <template #item.roles="{ item }">
          <v-chip
            v-for="role in item.roles ?? []"
            :key="role"
            size="small"
            color="primary"
            variant="tonal"
            class="mr-1"
          >
            {{ role }}
          </v-chip>
        </template>
        <template #item.isAdmin="{ item }">
          <v-icon
            :icon="item.isAdmin ? 'mdi-check-decagram' : 'mdi-minus-circle-outline'"
            :color="item.isAdmin ? 'success' : 'medium-emphasis'"
          />
        </template>
        <template #item.actions="{ item }">
          <v-btn icon="mdi-pencil-outline" size="small" variant="text" @click="openEdit(item)" />
          <v-btn
            icon="mdi-trash-can-outline"
            size="small"
            variant="text"
            color="error"
            @click="remove(item)"
          />
        </template>
        <template #no-data>
          <div class="empty-state">
            <v-icon icon="mdi-account-search-outline" size="42" />
            <p>{{ t('common.empty') }}</p>
          </div>
        </template>
      </v-data-table>
    </v-card>

    <v-dialog v-model="dialog" max-width="560" transition="dialog-bottom-transition">
      <v-card>
        <v-card-title class="text-h6 font-weight-bold px-6 pt-5">
          {{ editing ? t('common.edit') : t('user.create') }}
        </v-card-title>
        <v-card-text class="px-6">
          <v-form @submit.prevent="save">
            <v-text-field v-model="form.username" :label="t('common.name')" :disabled="Boolean(editing)" class="mb-3" />
            <v-text-field v-model="form.displayName" :label="t('common.displayName')" class="mb-3" />
            <v-text-field v-model="form.email" label="Email" prepend-inner-icon="mdi-email-outline" class="mb-3" />
            <v-text-field
              v-model="form.password"
              :label="t('user.password')"
              type="password"
              prepend-inner-icon="mdi-lock-outline"
              class="mb-3"
            />
            <v-text-field v-model="form.roles" :label="t('user.roles')" placeholder="admin,viewer" class="mb-3" />
            <v-switch v-model="form.isAdmin" :label="t('user.admin')" color="primary" inset hide-details />
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
.table-toolbar {
  padding: 12px 16px 4px;
}
.search-field {
  max-width: 320px;
}
.empty-state {
  display: grid;
  place-items: center;
  gap: 8px;
  padding: 40px 0;
  color: rgb(var(--v-theme-on-surface-variant));
}
</style>
