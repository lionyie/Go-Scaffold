import type { Component } from 'vue'
import type { Pinia } from 'pinia'
import type { RouteRecordRaw, Router } from 'vue-router'

export interface PluginNav {
  label: string
  icon: string
  order: number
  path: string
  permission: string
}

export interface PluginRoute {
  path: string
  name: string
  component: string
  permission: string
}

export interface PluginManifest {
  id: string
  name: string
  version: string
  description: string
  icon: string
  entryUrl: string
  styles: string[]
  routes: PluginRoute[]
  navigations: PluginNav[]
  permissions: string[]
}

export interface PluginNotify {
  success(message: string): void
  error(message: string): void
  warning(message: string): void
}

export interface PluginHost {
  version: string
  vue: typeof import('vue')
  vueRouter: typeof import('vue-router')
  piniaLib: typeof import('pinia')
  vueI18n: typeof import('vue-i18n')
  router: Router
  pinia: Pinia
  i18n: { global: { t: (key: string, ...args: unknown[]) => string } }
  hasPermission(code: string): boolean
  getAccessToken(): string | null
  notify: PluginNotify
  logger: Pick<Console, 'debug' | 'info' | 'warn' | 'error'>
}

export interface PluginHttp {
  get<T = unknown>(path: string): Promise<T>
  post<T = unknown>(path: string, body?: unknown): Promise<T>
  put<T = unknown>(path: string, body?: unknown): Promise<T>
  delete<T = unknown>(path: string): Promise<T>
}

export interface PluginContext {
  manifest: PluginManifest
  host: PluginHost
  http: PluginHttp
  t(key: string, ...args: unknown[]): string
  hasPermission(code: string): boolean
  registerRoute(record: RouteRecordRaw): () => void
  notify: PluginNotify
}

export interface VelaPluginModule {
  activate?(ctx: PluginContext): void | Promise<void>
  deactivate?(): void | Promise<void>
  components?: Record<string, Component>
}

export interface PluginLoadResult {
  manifest: PluginManifest
  status: 'loaded' | 'failed'
  error?: string
}
