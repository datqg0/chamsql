/** Standard backend response wrapper { code, message, data } */
export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

/** Paginated list response */
export interface PaginatedResponse<T> {
  data: T[]
  page: number
  pageSize: number
  total: number
}

/** JSON values for dynamic data */
export type JsonPrimitive = string | number | boolean | null
export type JsonValue = JsonPrimitive | JsonValue[] | { [key: string]: JsonValue }

/** Shared enums and types */
export type SubmissionStatus = 'pending' | 'accepted' | 'wrong_answer' | 'error' | 'timeout' | 'compilation_error'
export type ExamStatus = 'not_started' | 'in_progress' | 'finished' | 'expired'
export type ProblemDifficulty = 'easy' | 'medium' | 'hard'
export type DatabaseType = 'postgresql' | 'mysql' | 'sqlserver' | 'sqlite'
export type UserRole = 'admin' | 'lecturer' | 'student'
