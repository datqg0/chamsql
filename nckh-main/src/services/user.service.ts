import type { User } from '@/types/auth.types'

import { api } from './api/client'
import { API_ENDPOINTS } from './api/endpoints'

export interface UserListParams {
    page?: number
    pageSize?: number
}

export interface UserListResponse {
    code: number
    message: string
    data: {
        users: User[]
        total: number
        page: number
        pageSize: number
    }
}

export interface ImportUserDto {
    email: string
    username: string
    fullName: string
    studentId?: string
    role: 'student' | 'lecturer' | 'admin'
}

export interface ImportUsersDto {
    users: ImportUserDto[]
}

export const userService = {
    async getList(params?: UserListParams): Promise<UserListResponse> {
        const { data } = await api.get<UserListResponse>(API_ENDPOINTS.admin.users, {
            params,
        })
        return data
    },

    async importUsers(dto: ImportUsersDto): Promise<{ success: boolean; message: string }> {
        const { data } = await api.post<{ success: boolean; message: string }>(
            API_ENDPOINTS.admin.importUsers,
            dto
        )
        return data
    },

    async updateRole(userId: number, role: string): Promise<{ success: boolean; message: string }> {
        const { data } = await api.put<{ success: boolean; message: string }>(
            API_ENDPOINTS.admin.updateRole(userId),
            { role }
        )
        return data
    },

    async toggleActive(userId: number): Promise<{ success: boolean; message: string }> {
        const { data } = await api.put<{ success: boolean; message: string }>(
            API_ENDPOINTS.admin.toggleActive(userId)
        )
        return data
    },
}

// Re-export User type for convenience
export type { User }


