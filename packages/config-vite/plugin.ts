import vue from '@vitejs/plugin-vue'
import * as Pinia from 'pinia'
import { defineConfig, type Plugin, type UserConfig } from 'vite'
import * as Vue from 'vue'
import * as VueI18n from 'vue-i18n'
import * as VueRouter from 'vue-router'

type HostProperty = 'vue' | 'vueRouter' | 'piniaLib' | 'vueI18n'

// The plugin build must use the host's already-loaded library instances. A
// virtual module re-exports every entry from the host global, which keeps the
// API surface complete without maintaining a hand-written shim.
function hostShim(name: HostProperty, specifier: string, lib: object): Plugin {
  const virtualId = `virtual:vela-host-${name}`
  const resolvedId = `\0${virtualId}`
  const names = Object.keys(lib).filter(
    (key) => key !== 'default' && /^[A-Za-z_$][\w$]*$/.test(key),
  )
  const code = [
    `const host = window.__SCAFFOLD_PLUGIN_HOST__;`,
    `const lib = host.${name};`,
    ...names.map((key) => `export const ${key} = lib.${key};`),
    `export default lib;`,
  ].join('\n')

  return {
    name: `vela-host-${name}`,
    enforce: 'pre',
    resolveId(id) {
      if (id === specifier) return resolvedId
      return null
    },
    load(id) {
      if (id === resolvedId) return code
      return null
    },
  }
}

export interface PluginConfigOptions {
  port?: number
  entry?: string
}

export function definePluginConfig(options: PluginConfigOptions = {}): UserConfig {
  return defineConfig({
    plugins: [
      hostShim('vue', 'vue', Vue),
      hostShim('vueRouter', 'vue-router', VueRouter),
      hostShim('piniaLib', 'pinia', Pinia),
      hostShim('vueI18n', 'vue-i18n', VueI18n),
      vue(),
    ],
    server: {
      port: options.port ?? 5174,
      cors: true,
      strictPort: true,
    },
    build: {
      lib: {
        entry: options.entry ?? 'src/index.ts',
        formats: ['es'],
        fileName: () => 'index.js',
      },
      outDir: 'dist',
      emptyOutDir: true,
      cssCodeSplit: false,
    },
  })
}
