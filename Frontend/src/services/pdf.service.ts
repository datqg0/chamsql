import type { 
    PDFUploadResponse, 
    PDFStatusResponse, 
    ExtractedProblem, 
    ProblemSolution 
} from '@/types/pdf.types'

import { api } from './api/client'
import { API_ENDPOINTS } from './api/endpoints'

export interface UpdateSolutionResponse {
    id: number
    solution_query: string
    db_type: string
    status: string
    message: string
}

export const pdfService = {
    /**
     * Upload PDF file for problem extraction
     */
    async uploadPDF(file: File): Promise<PDFUploadResponse> {
        const formData = new FormData()
        formData.append('file', file)

        const response = await api.post(API_ENDPOINTS.pdf.upload, formData, {
            headers: {
                'Content-Type': 'multipart/form-data'
            }
        })
        return response.data
    },

    /**
     * Get upload status and extraction progress
     */
    async getUploadStatus(uploadId: number): Promise<PDFStatusResponse> {
        const response = await api.get(API_ENDPOINTS.pdf.status(uploadId))
        return response.data
    },

    /**
     * Get extracted problems from PDF
     */
    async getExtractedProblems(uploadId: number): Promise<ExtractedProblem[]> {
        const response = await api.get(API_ENDPOINTS.pdf.problems(uploadId))
        // Backend wraps with response.Success -> data.data.problems
        return response.data?.data?.problems || response.data?.problems || []
    },

    /**
     * Update solution query for a problem
     */
    async updateSolution(
        problemId: number,
        solution: ProblemSolution
    ): Promise<UpdateSolutionResponse> {
        const response = await api.put(
            API_ENDPOINTS.pdf.updateSolution(problemId),
            {
                solution_query: solution.solution_query,
                db_type: solution.db_type
            }
        )
        return response.data
    },

    /**
     * Confirm problem extraction
     */
    async confirmProblem(
        problemId: number,
        solution: ProblemSolution
    ): Promise<{ id: number; status: string; message: string }> {
        const response = await api.post(
            API_ENDPOINTS.pdf.confirmProblem(problemId),
            { solution_query: solution.solution_query, db_type: solution.db_type }
        )
        return response.data?.data ?? response.data
    },

    /**
     * Poll for extraction completion
     */
    async pollForCompletion(
        uploadId: number,
        onProgress?: (status: string) => void,
        maxAttempts = 30,
        intervalMs = 2000
    ): Promise<ExtractedProblem[]> {
        let attempts = 0

        while (attempts < maxAttempts) {
            await new Promise(resolve => setTimeout(resolve, intervalMs))
            const status = await this.getUploadStatus(uploadId)

            onProgress?.(status.status)

            if (status.status === 'completed') {
                return this.getExtractedProblems(uploadId)
            } else if (status.status === 'failed') {
                throw new Error(status.error_message || 'Extraction failed')
            }

            attempts++
        }

        throw new Error('Polling timeout')
    }
}

export default pdfService
