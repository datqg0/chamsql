import type {
    SubmissionListResponse,
} from '@/types/exam.types'

import { api } from './api/client'
import { API_ENDPOINTS } from './api/endpoints'

export interface SubmissionFilters {
    page?: number
    pageSize?: number
}

const mapExamResultToSubmission = (item: any): SubmissionListResponse['data'][number] => ({
    id: Number(item.exam_id ?? item.examId ?? 0),
    problemId: 0,
    userId: Number(item.student_id ?? item.studentId ?? 0),
    examId: Number(item.exam_id ?? item.examId ?? 0),
    examTitle: item.title ?? '',
    code: '',
    databaseType: 'postgresql',
    status: (item.status ?? 'accepted') as any,
    isCorrect: Number(item.total_score ?? item.totalScore ?? 0) > 0,
    score: Number(item.total_score ?? item.totalScore ?? 0),
    submittedAt: item.submitted_at ?? item.submittedAt ?? '',
    createdAt: item.submitted_at ?? item.submittedAt ?? '',
})

export const submissionsService = {
    async list(filters?: SubmissionFilters): Promise<SubmissionListResponse> {
        // Online exam history should come from student results endpoint.
        const { data } = await api.get<any>(
            API_ENDPOINTS.student.results,
            {
                params: {
                    page: filters?.page,
                    limit: filters?.pageSize,
                },
            }
        )
        const payload = data?.data ?? data
        const results = Array.isArray(payload?.results) ? payload.results : []

        return {
            data: results.map(mapExamResultToSubmission),
            page: Number(payload?.page ?? filters?.page ?? 1),
            pageSize: Number(payload?.limit ?? filters?.pageSize ?? 20),
            total: Number(payload?.total ?? results.length),
        }
    },
}
