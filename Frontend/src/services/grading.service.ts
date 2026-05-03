import type { GradingStats, GradingSubmission } from '@/types/grading.types'

export type Submission = GradingSubmission

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
        const { data } = await api.get<{ submissions: GradingSubmission[] }>(API_ENDPOINTS.lecturer.ungradedSubmissions(examId))
        return data.submissions || []
    },

    /**
     * Get grading statistics for an exam
     */
    async getGradingStats(examId: number): Promise<GradingStats> {
        const { data } = await api.get<GradingStats>(API_ENDPOINTS.lecturer.gradingStats(examId))
        return data
    },

    /**
     * View submission details for grading
     */
    async viewSubmission(submissionId: number): Promise<GradingSubmission> {
        const { data } = await api.get<GradingSubmission>(API_ENDPOINTS.lecturer.viewSubmission(submissionId))
        return data
    },

    /**
     * Grade a submission manually
     */
    async gradeSubmission(
        submissionId: number,
        score: number,
        feedback?: string
    ): Promise<GradingSubmission> {
        const { data } = await api.post<GradingSubmission>(API_ENDPOINTS.lecturer.gradeSubmission(submissionId), {
            score,
            feedback,
        })
        return data
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

        const { data } = await api.get<{ submissions: GradingSubmission[] }>(`/lecturer/submissions?${params.toString()}`)
        return data.submissions || []
    },
    /**
     * Auto grade a submission
     */
    async autoGrade(submissionId: number): Promise<GradingSubmission> {
        const { data } = await api.post<GradingSubmission>(API_ENDPOINTS.lecturer.autoGrade(submissionId))
        return data
    }
}

export default gradingService
