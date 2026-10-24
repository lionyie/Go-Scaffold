import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { toApiError } from '@scaffold/api-client'

import { api, rawClient } from '../api/client'

export interface AuthUser {
  id: string
  username: string
  displayName: string
  email: string
  isAdmin: boolean
  mustChangePassword: boolean
  status: string
  roles: string[]
}

interface SessionPayload {
  accessToken?: string
  refreshToken?: string
}

const REFRESH_KEY = 'scaffold.refreshToken'

export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref('')
  const refreshToken = ref(localStorage.getItem(REFRESH_KEY) ?? '')
  const user = ref<AuthUser | null>(null)
  const permissions = ref<string[]>([])
  const initialized = ref(false)

  const isAuthenticated = computed(() => Boolean(accessToken.value && user.value))
  const isAdmin = computed(() => Boolean(user.value?.isAdmin))

  function hasPermission(code?: string | null): boolean {
    if (!code) return true
    return isAdmin.value || permissions.value.includes(code)
  }

  function applySession(payload: SessionPayload): void {
    accessToken.value = payload.accessToken ?? ''
    if (payload.refreshToken) {
      refreshToken.value = payload.refreshToken
      localStorage.setItem(REFRESH_KEY, payload.refreshToken)
    }
  }

  function clear(): void {
    accessToken.value = ''
    refreshToken.value = ''
    user.value = null
    permissions.value = []
    localStorage.removeItem(REFRESH_KEY)
  }

  async function login(username: string, password: string): Promise<void> {
    const { data, error } = await rawClient.POST('/api/v1/auth/login', {
      body: { username, password },
    })
    if (error || !data) throw toApiError(error)
    applySession(data)
    await loadMe()
  }

  async function loadMe(): Promise<void> {
    const { data, error } = await api().GET('/api/v1/auth/me')
    if (error || !data) throw toApiError(error)
    user.value = (data.user ?? null) as AuthUser | null
    permissions.value = data.permissions ?? []
  }

  let pendingRefresh: Promise<boolean> | null = null

  async function refresh(): Promise<boolean> {
    if (!refreshToken.value) return false
    if (pendingRefresh) return pendingRefresh
    pendingRefresh = (async () => {
      try {
        const { data, error } = await rawClient.POST('/api/v1/auth/refresh', {
          body: { refreshToken: refreshToken.value },
        })
        if (error || !data?.accessToken) return false
        applySession(data)
        return true
      } catch {
        return false
      } finally {
        pendingRefresh = null
      }
    })()
    return pendingRefresh
  }

  async function restore(): Promise<void> {
    if (initialized.value) return
    initialized.value = true
    if (!refreshToken.value) return
    if (!(await refresh())) {
      clear()
      return
    }
    try {
      await loadMe()
    } catch {
      clear()
    }
  }

  async function applyTokens(access: string, refresh: string): Promise<void> {
    applySession({ accessToken: access, refreshToken: refresh })
    await loadMe()
  }

  async function changePassword(oldPassword: string, newPassword: string): Promise<void> {
    const { error } = await api().POST('/api/v1/auth/password', {
      body: { oldPassword, newPassword },
    })
    if (error) throw toApiError(error)
    await loadMe()
  }

  async function logout(): Promise<void> {
    try {
      if (refreshToken.value) {
        await rawClient.POST('/api/v1/auth/logout', {
          body: { refreshToken: refreshToken.value },
        })
      }
    } catch {
      // logout is best effort
    } finally {
      clear()
    }
  }

  return {
    accessToken,
    refreshToken,
    user,
    permissions,
    initialized,
    isAuthenticated,
    isAdmin,
    hasPermission,
    login,
    loadMe,
    refresh,
    restore,
    applyTokens,
    changePassword,
    logout,
    clear,
  }
})
