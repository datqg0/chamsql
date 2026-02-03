import type {
    Exam,
    CreateExamRequest,
    AddExamProblemRequest,
    AddParticipantsRequest,
    SubmitExamAnswerRequest,
    MyExam,
} from '@/types/exam.types'

import { api } from './api/client'
import { API_ENDPOINTS } from './api/endpoints'

export const examsService = {
    async list(): Promise<Exam[]> {
        const { data } = await api.get<any>(API_ENDPOINTS.exams.list)
        // Handle nested response format if exists
        if (data && data.data && Array.isArray(data.data.exams)) {
            return data.data.exams
        }
        if (data && Array.isArray(data.data)) {
            return data.data
        }
        if (Array.isArray(data)) {
            return data
        }
        return []
    },

    async create(exam: CreateExamRequest): Promise<Exam> {
        const { data } = await api.post<Exam>(API_ENDPOINTS.exams.create, exam)
        return data
    },

    async addProblem(examId: number, request: AddExamProblemRequest): Promise<void> {
        await api.post(API_ENDPOINTS.exams.addProblem(examId), request)
    },

    async addParticipants(examId: number, request: AddParticipantsRequest): Promise<void> {
        await api.post(API_ENDPOINTS.exams.addParticipants(examId), request)
    },

    async start(examId: number): Promise<{ success: boolean; message?: string }> {
        const { data } = await api.post<{ success: boolean; message?: string }>(
            API_ENDPOINTS.exams.start(examId)
        )
        return data
    },

    async submitAnswer(examId: number, request: SubmitExamAnswerRequest): Promise<{ success: boolean; status: string }> {
        const { data } = await api.post<{ success: boolean; status: string }>(
            API_ENDPOINTS.exams.submit(examId),
            request
        )
        return data
    },

    async finish(examId: number): Promise<{ success: boolean; score?: number }> {
        const { data } = await api.post<{ success: boolean; score?: number }>(
            API_ENDPOINTS.exams.finish(examId)
        )
        return data
    },

    async getMyExams(): Promise<MyExam[]> {
        const { data } = await api.get(API_ENDPOINTS.exams.myExams)
        // Handle both direct array and nested data structure
        if (Array.isArray(data)) {
            return data
        }
        if (data && Array.isArray(data.data)) {
            return data.data
        }
        return []
    },

    // Upload file for exam import (PDF, Excel, Doc)
    async importExamFile(file: File): Promise<{ success: boolean; examId?: number; message?: string }> {
        const formData = new FormData()
        formData.append('file', file)

        const { data } = await api.post<{ success: boolean; examId?: number; message?: string }>(
            `${API_ENDPOINTS.exams.create}/import`,
            formData,
            {
                headers: {
                    'Content-Type': 'multipart/form-data',
                },
            }
        )
        return data
    },
}
