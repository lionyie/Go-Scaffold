import type { PluginHttp } from './types'

export function createPluginHttp(baseUrl: string, getToken: () => string | null): PluginHttp {
  async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
    const headers = new Headers(init.headers)
    const token = getToken()
    if (token) headers.set('Authorization', `Bearer ${token}`)
    if (init.body && !headers.has('Content-Type')) {
      headers.set('Content-Type', 'application/json')
    }
    const response = await fetch(`${baseUrl}${path}`, { ...init, headers })
    const text = await response.text()
    const data = text ? JSON.parse(text) : undefined
    if (!response.ok) {
      throw new Error(data?.message || `plugin request failed: ${response.status}`)
    }
    return data as T
  }

  return {
    get: <T>(path: string) => request<T>(path),
    post: <T>(path: string, body?: unknown) =>
      request<T>(path, { method: 'POST', body: body === undefined ? undefined : JSON.stringify(body) }),
    put: <T>(path: string, body?: unknown) =>
      request<T>(path, { method: 'PUT', body: body === undefined ? undefined : JSON.stringify(body) }),
    delete: <T>(path: string) => request<T>(path, { method: 'DELETE' }),
  }
}
