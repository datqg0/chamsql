import type { Topic } from '@/types/exam.types'

import { api } from './api/client'
import { API_ENDPOINTS } from './api/endpoints'

export const topicsService = {
    async list(): Promise<Topic[]> {
        const { data } = await api.get<{ topics: Topic[]; total: number }>(API_ENDPOINTS.topics.list)
        
        // Handle unwrapped API response: { topics: [], total }
        if (data && Array.isArray(data.topics)) {
            return data.topics
        }
        
        // Fallback for direct array or other structures
        if (Array.isArray(data)) {
            return data
        }
        
        return []
    },

    async getBySlug(slug: string): Promise<Topic> {
        const { data } = await api.get<unknown>(API_ENDPOINTS.topics.bySlug(slug))
        return data.data || data
    },

    async create(topic: Omit<Topic, 'id'>): Promise<Topic> {
        const { data } = await api.post<unknown>(API_ENDPOINTS.topics.create, topic)
        return data.data || data
    },
}
