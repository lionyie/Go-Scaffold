<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'

import { menu as coreMenu } from '../router/menu'
import { currentLocale, setLocale } from '../i18n'
import { useAuthStore } from '../stores/auth'
import { usePluginStore } from '../stores/plugin'

const auth = useAuthStore()
const plugins = usePluginStore()
const router = useRouter()
const route = useRoute()
const { t } = useI18n()

const menu = computed(() => coreMenu.filter((item) => auth.hasPermission(item.permission)))
const pluginMenu = computed(() =>
  plugins.navigations.filter((item) => auth.hasPermission(item.permission)),
)
const activePath = computed(() => route.path)
const title = computed(() => {
  const key = route.meta.titleKey
  return typeof key === 'string' ? t(key) : t('app.name')
})

async function onLogout(): Promise<void> {
  await auth.logout()
  await router.replace('/login')
}

function toggleLocale(): void {
  setLocale(currentLocale() === 'zh-CN' ? 'en-US' : 'zh-CN')
}
</script>

<template>
  <el-container class="layout">
    <el-aside width="220px" class="sidebar">
      <div class="logo">{{ t('app.name') }}</div>
      <el-menu :default-active="activePath" router>
        <el-menu-item v-for="item in menu" :key="item.path" :index="item.path">
          {{ item.titleKey ? t(item.titleKey) : item.label }}
        </el-menu-item>
        <el-menu-item v-for="item in pluginMenu" :key="item.path" :index="item.path">
          {{ t(item.label) }}
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="header">
        <span>{{ title }}</span>
        <div class="header-actions">
          <el-button link @click="toggleLocale">
            {{ currentLocale() === 'zh-CN' ? 'EN' : '中文' }}
          </el-button>
          <span class="user">{{ auth.user?.displayName || auth.user?.username }}</span>
          <el-button link type="danger" @click="onLogout">{{ t('common.logout') }}</el-button>
        </div>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>
