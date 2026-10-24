<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'

interface Note {
  id?: string
  title?: string
  body?: string
}

const notes = ref<Note[]>([])
const title = ref('')
const body = ref('')
const loading = ref(false)

async function load(): Promise<void> {
  loading.value = true
  try {
    const response = await fetch('/api/v1/notes')
    if (!response.ok) throw new Error(await response.text())
    const payload = (await response.json()) as { items?: Note[] }
    notes.value = payload.items ?? []
  } catch (error) {
    ElMessage.error(String(error))
  } finally {
    loading.value = false
  }
}

async function create(): Promise<void> {
  if (!title.value.trim()) return
  try {
    const response = await fetch('/api/v1/notes', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ note: { title: title.value, body: body.value } }),
    })
    if (!response.ok) throw new Error(await response.text())
    title.value = ''
    body.value = ''
    await load()
  } catch (error) {
    ElMessage.error(String(error))
  }
}

onMounted(load)
</script>

<template>
  <div class="page-card">
    <el-card shadow="never">
      <template #header>Notes</template>
      <div class="composer">
        <el-input v-model="title" placeholder="Title" style="max-width: 240px" />
        <el-input v-model="body" placeholder="Body" style="max-width: 320px" />
        <el-button type="primary" @click="create">Add</el-button>
        <el-button @click="load">Refresh</el-button>
      </div>
      <el-table v-loading="loading" :data="notes" border style="margin-top: 16px">
        <el-table-column prop="title" label="Title" min-width="180" />
        <el-table-column prop="body" label="Body" min-width="280" />
      </el-table>
    </el-card>
  </div>
</template>

<style scoped>
.composer {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
</style>
