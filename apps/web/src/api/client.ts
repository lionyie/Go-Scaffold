import { createApiClient, type ApiClient } from '@scaffold/api-client'

let client: ApiClient | null = null
let tokenProvider: () => string | null = () => null
let refreshHandler: (() => Promise<boolean>) | null = null

export function configureApi(options: {
  getToken: () => string | null
  refresh?: () => Promise<boolean>
}): void {
  tokenProvider = options.getToken
  refreshHandler = options.refresh ?? null
}

/** The authenticated client with bearer injection and single retry on 401. */
export function api(): ApiClient {
  if (client) return client
  client = createApiClient('')
  client.use({
    onRequest({ request }) {
      const token = tokenProvider()
      if (token) request.headers.set('Authorization', `Bearer ${token}`)
      return request
    },
    async onResponse({ request, response }) {
      if (response.status !== 401 || !refreshHandler) return response
      if (request.headers.get('x-vela-retried') === '1') return response
      const refreshed = await refreshHandler()
      if (!refreshed) return response
      const retry = request.clone()
      retry.headers.set('x-vela-retried', '1')
      const token = tokenProvider()
      if (token) retry.headers.set('Authorization', `Bearer ${token}`)
      return fetch(retry)
    },
  })
  return client
}

/** A client without interceptors, used for login and token refresh. */
export const rawClient = createApiClient('')
