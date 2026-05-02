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

interface BaseResponse<T> {
    code?: number;
    message?: string;
    data?: T;
}

interface ListExamsData {
    exams: Exam[];
}

interface PdfUploadResponse {
    id?: number;
    message?: string;
}

export const examsService = {
    async list(): Promise<Exam[]> {
        const { data } = await api.get<BaseResponse<Exam[] | ListExamsData> | Exam[]>(API_ENDPOINTS.exams.list)
        
        if (Array.isArray(data)) {
            return data
        }
        if (data && data.data) {
            if (Array.isArray(data.data)) {
                return data.data
            }
            if ('exams' in data.data && Array.isArray(data.data.exams)) {
                return data.data.exams
            }
        }
        return []
    },

    async create(exam: CreateExamRequest): Promise<Exam> {
        const { data } = await api.post<BaseResponse<Exam> | Exam>(API_ENDPOINTS.exams.create, exam)
        if (data && 'data' in data && data.data) {
            return data.data
        }
        return data as Exam
    },

    async getById(examId: number): Promise<Exam> {
        const { data } = await api.get<BaseResponse<Exam> | Exam>(API_ENDPOINTS.exams.byId(examId))
        if (data && 'data' in data && data.data) {
            return data.data
        }
        return data as Exam
    },

    async addProblem(examId: number, request: AddExamProblemRequest): Promise<void> {
        await api.post(API_ENDPOINTS.exams.addProblem(examId), request)
    },

    async addParticipants(examId: number, request: AddParticipantsRequest): Promise<void> {
        await api.post(API_ENDPOINTS.exams.addParticipants(examId), request)
    },

    async listParticipants(examId: number): Promise<ExamParticipant[]> {
        const { data } = await api.get<BaseResponse<ExamParticipant[]> | ExamParticipant[]>(API_ENDPOINTS.exams.listParticipants(examId))
        if (Array.isArray(data)) {
            return data
        }
        if (data && data.data && Array.isArray(data.data)) {
            return data.data
        }
        return []
    },



    async getMyExams(): Promise<MyExam[]> {
        const { data } = await api.get<BaseResponse<MyExam[]> | MyExam[]>(API_ENDPOINTS.exams.myExams)
        if (Array.isArray(data)) {
            return data
        }
        if (data && data.data && Array.isArray(data.data)) {
            return data.data
        }
        return []
    },

    // Upload file for exam import (PDF, Excel, Doc)
    async importExamFile(file: File): Promise<{ success: boolean; examId?: number; message?: string }> {
        const formData = new FormData()
        formData.append('file', file)

        const { data } = await api.post<BaseResponse<PdfUploadResponse> | PdfUploadResponse>(
            API_ENDPOINTS.pdf.upload,
            formData,
            {
                headers: {
                    'Content-Type': 'multipart/form-data',
                },
            }
        )

        const payload = (data && 'data' in data && data.data) ? data.data : (data as PdfUploadResponse)
        
        return {
            success: Boolean(payload?.id),
            message: payload?.message,
        }
    },
}
