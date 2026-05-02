import type { User } from '@/types/auth.types'
import type {
    AdminStats,
    ImportUsersRequest,
} from '@/types/exam.types'

import { api } from './api/client'
import { API_ENDPOINTS } from './api/endpoints'

export interface UserListResponse {
    data: User[]
    page: number
    pageSize: number
    total: number
}

export const adminService = {
    async getStats(): Promise<AdminStats> {
        const { data } = await api.get<AdminStats>(API_ENDPOINTS.admin.stats)
        return data
    },

    async getUsers(page = 1, pageSize = 20): Promise<UserListResponse> {
        const { data } = await api.get<UserListResponse>(
            API_ENDPOINTS.admin.users,
            { params: { page, pageSize } }
        )
        return data
    },

    async importUsers(request: ImportUsersRequest): Promise<{ success: boolean; message: string }> {
        const { data } = await api.post<{ success: boolean; message: string }>(
            API_ENDPOINTS.admin.importUsers,
            request
        )
        return data
    },

    async updateUserRole(userId: number, role: string): Promise<{ success: boolean; message: string }> {
        const { data } = await api.put<{ success: boolean; message: string }>(
            API_ENDPOINTS.admin.updateRole(userId),
            { role }
        )
        return data
    },

    async toggleUserActive(userId: number): Promise<{ message: string }> {
        const { data } = await api.patch<{ message: string }>(
            API_ENDPOINTS.admin.toggleActive(userId)
        )
        return data
    },

    async deleteUser(userId: number): Promise<void> {
        await api.delete(API_ENDPOINTS.admin.deleteUser(userId))
    },

    async getDashboard(): Promise<DashboardResponse> {
        const { data } = await api.get<DashboardResponse>(API_ENDPOINTS.admin.dashboard)
        return data
    },

    async getPerformanceTimeline(userId: number, problemId?: number): Promise<PerformanceTimelineResponse> {
        const { data } = await api.get<PerformanceTimelineResponse>(API_ENDPOINTS.admin.timeline, {
            params: { userId, problemId }
        })
        return data
    },

    async getAuditLog(page = 1, pageSize = 20): Promise<AuditLogResponse> {
        const { data } = await api.get<AuditLogResponse>(API_ENDPOINTS.admin.auditLog, {
            params: { page, pageSize }
        })
        return data
    },
}

export interface DashboardResponse {
    overview: {
        totalUsers: number
        totalProblems: number
        totalSubmissions: number
        activeUsersWeek: number
        avgSolveTimeMs: number
        usersByRole: Record<string, number>
    }
    gradingStats: {
        totalSubmissions: number
        avgGradingTimeMs: number
        minGradingTimeMs: number
        maxGradingTimeMs: number
        totalCorrect: number
        totalUsers: number
        totalProblemsAttempted: number
        passRate: number
    }
    dailySubmissions: {
        date: string
        totalSubmissions: number
        correctCount: number
        avgExecutionMs: number
    }[]
    passRates: {
        id: number
        title: string
        difficulty: string
        totalSubmissions: number
        correctCount: number
        passRate: number
    }[]
    topProblems: {
        id: number
        title: string
        slug: string
        difficulty: string
        submissionCount: number
        uniqueUsers: number
    }[]
}

export interface PerformanceTimelineResponse {
    userId: number
    problemId?: number
    timeline: {
        date: string
        avgTimeMs: number
        bestTimeMs: number
        submissionCount: number
        correctCount: number
    }[]
}

export interface AuditLogResponse {
    logs: {
        id: number
        action: string
        userId?: number
        resourceType?: string
        resourceId?: number
        oldValue?: string
        newValue?: string
        reason?: string
        ipAddress?: string
        userAgent?: string
        createdAt: string
    }[]
    total: number
    page: number
    pageSize: number
}
