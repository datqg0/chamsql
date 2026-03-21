import { useQuery } from '@tanstack/react-query'

import { roleService } from '@/services/role.service'

export const useRoles = () => {
    return useQuery({
        queryKey: ['roles'],
        queryFn: () => roleService.getRoles(),
        staleTime: 5 * 60 * 1000, // 5 minutes - roles rarely change
    })
}
