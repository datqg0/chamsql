import type { DashboardResponse, SystemStats } from '@/types/admin.types'
import { api } from './api/client'
import { API_ENDPOINTS } from './api/endpoints'

export const adminService = {
    /**
     * Get comprehensive dashboard analytics
     */
    async getDashboard(): Promise<DashboardResponse> {
        const { data } = await api.get<DashboardResponse>(API_ENDPOINTS.admin.dashboard)
        return data
    },

    /**
     * Get system statistics
     */
    async getSystemStats(): Promise<SystemStats> {
        const { data } = await api.get<SystemStats>(API_ENDPOINTS.admin.stats)
        return data
    },
}

export default adminService
