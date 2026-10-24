import createClient from 'openapi-fetch'

import type { paths } from './generated/schema'

export type ApiClient = ReturnType<typeof createClient<paths>>
export type { paths }

export function createApiClient(baseUrl = '', fetchImpl?: typeof fetch): ApiClient {
  return createClient<paths>({
    baseUrl,
    ...(fetchImpl ? { fetch: fetchImpl } : {}),
  })
}
