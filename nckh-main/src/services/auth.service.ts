import type {
    AuthResponse,
    LoginDto,
    RegisterDto,
} from '@/types/auth.types'

import { api } from './api/client'
import { API_ENDPOINTS } from './api/endpoints'

export const authService = {
    async login(dto: LoginDto): Promise<AuthResponse> {
        const { data } = await api.post<AuthResponse>(
            API_ENDPOINTS.auth.login,
            dto
        )
        return data
    },

    async register(dto: RegisterDto): Promise<AuthResponse> {
        const { data } = await api.post<AuthResponse>(
            API_ENDPOINTS.auth.register,
            dto
        )
        return data
    },

    async logout(): Promise<void> {
        // API không có endpoint logout, chỉ clear token ở client
        return Promise.resolve()
    },
}

