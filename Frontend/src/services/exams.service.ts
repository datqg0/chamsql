import type {
    Exam,
    CreateExamRequest,
    AddExamProblemRequest,
    AddParticipantsRequest,
    ExamParticipant,
    MyExam,
} from '@/types/exam.types'

import { api } from './api/client'
import { API_ENDPOINTS } from './api/endpoints'

interface PdfUploadResponse {
    id?: number;
    message?: string;
}

export const examsService = {
    async list(): Promise<Exam[]> {
        const { data } = await api.get<{ exams: Exam[] } | Exam[]>(API_ENDPOINTS.exams.list)
        
        if (Array.isArray(data)) {
            return data
        }
        if (data && 'exams' in data && Array.isArray(data.exams)) {
            return data.exams
        }
        return []
    },

    async create(exam: CreateExamRequest): Promise<Exam> {
        const { data } = await api.post<Exam>(API_ENDPOINTS.exams.create, exam)
        return data
    },

    async getById(examId: number): Promise<Exam> {
        const { data } = await api.get<Exam>(API_ENDPOINTS.exams.byId(examId))
        return data
    },

    async addProblem(examId: number, request: AddExamProblemRequest): Promise<void> {
        await api.post(API_ENDPOINTS.exams.addProblem(examId), request)
    },

    async addParticipants(examId: number, request: AddParticipantsRequest): Promise<void> {
        await api.post(API_ENDPOINTS.exams.addParticipants(examId), request)
    },

    async listParticipants(examId: number): Promise<ExamParticipant[]> {
        const { data } = await api.get<ExamParticipant[]>(API_ENDPOINTS.exams.listParticipants(examId))
        return Array.isArray(data) ? data : []
    },

    async getMyExams(): Promise<MyExam[]> {
        const { data } = await api.get<MyExam[]>(API_ENDPOINTS.exams.myExams)
        return Array.isArray(data) ? data : []
    },

    // Upload file for exam import (PDF, Excel, Doc)
    async importExamFile(file: File): Promise<{ success: boolean; examId?: number; message?: string }> {
        const formData = new FormData()
        formData.append('file', file)

        const { data } = await api.post<PdfUploadResponse>(
            API_ENDPOINTS.pdf.upload,
            formData,
            {
                headers: {
                    'Content-Type': 'multipart/form-data',
                },
            }
        )

        return {
            success: Boolean(data?.id),
            message: data?.message,
        }
    },
}
