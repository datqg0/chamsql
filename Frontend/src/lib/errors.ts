import type { AxiosError } from 'axios'

interface BackendErrorBody {
  message?: string
  code?: number
}

export function isAxiosError(error: unknown): error is AxiosError<BackendErrorBody> {
  return (
    typeof error === 'object' &&
    error !== null &&
    'isAxiosError' in error &&
    (error as Record<string, unknown>).isAxiosError === true
  ) || (
    typeof error === 'object' &&
    error !== null &&
    'name' in error &&
    (error as Record<string, unknown>).name === 'AxiosError'
  )
}

export function extractErrorMessage(error: unknown, fallback = 'Có lỗi xảy ra'): string {
  if (isAxiosError(error)) {
    return error.response?.data?.message ?? (error as Record<string, unknown>).message as string ?? fallback
  }
  if (error instanceof Error) {
    return error.message
  }
  if (typeof error === 'string') {
      return error
  }
  return fallback
}
