<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'

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
    ElMessage.error(errorMessage(error, (key) => t(key)))
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
    await load()
  } catch (error) {
    ElMessage.error(errorMessage(error, (key) => t(key)))
  }
}

async function remove(row: RoleRow): Promise<void> {
  try {
    await ElMessageBox.confirm(t('role.deleteConfirm', { name: row.name }), t('common.confirm'))
  } catch {
    return
  }
  const { error } = await api().DELETE('/api/v1/roles/{id}', {
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
          <span>{{ t('role.title') }}</span>
          <div>
            <el-button @click="load">{{ t('common.refresh') }}</el-button>
            <el-button type="primary" @click="openCreate">{{ t('role.create') }}</el-button>
          </div>
        </div>
      </template>

      <el-table v-loading="loading" :data="rows" border>
        <el-table-column prop="name" :label="t('common.name')" min-width="160" />
        <el-table-column prop="displayName" :label="t('common.displayName')" min-width="160" />
        <el-table-column prop="description" :label="t('common.description')" min-width="220" />
        <el-table-column :label="t('role.permissions')" min-width="120">
          <template #default="{ row }">{{ (row.permissions ?? []).length }}</template>
        </el-table-column>
        <el-table-column :label="t('role.system')" width="110">
          <template #default="{ row }">{{ row.isSystem ? 'yes' : 'no' }}</template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="160" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEdit(row)">{{ t('common.edit') }}</el-button>
            <el-button
              link
              type="danger"
              :disabled="row.isSystem"
              @click="remove(row)"
            >{{ t('common.delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialog" :title="editing ? t('common.edit') : t('role.create')" width="620px">
      <el-form label-position="top">
        <el-form-item :label="t('common.name')">
          <el-input v-model="form.name" :disabled="Boolean(editing)" placeholder="editor" />
        </el-form-item>
        <el-form-item :label="t('common.displayName')">
          <el-input v-model="form.displayName" />
        </el-form-item>
        <el-form-item :label="t('common.description')">
          <el-input v-model="form.description" type="textarea" />
        </el-form-item>
        <el-form-item :label="t('role.permissions')">
          <el-checkbox-group v-model="form.permissions" class="permission-grid">
            <el-checkbox v-for="permission in permissions" :key="permission.code" :value="permission.code!">
              {{ permission.code }}
            </el-checkbox>
          </el-checkbox-group>
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
.permission-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 4px 12px;
}
</style>
