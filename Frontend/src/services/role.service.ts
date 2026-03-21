import { api } from './api/client'

export interface RoleDTO {
    id: string
    name: string
    description: string
}

export const roleService = {
    /**
     * Get all system roles
     */
    getRoles: async (): Promise<RoleDTO[]> => {
        const { data } = await api.get<{ data: RoleDTO[] }>('/admin/roles')
        return data.data
    },
}
