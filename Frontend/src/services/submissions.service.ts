import type {
    Submission,
    SubmissionListResponse,
} from '@/types/exam.types'

import { api } from './api/client'
import { API_ENDPOINTS } from './api/endpoints'

export interface SubmissionFilters {
    page?: number
    pageSize?: number
}

export const submissionsService = {
    async list(filters?: SubmissionFilters): Promise<SubmissionListResponse> {
        const { data } = await api.get<SubmissionListResponse>(
            API_ENDPOINTS.submissions.list,
            { params: filters }
        )
        return data
    },
}
