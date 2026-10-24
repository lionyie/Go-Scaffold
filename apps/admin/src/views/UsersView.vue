<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'

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

async function load(): Promise<void> {
  loading.value = true
  try {
    const { data, error } = await api().GET('/api/v1/users', {
      params: { query: { pageSize: 100 } },
    })
    if (error) throw error
    rows.value = (data?.items ?? []) as UserRow[]
  } catch (error) {
    ElMessage.error(errorMessage(error, (key) => t(key)))
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
    roles: form.roles.split(',').map((item) => item.trim()).filter(Boolean),
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
    await load()
  } catch (error) {
    ElMessage.error(errorMessage(error, (key) => t(key)))
  }
}

async function remove(row: UserRow): Promise<void> {
  try {
    await ElMessageBox.confirm(t('user.deleteConfirm', { name: row.username }), t('common.confirm'))
  } catch {
    return
  }
  const { error } = await api().DELETE('/api/v1/users/{id}', {
    params: { path: { id: row.id! } },
  })
  if (error) {
    ElMessage.error(errorMessage(error, (key) => t(key)))
    return
  }
  await load()
}

onMounted(load)
</script>

<template>
  <div class="page-card">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>{{ t('user.title') }}</span>
          <div>
            <el-button @click="load">{{ t('common.refresh') }}</el-button>
            <el-button type="primary" @click="openCreate">{{ t('user.create') }}</el-button>
          </div>
        </div>
      </template>

      <el-table v-loading="loading" :data="rows" border>
        <el-table-column prop="username" :label="t('common.name')" min-width="160" />
        <el-table-column prop="displayName" :label="t('common.displayName')" min-width="160" />
        <el-table-column prop="email" label="Email" min-width="200" />
        <el-table-column :label="t('user.roles')" min-width="180">
          <template #default="{ row }">
            <el-tag v-for="role in row.roles ?? []" :key="role" style="margin-right: 4px">
              {{ role }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('user.admin')" width="110">
          <template #default="{ row }">{{ row.isAdmin ? 'yes' : 'no' }}</template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="160" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEdit(row)">{{ t('common.edit') }}</el-button>
            <el-button link type="danger" @click="remove(row)">{{ t('common.delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialog" :title="editing ? t('common.edit') : t('user.create')" width="520px">
      <el-form label-position="top">
        <el-form-item :label="t('common.name')">
          <el-input v-model="form.username" :disabled="Boolean(editing)" />
        </el-form-item>
        <el-form-item :label="t('common.displayName')">
          <el-input v-model="form.displayName" />
        </el-form-item>
        <el-form-item label="Email">
          <el-input v-model="form.email" />
        </el-form-item>
        <el-form-item :label="t('user.password')">
          <el-input v-model="form.password" type="password" />
        </el-form-item>
        <el-form-item :label="t('user.roles')">
          <el-input v-model="form.roles" placeholder="admin,viewer" />
        </el-form-item>
        <el-form-item :label="t('user.admin')">
          <el-switch v-model="form.isAdmin" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" @click="save">{{ t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
</style>
