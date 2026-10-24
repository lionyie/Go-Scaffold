<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { notifyApi } from '@scaffold/ui'

interface Note {
  id?: string
  title?: string
  body?: string
}

const notes = ref<Note[]>([])
const title = ref('')
const body = ref('')
const loading = ref(false)
const saving = ref(false)

const headers = [
  { title: 'Title', key: 'title' },
  { title: 'Body', key: 'body' },
]

async function load(): Promise<void> {
  loading.value = true
  try {
    const response = await fetch('/api/v1/notes')
    if (!response.ok) throw new Error(await response.text())
    const payload = (await response.json()) as { items?: Note[] }
    notes.value = payload.items ?? []
  } catch (error) {
    notifyApi.error(String(error))
  } finally {
    loading.value = false
  }
}

async function create(): Promise<void> {
  if (!title.value.trim()) return
  saving.value = true
  try {
    const response = await fetch('/api/v1/notes', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ note: { title: title.value, body: body.value } }),
    })
    if (!response.ok) throw new Error(await response.text())
    title.value = ''
    body.value = ''
    notifyApi.success('Note added')
    await load()
  } catch (error) {
    notifyApi.error(String(error))
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<template>
  <v-card class="pa-2">
    <v-card-item>
      <template #title>Notes</template>
      <template #subtitle>A generated extension: proto, Go service and a page in both apps.</template>
    </v-card-item>
    <v-card-text>
      <div class="composer">
        <v-text-field v-model="title" label="Title" prepend-inner-icon="mdi-format-title" />
        <v-text-field v-model="body" label="Body" prepend-inner-icon="mdi-text" />
        <v-btn color="primary" :loading="saving" prepend-icon="mdi-plus" @click="create">Add</v-btn>
        <v-btn variant="tonal" prepend-icon="mdi-refresh" @click="load">Refresh</v-btn>
      </div>
      <v-data-table :headers="headers" :items="notes" :loading="loading" class="mt-4" hover />
    </v-card-text>
  </v-card>
</template>

<style scoped>
.composer {
  display: grid;
  grid-template-columns: minmax(160px, 1fr) minmax(220px, 2fr) auto auto;
  gap: 12px;
  align-items: center;
}
@media (max-width: 720px) {
  .composer {
    grid-template-columns: 1fr;
  }
}
</style>
