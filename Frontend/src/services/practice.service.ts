import { api } from './api/client'

const BASE = '/student/practice'

export const practiceService = {
  async listProblems(params?: { page?: number; pageSize?: number; difficulty?: string }) {
    const { data } = await api.get(`${BASE}/problems`, { params })
    return data?.data ?? data
  },

  async getProblemById(id: number) {
    const { data } = await api.get(`${BASE}/problems/${id}`)
    return data?.data ?? data
  },

  async getProblemBySlug(slug: string) {
    const { data } = await api.get(`${BASE}/problems/slug/${slug}`)
    return data?.data ?? data
  },

  async submitCode(id: number, req: { code: string; database_type?: string }) {
    const { data } = await api.post(`${BASE}/problems/${id}/submit`, req)
    return data?.data ?? data
  },

  async listSubmissions(id: number, params?: { page?: number; pageSize?: number }) {
    const { data } = await api.get(`${BASE}/problems/${id}/submissions`, { params })
    return data?.data ?? data
  },
}
