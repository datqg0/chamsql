import type { GradingStats, GradingSubmission } from '@/types/grading.types'

import { api } from './api/client'
import { API_ENDPOINTS } from './api/endpoints'

export interface GradeSubmissionRequest {
    score: number
    feedback?: string
}

export const gradingService = {
    /**
     * Get list of ungraded submissions for an exam
     */
    async listUngradedSubmissions(examId: number): Promise<GradingSubmission[]> {
        const response = await api.get(API_ENDPOINTS.lecturer.ungradedSubmissions(examId))
        return response.data?.submissions || []
    },

    /**
     * Get grading statistics for an exam
     */
    async getGradingStats(examId: number): Promise<GradingStats> {
        const response = await api.get(API_ENDPOINTS.lecturer.gradingStats(examId))
        return response.data
    },

    /**
     * View submission details for grading
     */
    async viewSubmission(submissionId: number): Promise<GradingSubmission> {
        const response = await api.get(API_ENDPOINTS.lecturer.viewSubmission(submissionId))
        return response.data
    },

    /**
     * Grade a submission manually
     */
    async gradeSubmission(
        submissionId: number,
        score: number,
        feedback?: string
    ): Promise<GradingSubmission> {
        const response = await api.post(API_ENDPOINTS.lecturer.gradeSubmission(submissionId), {
            score,
            feedback,
        })
        return response.data
    },

    /**
     * Get all submissions with filters
     */
    async listSubmissions(filters?: {
        examId?: number
        status?: string
        studentCode?: string
    }): Promise<GradingSubmission[]> {
        const params = new URLSearchParams()
        if (filters?.examId) params.append('exam_id', filters.examId.toString())
        if (filters?.status) params.append('status', filters.status)
        if (filters?.studentCode) params.append('student_code', filters.studentCode)

        const response = await api.get(`/submissions?${params.toString()}`)
        return response.data?.submissions || []
    },
}

export default gradingService
