<script setup lang="ts">
import { confirmState, settleConfirm } from './confirm'
</script>

<template>
  <v-dialog v-model="confirmState.show" max-width="440" persistent transition="dialog-bottom-transition">
    <v-card class="confirm-card">
      <div class="confirm-card__glow" />
      <v-card-title class="text-h6 font-weight-bold pt-6 px-6">
        {{ confirmState.title }}
      </v-card-title>
      <v-card-text class="text-medium-emphasis px-6 pt-2">
        {{ confirmState.message }}
      </v-card-text>
      <v-card-actions class="px-6 pb-5">
        <v-spacer />
        <v-btn variant="text" @click="settleConfirm(false)">{{ confirmState.cancelText }}</v-btn>
        <v-btn color="error" variant="flat" @click="settleConfirm(true)">
          {{ confirmState.confirmText }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<style scoped>
.confirm-card {
  position: relative;
  overflow: hidden;
}
.confirm-card__glow {
  position: absolute;
  inset: -40% -10% auto;
  height: 160px;
  background: radial-gradient(circle, rgba(239, 68, 68, 0.22), transparent 70%);
  pointer-events: none;
}
</style>
