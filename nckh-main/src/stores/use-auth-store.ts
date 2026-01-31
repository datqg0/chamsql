import type { User } from '@/types/auth.types'

import { create } from 'zustand'
import { persist, devtools } from 'zustand/middleware'

interface AuthState {
    token: string | null
    user: User | null
    isAuthenticated: boolean
    userRole: string | null

    setAuth: (token: string, user: User, role?: string) => void
    logout: () => void
    updateUser: (user: Partial<User>) => void
    isOperator: () => boolean
}

export const useAuthStore = create<AuthState>()(
    devtools(
        persist(
            (set, get) => ({
                token: null,
                user: null,
                isAuthenticated: false,
                userRole: null,

                setAuth: (token, user, role) => {
                    // Decode JWT để lấy thông tin user từ token nếu cần
                    let roleFromToken = role
                    try {
                        const payload = JSON.parse(atob(token.split('.')[1]))
                        roleFromToken = role || payload.role || null
                    } catch {
                        // Ignore decode error
                    }

                    // Map legacy user fields if needed
                    const normalizedUser: User = {
                        ...user,
                        role: roleFromToken || user.role || 'student', // Default fallback
                        name: user.fullName || user.username || 'User'
                    }

                    set(
                        {
                            token,
                            user: normalizedUser,
                            isAuthenticated: true,
                            userRole: roleFromToken || normalizedUser.role || null,
                        },
                        false,
                        'setAuth'
                    )
                },

                logout: () =>
                    set(
                        {
                            token: null,
                            user: null,
                            isAuthenticated: false,
                            userRole: null,
                        },
                        false,
                        'logout'
                    ),

                updateUser: (userData) =>
                    set(
                        (state) => ({
                            user: state.user ? { ...state.user, ...userData } : null,
                        }),
                        false,
                        'updateUser'
                    ),

                isOperator: () => {
                    const state = get()
                    return state.userRole === 'admin' || state.userRole === 'operator'
                },
            }),
            {
                name: 'auth-storage',
                partialize: (state) => ({
                    token: state.token,
                    user: state.user,
                    userRole: state.userRole,
                }),
            }
        ),
        { name: 'AuthStore' }
    )
)
