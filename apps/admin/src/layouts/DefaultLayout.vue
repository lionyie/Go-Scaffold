<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useTheme } from 'vuetify'

import { menu } from '../router/menu'
import { currentLocale, setLocale } from '../i18n'
import { useAuthStore } from '../stores/auth'
import { usePluginStore } from '../stores/plugin'
import ChangePasswordDialog from '../components/ChangePasswordDialog.vue'

const auth = useAuthStore()
const plugins = usePluginStore()
const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const theme = useTheme()

const isDark = computed(() => theme.global.current.value.dark)
const activePath = computed(() => route.path)
const title = computed(() => {
  const key = route.meta.titleKey
  return typeof key === 'string' ? t(key) : t('app.name')
})

const menuItems = computed(() =>
  menu
    .filter((item) => auth.hasPermission(item.permission))
    .map((item) => ({
      ...item,
      label: item.titleKey ? t(item.titleKey) : item.label ?? item.path,
      icon: iconFor(item.path),
    })),
)

const pluginItems = computed(() => plugins.navigations.filter((item) => auth.hasPermission(item.permission)))

function iconFor(path: string): string {
  const icons: Record<string, string> = {
    '/dashboard': 'mdi-view-dashboard-outline',
    '/users': 'mdi-account-group-outline',
    '/roles': 'mdi-shield-account-outline',
    '/permissions': 'mdi-key-chain-variant',
    '/plugins': 'mdi-puzzle-outline',
    '/profile': 'mdi-account-circle-outline',
  }
  return icons[path] ?? 'mdi-circle-small'
}

function toggleTheme(): void {
  const next = isDark.value ? 'scaffoldLight' : 'scaffoldDark'
  theme.global.name.value = next
  localStorage.setItem('scaffold.theme', next)
}

function toggleLocale(): void {
  setLocale(currentLocale() === 'zh-CN' ? 'en-US' : 'zh-CN')
}

async function onLogout(): Promise<void> {
  await auth.logout()
  await router.replace('/login')
}
</script>

<template>
  <v-navigation-drawer permanent width="264" class="admin-drawer">
    <div class="brand">
      <div class="brand__mark">
        <v-icon icon="mdi-hexagon-multiple" size="26" />
      </div>
      <div>
        <div class="brand__name">{{ t('app.name') }}</div>
        <div class="brand__hint">{{ t('app.tagline') }}</div>
      </div>
    </div>

    <v-list nav class="px-3">
      <v-list-item
        v-for="item in menuItems"
        :key="item.path"
        :to="item.path"
        :prepend-icon="item.icon"
        :active="activePath === item.path"
        rounded="xl"
        class="nav-item"
      >
        <v-list-item-title class="font-weight-medium">{{ item.label }}</v-list-item-title>
      </v-list-item>
      <v-list-item
        v-for="item in pluginItems"
        :key="item.path"
        :to="item.path"
        :prepend-icon="item.icon || 'mdi-puzzle'"
        rounded="xl"
        class="nav-item"
      >
        <v-list-item-title class="font-weight-medium">{{ item.label }}</v-list-item-title>
      </v-list-item>
    </v-list>

    <template #append>
      <div class="drawer-footer">
        <v-chip size="small" variant="tonal" color="white" prepend-icon="mdi-shield-check">
          RBAC enabled
        </v-chip>
      </div>
    </template>
  </v-navigation-drawer>

  <v-app-bar class="admin-bar" height="74" flat>
    <v-app-bar-title>
      <span class="bar-title">{{ title }}</span>
    </v-app-bar-title>
    <v-spacer />
    <v-btn icon variant="text" :aria-label="t('common.language')" @click="toggleLocale">
      <v-icon>{{ currentLocale() === 'zh-CN' ? 'mdi-translate' : 'mdi-alphabetical' }}</v-icon>
    </v-btn>
    <v-btn icon variant="text" aria-label="theme" @click="toggleTheme">
      <v-icon>{{ isDark ? 'mdi-weather-sunny' : 'mdi-weather-night' }}</v-icon>
    </v-btn>
    <v-menu location="bottom end">
      <template #activator="{ props }">
        <v-btn v-bind="props" variant="tonal" rounded="pill" class="user-chip">
          <v-avatar size="26" color="primary" class="mr-2">
            <span class="text-caption font-weight-bold">
              {{ (auth.user?.displayName || auth.user?.username || '?').slice(0, 1).toUpperCase() }}
            </span>
          </v-avatar>
          {{ auth.user?.displayName || auth.user?.username }}
          <v-icon end icon="mdi-chevron-down" size="18" />
        </v-btn>
      </template>
      <v-list density="compact" rounded="lg">
        <v-list-item prepend-icon="mdi-account-outline" :to="'/profile'" :title="t('nav.profile')" />
        <v-list-item prepend-icon="mdi-logout" :title="t('common.logout')" @click="onLogout" />
      </v-list>
    </v-menu>
  </v-app-bar>

  <v-main class="admin-main">
    <div class="page-shell">
      <router-view v-slot="{ Component }">
        <transition name="page" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </div>
  </v-main>

  <ChangePasswordDialog />
</template>

<style scoped>
.admin-drawer {
  color: #fff;
  background:
    radial-gradient(circle at 20% 0%, rgba(255, 255, 255, 0.22), transparent 42%),
    linear-gradient(165deg, #4f46e5 0%, #6366f1 42%, #0891b2 100%);
}
.brand {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 24px 22px 18px;
}
.brand__mark {
  display: grid;
  place-items: center;
  width: 44px;
  height: 44px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.18);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.35);
}
.brand__name {
  font-size: 1.05rem;
  font-weight: 800;
  letter-spacing: -0.02em;
}
.brand__hint {
  font-size: 0.72rem;
  opacity: 0.72;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}
.nav-item {
  margin-bottom: 4px;
  color: rgba(255, 255, 255, 0.82);
  transition: transform 0.18s ease, background-color 0.18s ease;
}
.nav-item:hover {
  transform: translateX(3px);
  background: rgba(255, 255, 255, 0.1);
}
.nav-item.v-list-item--active {
  color: #fff;
  background: rgba(255, 255, 255, 0.2);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.28);
}
.drawer-footer {
  padding: 16px 20px 22px;
}
.admin-bar {
  border-bottom: 1px solid rgba(var(--v-border-color), 0.08);
  background: rgba(var(--v-theme-surface), 0.72);
  backdrop-filter: blur(18px);
}
.bar-title {
  font-size: 1.05rem;
  font-weight: 700;
  letter-spacing: -0.02em;
}
.user-chip {
  text-transform: none;
}
.admin-main {
  min-height: 100vh;
  background:
    radial-gradient(900px 420px at 80% -10%, rgba(var(--v-theme-primary), 0.12), transparent 60%),
    radial-gradient(700px 380px at -10% 20%, rgba(var(--v-theme-secondary), 0.1), transparent 55%),
    rgb(var(--v-theme-background));
}
.page-shell {
  width: min(1280px, 100%);
  margin: 0 auto;
  padding: 28px 28px 56px;
}
</style>
