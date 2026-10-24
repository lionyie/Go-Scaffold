<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'

import { notifyApi } from '@scaffold/ui'

import { errorMessage } from '../api/errors'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const { t } = useI18n()

const visible = computed(() => Boolean(auth.user?.mustChangePassword))
const form = reactive({ oldPassword: '', newPassword: '', confirmPassword: '' })
const saving = ref(false)
const error = ref('')

async function submit(): Promise<void> {
  error.value = ''
  if (form.newPassword.length < 8) {
    error.value = t('profile.newPasswordTooShort')
    return
  }
  if (form.newPassword !== form.confirmPassword) {
    error.value = t('profile.passwordMismatch')
    return
  }
  saving.value = true
  try {
    await auth.changePassword(form.oldPassword, form.newPassword)
    form.oldPassword = ''
    form.newPassword = ''
    form.confirmPassword = ''
    notifyApi.success(t('profile.changed'))
    await router.replace('/dashboard')
  } catch (err) {
    error.value = errorMessage(err, (key) => t(key))
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <v-dialog :model-value="visible" persistent max-width="480" transition="dialog-bottom-transition">
    <v-card class="password-card">
      <div class="password-card__glow" />
      <div class="password-card__icon">
        <v-icon icon="mdi-lock-reset" size="30" />
      </div>
      <v-card-title class="text-h5 font-weight-bold px-8 pt-5">
        {{ t('profile.mustChangeTitle') }}
      </v-card-title>
      <v-card-text class="px-8 pt-1 text-medium-emphasis">
        {{ t('profile.mustChange') }}
      </v-card-text>

      <v-form class="px-8" @submit.prevent="submit">
        <v-text-field
          v-model="form.oldPassword"
          :label="t('profile.oldPassword')"
          prepend-inner-icon="mdi-lock-outline"
          type="password"
          autocomplete="current-password"
          class="mb-3"
        />
        <v-text-field
          v-model="form.newPassword"
          :label="t('profile.newPassword')"
          prepend-inner-icon="mdi-lock-plus-outline"
          type="password"
          autocomplete="new-password"
          :hint="t('profile.passwordHint')"
          :persistent-hint="true"
          class="mb-5"
        />
        <v-text-field
          v-model="form.confirmPassword"
          :label="t('profile.confirmPassword')"
          prepend-inner-icon="mdi-lock-check-outline"
          type="password"
          autocomplete="new-password"
          class="mb-2"
        />
        <v-alert v-if="error" type="error" variant="tonal" density="compact" class="mb-4">
          {{ error }}
        </v-alert>
        <v-btn type="submit" color="primary" size="large" block :loading="saving" class="mb-2">
          {{ t('profile.changePassword') }}
        </v-btn>
      </v-form>
      <div class="pb-6" />
    </v-card>
  </v-dialog>
</template>

<style scoped>
.password-card {
  position: relative;
  overflow: hidden;
  border: 1px solid rgba(var(--v-theme-primary), 0.18);
}
.password-card__glow {
  position: absolute;
  inset: -45% -10% auto;
  height: 220px;
  background: radial-gradient(circle, rgba(var(--v-theme-primary), 0.24), transparent 68%);
  pointer-events: none;
}
.password-card__icon {
  display: grid;
  place-items: center;
  width: 58px;
  height: 58px;
  margin: 28px 32px 0;
  border-radius: 18px;
  color: #fff;
  background: linear-gradient(135deg, rgb(var(--v-theme-primary)), rgb(var(--v-theme-secondary)));
  box-shadow: 0 16px 32px -12px rgba(var(--v-theme-primary), 0.65);
}
</style>
