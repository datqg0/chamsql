import type {
    AdminStats,
    ImportUsersRequest,
} from '@/types/exam.types'
import type { User } from '@/types/auth.types'

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
}
