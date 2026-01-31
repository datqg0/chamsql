import type { Topic } from '@/types/exam.types'

import { api } from './api/client'
import { API_ENDPOINTS } from './api/endpoints'

export const topicsService = {
    async list(): Promise<Topic[]> {
        const { data } = await api.get(API_ENDPOINTS.topics.list)
        // Handle API response structure: { code, message, data: { topics: [], total } }
        if (Array.isArray(data)) {
            return data
        }
        if (data && data.data && Array.isArray(data.data.topics)) {
            return data.data.topics
        }
        if (data && Array.isArray(data.data)) {
            return data.data
        }
        return []
    },

    async getBySlug(slug: string): Promise<Topic> {
        const { data } = await api.get<Topic>(API_ENDPOINTS.topics.bySlug(slug))
        return data
    },

    async create(topic: Omit<Topic, 'id'>): Promise<Topic> {
        const { data } = await api.post<Topic>(API_ENDPOINTS.topics.create, topic)
        return data
    },
}
