export interface ApiErrorBody {
  code?: number
  reason?: string
  message?: string
  metadata?: Record<string, string>
}

export class ApiError extends Error {
  readonly code: number
  readonly reason: string
  readonly metadata?: Record<string, string>

  constructor(body: ApiErrorBody) {
    super(body.message || body.reason || 'request failed')
    this.name = 'ApiError'
    this.code = body.code ?? 0
    this.reason = body.reason ?? 'UNKNOWN'
    this.metadata = body.metadata
  }
}

export function toApiError(error: unknown): ApiError {
  if (error instanceof ApiError) return error
  if (error && typeof error === 'object') {
    const body = error as ApiErrorBody
    if (body.reason || body.message) return new ApiError(body)
  }
  return new ApiError({ code: 0, reason: 'NETWORK_ERROR', message: String(error) })
}
