<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import { PageHeader, notifyApi } from '@scaffold/ui'

import { errorMessage } from '../api/errors'
import { useAuthStore } from '../stores/auth'

interface ProviderInfo {
  name: string
  displayName: string
}

const { t } = useI18n()
const auth = useAuthStore()
const providers = ref<ProviderInfo[]>([])
const form = reactive({ oldPassword: '', newPassword: '', confirmPassword: '' })
const saving = ref(false)

const canSubmit = computed(
  () => form.oldPassword.length > 0 && form.newPassword.length >= 8 && form.newPassword === form.confirmPassword,
)

async function loadProviders(): Promise<void> {
  try {
    const response = await fetch('/api/v1/auth/oauth/providers')
    if (!response.ok) return
    const payload = (await response.json()) as { providers?: ProviderInfo[] }
    providers.value = payload.providers ?? []
  } catch {
    providers.value = []
  }
}

async function changePassword(): Promise<void> {
  if (!canSubmit.value) return
  saving.value = true
  try {
    await auth.changePassword(form.oldPassword, form.newPassword)
    form.oldPassword = ''
    form.newPassword = ''
    form.confirmPassword = ''
    notifyApi.success(t('profile.changed'))
  } catch (error) {
    notifyApi.error(errorMessage(error, (key) => t(key)))
  } finally {
    saving.value = false
  }
}

onMounted(loadProviders)
</script>

<template>
  <div class="page-card">
    <PageHeader :title="t('profile.title')" :description="auth.user?.username" />

    <v-alert
      v-if="auth.user?.mustChangePassword"
      type="warning"
      variant="tonal"
      class="mb-5"
      prominent
      icon="mdi-lock-alert"
    >
      {{ t('profile.mustChange') }}
    </v-alert>

    <v-row>
      <v-col cols="12" lg="7">
        <v-card class="pa-2">
          <v-card-item>
            <template #title>{{ t('profile.title') }}</template>
          </v-card-item>
          <v-card-text>
            <v-list lines="two">
              <v-list-item prepend-icon="mdi-account-outline" :title="t('common.name')">
                <template #subtitle>{{ auth.user?.username }}</template>
              </v-list-item>
              <v-list-item prepend-icon="mdi-badge-account-outline" :title="t('common.displayName')">
                <template #subtitle>{{ auth.user?.displayName || '—' }}</template>
              </v-list-item>
              <v-list-item prepend-icon="mdi-email-outline" title="Email">
                <template #subtitle>{{ auth.user?.email || '—' }}</template>
              </v-list-item>
              <v-list-item prepend-icon="mdi-crown-outline" :title="t('user.admin')">
                <template #subtitle>{{ auth.user?.isAdmin ? 'yes' : 'no' }}</template>
              </v-list-item>
              <v-list-item prepend-icon="mdi-shield-check-outline" :title="t('user.roles')">
                <template #subtitle>
                  <v-chip
                    v-for="role in auth.user?.roles ?? []"
                    :key="role"
                    size="small"
                    color="primary"
                    variant="tonal"
                    class="mr-1 mt-1"
                  >
                    {{ role }}
                  </v-chip>
                </template>
              </v-list-item>
            </v-list>
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" lg="5">
        <v-card>
          <v-card-item>
            <template #title>{{ t('profile.changePassword') }}</template>
          </v-card-item>
          <v-card-text>
            <v-form @submit.prevent="changePassword">
              <v-text-field
                v-model="form.oldPassword"
                :label="t('profile.oldPassword')"
                type="password"
                prepend-inner-icon="mdi-lock-outline"
                autocomplete="current-password"
                class="mb-3"
              />
              <v-text-field
                v-model="form.newPassword"
                :label="t('profile.newPassword')"
                type="password"
                prepend-inner-icon="mdi-lock-plus-outline"
                autocomplete="new-password"
                :hint="t('profile.passwordHint')"
                :persistent-hint="true"
                class="mb-4"
              />
              <v-text-field
                v-model="form.confirmPassword"
                :label="t('profile.confirmPassword')"
                type="password"
                prepend-inner-icon="mdi-lock-check-outline"
                autocomplete="new-password"
                class="mb-4"
              />
              <v-btn
                type="submit"
                color="primary"
                block
                :loading="saving"
                :disabled="!canSubmit"
              >
                {{ t('common.save') }}
              </v-btn>
            </v-form>
          </v-card-text>
        </v-card>

        <v-card v-if="providers.length > 0" class="mt-4">
          <v-card-item>
            <template #title>{{ t('oauth.linked') }}</template>
          </v-card-item>
          <v-card-text class="d-flex flex-wrap ga-2">
            <v-chip
              v-for="provider in providers"
              :key="provider.name"
              color="secondary"
              variant="tonal"
              prepend-icon="mdi-link-variant"
            >
              {{ provider.displayName }}
            </v-chip>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>
