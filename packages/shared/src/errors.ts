export interface ApiErrorBody {
  code?: number
  reason?: string
  message?: string
  metadata?: Record<string, string>
}

export const ErrorReason = {
  Unauthorized: 'UNAUTHORIZED',
  Forbidden: 'FORBIDDEN',
  InvalidCredentials: 'INVALID_CREDENTIALS',
  TokenInvalid: 'TOKEN_INVALID',
  TokenExpired: 'TOKEN_EXPIRED',
  ProjectNotFound: 'PROJECT_NOT_FOUND',
  ProjectAlreadyExists: 'PROJECT_ALREADY_EXISTS',
} as const
