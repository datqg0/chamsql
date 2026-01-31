import type {
    Problem,
    RunQueryRequest,
    RunQueryResponse,
    SubmitSolutionRequest,
    SubmitSolutionResponse,
} from '@/types/exam.types'

import { api } from './api/client'
import { API_ENDPOINTS } from './api/endpoints'

export interface ProblemFilters {
    page?: number
    pageSize?: number
    difficulty?: 'easy' | 'medium' | 'hard'
    topicId?: number
}

export const problemsService = {
    async list(filters?: ProblemFilters): Promise<Problem[]> {
        const { data } = await api.get(
            API_ENDPOINTS.problems.list,
            { params: filters }
        )
        // Handle API response structure: { code, message, data: { problems: [], total } }
        if (Array.isArray(data)) {
            return data
        }
        if (data && data.data && Array.isArray(data.data.problems)) {
            return data.data.problems
        }
        if (data && Array.isArray(data.data)) {
            return data.data
        }
        return []
    },

    async getBySlug(slug: string): Promise<Problem> {
        const { data } = await api.get<Problem>(API_ENDPOINTS.problems.bySlug(slug))
        return data
    },

    async create(problem: Omit<Problem, 'id'>): Promise<Problem> {
        const { data } = await api.post<Problem>(API_ENDPOINTS.problems.create, problem)
        return data
    },

    async run(problemId: number, request: RunQueryRequest): Promise<RunQueryResponse> {
        const { data } = await api.post<RunQueryResponse>(
            API_ENDPOINTS.problems.run(problemId),
            request
        )
        return data
    },

    async submit(problemId: number, request: SubmitSolutionRequest): Promise<SubmitSolutionResponse> {
        const { data } = await api.post<SubmitSolutionResponse>(
            API_ENDPOINTS.problems.submit(problemId),
            request
        )
        return data
    },
}
