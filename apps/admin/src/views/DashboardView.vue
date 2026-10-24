<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import { api } from '../api/client'
import { useAuthStore } from '../stores/auth'

const { t } = useI18n()
const auth = useAuthStore()
const userCount = ref<number | null>(null)
const roleCount = ref<number | null>(null)
const pluginCount = ref<number | null>(null)

const cards = [
  { key: 'users', icon: 'mdi-account-group', color: 'primary', value: userCount },
  { key: 'roles', icon: 'mdi-shield-account', color: 'secondary', value: roleCount },
  { key: 'plugins', icon: 'mdi-puzzle', color: 'accent', value: pluginCount },
]

onMounted(async () => {
  const [users, roles, plugins] = await Promise.all([
    api().GET('/api/v1/users', { params: { query: { pageSize: 1 } } }),
    api().GET('/api/v1/roles', { params: { query: { pageSize: 1 } } }),
    api().GET('/api/v1/plugins'),
  ])
  userCount.value = users.data?.totalSize ?? 0
  roleCount.value = roles.data?.totalSize ?? 0
  pluginCount.value = plugins.data?.items?.length ?? 0
})
</script>

<template>
  <div class="page-card">
    <div class="welcome glass">
      <div>
        <div class="eyebrow">Dashboard</div>
        <h2 class="welcome__title">
          Welcome back, <span class="gradient-text">{{ auth.user?.displayName || auth.user?.username }}</span>
        </h2>
        <p class="welcome__subtitle">
          Here is the current shape of your scaffold. Start a project, add an extension, or invite a
          teammate.
        </p>
      </div>
      <v-icon icon="mdi-rocket-launch-outline" size="72" class="welcome__icon" />
    </div>

    <v-row class="mt-5">
      <v-col v-for="card in cards" :key="card.key" cols="12" sm="6" lg="4">
        <v-card class="metric-card" :color="card.color" theme="dark">
          <v-card-text class="pa-6">
            <div class="d-flex align-center justify-space-between">
              <div>
                <div class="metric-label">{{ t(`nav.${card.key}`) }}</div>
                <div class="metric-value">{{ card.value.value ?? '—' }}</div>
              </div>
              <v-avatar size="52" color="white" variant="tonal">
                <v-icon :icon="card.icon" size="28" />
              </v-avatar>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-row class="mt-1">
      <v-col cols="12" lg="7">
        <v-card class="pa-2" height="100%">
          <v-card-item>
            <template #title>{{ t('profile.title') }}</template>
            <template #subtitle>{{ t('app.name') }}</template>
          </v-card-item>
          <v-card-text>
            <v-list lines="two">
              <v-list-item prepend-icon="mdi-account-outline" :title="t('common.displayName')">
                <template #subtitle>{{ auth.user?.displayName || '—' }}</template>
              </v-list-item>
              <v-list-item prepend-icon="mdi-shield-check-outline" :title="t('user.roles')">
                <template #subtitle>
                  <v-chip
                    v-for="role in auth.user?.roles ?? []"
                    :key="role"
                    size="small"
                    class="mr-1 mt-1"
                    color="primary"
                    variant="tonal"
                  >
                    {{ role }}
                  </v-chip>
                </template>
              </v-list-item>
              <v-list-item prepend-icon="mdi-crown-outline" :title="t('user.admin')">
                <template #subtitle>{{ auth.user?.isAdmin ? 'yes' : 'no' }}</template>
              </v-list-item>
            </v-list>
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="12" lg="5">
        <v-card class="pa-2" height="100%">
          <v-card-item>
            <template #title>Quick actions</template>
          </v-card-item>
          <v-card-text class="d-flex flex-column ga-3">
            <v-btn to="/users" color="primary" variant="tonal" prepend-icon="mdi-account-plus">
              {{ t('user.create') }}
            </v-btn>
            <v-btn to="/roles" color="secondary" variant="tonal" prepend-icon="mdi-shield-plus">
              {{ t('role.create') }}
            </v-btn>
            <v-btn to="/plugins" color="accent" variant="tonal" prepend-icon="mdi-puzzle-plus">
              {{ t('nav.plugins') }}
            </v-btn>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<style scoped>
.welcome {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  padding: 30px 34px;
  border-radius: 26px;
}
.eyebrow {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: rgb(var(--v-theme-primary));
}
.welcome__title {
  margin: 8px 0 8px;
  font-size: clamp(1.5rem, 2.4vw, 2.1rem);
  font-weight: 800;
  letter-spacing: -0.035em;
}
.welcome__subtitle {
  margin: 0;
  max-width: 560px;
  color: rgb(var(--v-theme-on-surface-variant));
  line-height: 1.6;
}
.welcome__icon {
  opacity: 0.16;
  transform: rotate(-8deg);
}
.metric-label {
  font-size: 0.78rem;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  opacity: 0.82;
}
.metric-value {
  margin-top: 10px;
  font-size: 2.4rem;
  font-weight: 850;
  letter-spacing: -0.04em;
}
</style>
