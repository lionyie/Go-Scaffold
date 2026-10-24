import { createApp } from 'vue'
import * as Vue from 'vue'
import { createPinia } from 'pinia'
import * as PiniaLib from 'pinia'
import * as VueI18n from 'vue-i18n'
import * as VueRouter from 'vue-router'
import * as VuetifyLib from 'vuetify'
import 'vuetify/styles'
import '@mdi/font/css/materialdesignicons.css'

import { AppSnackbar, ConfirmDialog, createScaffoldVuetify, notifyApi } from '@scaffold/ui'
import { installPluginHost } from '@scaffold/plugin-runtime'

import App from './App.vue'
import { configureApi } from './api/client'
import { i18n } from './i18n'
import router from './router'
import './styles/index.css'

const savedTheme = localStorage.getItem('scaffold.theme')
const vuetify = createScaffoldVuetify(savedTheme === 'scaffoldDark' ? 'scaffoldDark' : 'scaffoldLight')

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(i18n)
app.use(vuetify)
app.component('AppSnackbar', AppSnackbar)
app.component('ConfirmDialog', ConfirmDialog)

configureApi({ getToken: () => null })

installPluginHost({
  version: '0.1.0',
  vue: Vue,
  vueRouter: VueRouter,
  piniaLib: PiniaLib,
  vueI18n: VueI18n,
  vuetifyLib: VuetifyLib,
  router,
  pinia,
  i18n: i18n as unknown as { global: { t: (key: string, ...args: unknown[]) => string } },
  hasPermission: () => true,
  getAccessToken: () => null,
  notify: notifyApi,
  logger: console,
})

app.mount('#app')

await router.isReady()
