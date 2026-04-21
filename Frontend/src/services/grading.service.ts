import { api } from './api/client'
import { API_ENDPOINTS } from './api/endpoints'

export interface Submission {
    id: number
    studentID: number
    studentCode: string
    studentName: string
    examID: number
    examTitle: string
    problemID: number
    problemTitle: string
    submittedCode: string
    status: 'pending' | 'graded' | 'error'
    score: number | null
    maxScore: number
    feedback: string | null
    executionTimeMs: number | null
    submittedAt: string
    gradedAt: string | null
    gradedBy: string | null
    errorMessage: string | null
}

export interface GradingStats {
    totalSubmissions: number
    pendingCount: number
    gradedCount: number
    errorCount: number
    averageScore: number
}

export interface GradeSubmissionRequest {
    score: number
    feedback?: string
}

export interface BulkGradeRequest {
    submissionIDs: number[]
    score: number
    feedback?: string
}

export const gradingService = {
    /**
     * Get list of ungraded submissions for an exam
     */
    async listUngradedSubmissions(examId: number): Promise<Submission[]> {
        const response = await api.get(API_ENDPOINTS.lecturer.ungradedSubmissions(examId))
        return response.data?.submissions || []
    },

    /**
     * Get grading statistics for an exam
     */
    async getGradingStats(examId: number): Promise<GradingStats> {
        const response = await api.get(API_ENDPOINTS.lecturer.gradingStats(examId))
        return response.data || {
            totalSubmissions: 0,
            pendingCount: 0,
            gradedCount: 0,
            errorCount: 0,
            averageScore: 0,
        }
    },

    /**
     * View submission details for grading
     */
    async viewSubmission(submissionId: number): Promise<Submission> {
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
    ): Promise<Submission> {
        const response = await api.post(API_ENDPOINTS.lecturer.gradeSubmission(submissionId), {
            score,
            feedback,
        })
        return response.data
    },

    /**
     * Auto-grade a submission
     */
    async autoGrade(submissionId: number): Promise<Submission> {
        const response = await api.post(
            `${API_ENDPOINTS.lecturer.viewSubmission(submissionId)}/auto-grade`
        )
        return response.data
    },

    /**
     * Bulk grade multiple submissions
     */
    async bulkGrade(submissionIds: number[], score: number, feedback?: string): Promise<void> {
        await api.post(API_ENDPOINTS.lecturer.bulkGrade, {
            submissionIDs: submissionIds,
            score,
            feedback,
        })
    },

    /**
     * Get all submissions with filters
     */
    async listSubmissions(filters?: {
        examId?: number
        status?: string
        studentCode?: string
    }): Promise<Submission[]> {
        const params = new URLSearchParams()
        if (filters?.examId) params.append('exam_id', filters.examId.toString())
        if (filters?.status) params.append('status', filters.status)
        if (filters?.studentCode) params.append('student_code', filters.studentCode)

        const response = await api.get(`/lecturer/submissions?${params.toString()}`)
        return response.data?.submissions || []
    },
}

export default gradingService
