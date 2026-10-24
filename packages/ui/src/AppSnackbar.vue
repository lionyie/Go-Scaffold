<script setup lang="ts">
import { notifyState } from './notify'
</script>

<template>
  <v-snackbar
    v-model="notifyState.show"
    :color="notifyState.color"
    :timeout="3200"
    location="top"
    rounded="pill"
    transition="slide-y-transition"
    class="app-snackbar"
  >
    <div class="d-flex align-center ga-2">
      <v-icon :icon="notifyState.icon" size="20" />
      <span class="font-weight-medium">{{ notifyState.text }}</span>
    </div>
  </v-snackbar>
</template>

<style scoped>
.app-snackbar :deep(.v-snackbar__content) {
  padding: 10px 20px;
}
</style>
