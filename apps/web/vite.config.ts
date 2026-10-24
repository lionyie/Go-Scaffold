import { fileURLToPath, URL } from 'node:url'

import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'

import app from './app.json'

export default defineConfig({
  base: app.base,
  plugins: [vue()],
  resolve: {
    dedupe: ['vue', 'vue-router', 'pinia', 'vue-i18n', 'vuetify'],
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      '@scaffold/api-client': fileURLToPath(new URL('../../packages/api-client/src', import.meta.url)),
      '@scaffold/plugin-runtime': fileURLToPath(
        new URL('../../packages/plugin-runtime/src', import.meta.url),
      ),
      '@scaffold/shared': fileURLToPath(new URL('../../packages/shared/src', import.meta.url)),
      '@scaffold/ui': fileURLToPath(new URL('../../packages/ui/src', import.meta.url)),
    },
  },
  server: {
    port: app.devPort,
    proxy: Object.fromEntries(
      app.proxy.map((prefix) => [prefix, { target: 'http://127.0.0.1:8000', changeOrigin: true }]),
    ),
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
  },
})
