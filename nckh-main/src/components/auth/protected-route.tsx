import { Navigate } from '@tanstack/react-router'

import { useAuthStore } from '@/stores/use-auth-store'

interface ProtectedRouteProps {
    children: React.ReactNode
    requiredRoles?: string[]
    requireOperator?: boolean
}

export function ProtectedRoute({
    children,
    requiredRoles,
    requireOperator = false,
}: ProtectedRouteProps) {
    const { isAuthenticated, userRole, isOperator } = useAuthStore()

    if (!isAuthenticated) {
        return <Navigate to="/" />
    }

    if (requireOperator && !isOperator()) {
        return <Navigate to="/dashboard" />
    }

    if (requiredRoles && userRole && !requiredRoles.includes(userRole)) {
        // Allow operator to access everything
        if (isOperator()) {
            return <>{children}</>
        }
        return <Navigate to="/dashboard" />
    }

    return <>{children}</>
}

