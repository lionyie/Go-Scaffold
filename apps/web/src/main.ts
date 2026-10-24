import { createApp } from 'vue'
import * as Vue from 'vue'
import ElementPlus, { ElMessage } from 'element-plus'
import * as ElementPlusLib from 'element-plus'
import 'element-plus/dist/index.css'
import { createPinia } from 'pinia'
import * as PiniaLib from 'pinia'
import * as VueI18n from 'vue-i18n'
import * as VueRouter from 'vue-router'

import { installPluginHost } from '@scaffold/plugin-runtime'

import App from './App.vue'
import { configureApi } from './api/client'
import { i18n } from './i18n'
import router from './router'
import './styles/index.css'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(i18n)
app.use(ElementPlus)

configureApi({ getToken: () => null })

installPluginHost({
  version: '0.1.0',
  vue: Vue,
  vueRouter: VueRouter,
  piniaLib: PiniaLib,
  vueI18n: VueI18n,
  elementPlus: ElementPlusLib,
  router,
  pinia,
  i18n: i18n as unknown as { global: { t: (key: string, ...args: unknown[]) => string } },
  hasPermission: () => true,
  getAccessToken: () => null,
  notify: {
    success: (message) => ElMessage.success(message),
    error: (message) => ElMessage.error(message),
    warning: (message) => ElMessage.warning(message),
  },
  logger: console,
})

app.mount('#app')

await router.isReady()
