import type { Problem, PaginatedResponse, SubmissionStatus } from '@/types'
import type { DatabaseType, ProblemDifficulty } from '@/types/api.types'

import { api } from './api/client'

const BASE = '/student/practice'

export interface PracticeSubmitRequest {
  code: string
  database_type?: DatabaseType
}

export interface PracticeSubmitResult {
  submissionId: number
  problemId: number
  status: SubmissionStatus
  isCorrect: boolean
  executionTimeMs?: number
  errorMessage?: string
  submittedAt: string
}

export const practiceService = {
  async listProblems(params?: {
    page?: number
    pageSize?: number
    difficulty?: ProblemDifficulty
  }): Promise<PaginatedResponse<Problem>> {
    const { data } = await api.get<PaginatedResponse<Problem>>(
      `${BASE}/problems`,
      { params }
    )
    return data
  },

  async getProblemById(id: number): Promise<Problem> {
    const { data } = await api.get<Problem>(`${BASE}/problems/${id}`)
    return data
  },

  async getProblemBySlug(slug: string): Promise<Problem> {
    const { data } = await api.get<Problem>(`${BASE}/problems/slug/${slug}`)
    return data
  },

  async submitCode(
    id: number,
    req: PracticeSubmitRequest
  ): Promise<PracticeSubmitResult> {
    const { data } = await api.post<PracticeSubmitResult>(
      `${BASE}/problems/${id}/submit`,
      req
    )
    return data
  },

  async listSubmissions(
    id: number,
    params?: { page?: number; pageSize?: number }
  ): Promise<PaginatedResponse<unknown>> {
    const { data } = await api.get<PaginatedResponse<unknown>>(
      `${BASE}/problems/${id}/submissions`,
      { params }
    )
    return data
  },
}
