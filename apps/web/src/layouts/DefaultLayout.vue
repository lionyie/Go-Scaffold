<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { useTheme } from 'vuetify'

import { menu } from '../router/menu'
import { currentLocale, setLocale } from '../i18n'
import { usePluginStore } from '../stores/plugin'

const plugins = usePluginStore()
const route = useRoute()
const { t } = useI18n()
const theme = useTheme()

const isDark = computed(() => theme.global.current.value.dark)
const activePath = computed(() => route.path)
const title = computed(() => {
  const key = route.meta.titleKey
  return typeof key === 'string' ? t(key) : t('app.name')
})

const navItems = computed(() =>
  menu.map((item) => ({
    ...item,
    label: item.titleKey ? t(item.titleKey) : item.label ?? item.path,
    icon: item.path === '/' ? 'mdi-home-variant-outline' : 'mdi-puzzle-outline',
  })),
)

function toggleTheme(): void {
  const next = isDark.value ? 'scaffoldLight' : 'scaffoldDark'
  theme.global.name.value = next
  localStorage.setItem('scaffold.theme', next)
}

function toggleLocale(): void {
  setLocale(currentLocale() === 'zh-CN' ? 'en-US' : 'zh-CN')
}
</script>

<template>
  <v-navigation-drawer permanent width="248" class="web-drawer">
    <div class="brand">
      <div class="brand__mark"><v-icon icon="mdi-hexagon-multiple" size="24" /></div>
      <div class="brand__text">
        <div class="brand__name">{{ t('app.name') }}</div>
        <div class="brand__hint">{{ t('app.tagline') }}</div>
      </div>
    </div>
    <v-list nav class="px-3">
      <v-list-item
        v-for="item in navItems"
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
        v-for="item in plugins.navigations"
        :key="item.path"
        :to="item.path"
        :prepend-icon="item.icon || 'mdi-puzzle'"
        rounded="xl"
        class="nav-item"
      >
        <v-list-item-title class="font-weight-medium">{{ item.label }}</v-list-item-title>
      </v-list-item>
    </v-list>
  </v-navigation-drawer>

  <v-app-bar class="web-bar" height="74" flat>
    <v-app-bar-title>
      <span class="bar-title">{{ title }}</span>
    </v-app-bar-title>
    <v-spacer />
    <v-btn icon variant="text" @click="toggleLocale">
      <v-icon>{{ currentLocale() === 'zh-CN' ? 'mdi-translate' : 'mdi-alphabetical' }}</v-icon>
    </v-btn>
    <v-btn icon variant="text" @click="toggleTheme">
      <v-icon>{{ isDark ? 'mdi-weather-sunny' : 'mdi-weather-night' }}</v-icon>
    </v-btn>
    <v-btn color="primary" variant="flat" rounded="pill" href="/admin/" prepend-icon="mdi-shield-account">
      Admin
    </v-btn>
  </v-app-bar>

  <v-main class="web-main">
    <div class="page-shell">
      <router-view v-slot="{ Component }">
        <transition name="page" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </div>
  </v-main>
</template>

<style scoped>
.web-drawer {
  border-right: 1px solid rgba(var(--v-border-color), 0.08);
  background: rgba(var(--v-theme-surface), 0.72);
  backdrop-filter: blur(18px);
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
  width: 42px;
  height: 42px;
  border-radius: 14px;
  color: #fff;
  background: linear-gradient(135deg, rgb(var(--v-theme-primary)), rgb(var(--v-theme-secondary)));
  box-shadow: 0 14px 28px -12px rgba(var(--v-theme-primary), 0.7);
}
.brand__name {
  font-size: 1.05rem;
  font-weight: 800;
  letter-spacing: -0.02em;
}
.brand__hint {
  font-size: 0.68rem;
  color: rgb(var(--v-theme-on-surface-variant));
  letter-spacing: 0.08em;
  text-transform: uppercase;
}
.nav-item {
  margin-bottom: 4px;
  transition: transform 0.18s ease, background-color 0.18s ease;
}
.nav-item:hover {
  transform: translateX(3px);
}
.nav-item.v-list-item--active {
  background: rgba(var(--v-theme-primary), 0.12);
}
.web-bar {
  border-bottom: 1px solid rgba(var(--v-border-color), 0.08);
  background: rgba(var(--v-theme-surface), 0.72);
  backdrop-filter: blur(18px);
}
.bar-title {
  font-size: 1.05rem;
  font-weight: 700;
  letter-spacing: -0.02em;
}
.web-main {
  min-height: 100vh;
  background:
    radial-gradient(900px 460px at 85% -10%, rgba(var(--v-theme-secondary), 0.12), transparent 60%),
    radial-gradient(700px 380px at -10% 30%, rgba(var(--v-theme-primary), 0.1), transparent 55%),
    rgb(var(--v-theme-background));
}
.page-shell {
  width: min(1240px, 100%);
  margin: 0 auto;
  padding: 32px 28px 64px;
}
</style>
