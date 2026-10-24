import { ApiError, toApiError } from '@scaffold/api-client'

export { ApiError, toApiError }

export function errorMessage(error: unknown, t: (key: string) => string): string {
  const apiError = toApiError(error)
  const key = `errors.${apiError.reason}`
  const translated = t(key)
  return translated === key ? apiError.message : translated
}
