import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from '@tanstack/react-router'
import toast from 'react-hot-toast'

import { extractErrorMessage } from '@/lib/errors'
import { authService } from '@/services/auth.service'
import { useAuthStore } from '@/stores/use-auth-store'
import type { LoginDto, RegisterDto, User } from '@/types/auth.types'


export function useLogin() {
    const setAuth = useAuthStore((state) => state.setAuth)

    return useMutation({
        mutationFn: (dto: LoginDto) => authService.login(dto),
        onSuccess: (data) => {
            if (data.accessToken) {
                const userData = data.user
                const user: User = {
                    id: userData.id,
                    fullName: userData.fullName,
                    username: userData.username,
                    email: userData.email,
                    role: userData.role,
                    studentId: userData.studentId,
                    isActive: true, // Default to true on login
                    // Backward compatibility
                    name: userData.fullName,
                }
                setAuth(data.accessToken, user, userData.role)
                toast.success('Đăng nhập thành công!')
            } else {
                toast.error('Đăng nhập thất bại!')
            }
        },
        onError: (error: unknown) => {
            toast.error(extractErrorMessage(error, 'Đăng nhập thất bại. Vui lòng thử lại!'))
        },
    })
}

export function useRegister() {
    const setAuth = useAuthStore((state) => state.setAuth)

    return useMutation({
        mutationFn: (dto: RegisterDto) => authService.register(dto),
        onSuccess: (data) => {
            // Handle success - check for accessToken regardless of code
            if (data.accessToken) {
                // Auto-login sau khi đăng ký thành công
                const userData = data.user
                const user: User = {
                    id: userData.id,
                    fullName: userData.fullName,
                    username: userData.username,
                    email: userData.email,
                    role: userData.role || 'student', // Default role
                    studentId: userData.studentId,
                    isActive: true, // Default to true on register
                    name: userData.fullName,
                }
                setAuth(data.accessToken, user)
                toast.success('Đăng ký thành công!')
            } else {
                toast.error('Đăng ký thất bại!')
            }
        },
        onError: (error: unknown) => {
            toast.error(extractErrorMessage(error, 'Đăng ký thất bại. Vui lòng thử lại!'))
        },
    })
}

export function useLogout() {
    const logout = useAuthStore((state) => state.logout)
    const queryClient = useQueryClient()
    const navigate = useNavigate()

    return useMutation({
        mutationFn: () => authService.logout(),
        onSuccess: () => {
            logout()
            queryClient.clear()
            navigate({ to: '/' })
        },
        onError: () => {
            // Vẫn logout ngay cả khi API fail
            logout()
            queryClient.clear()
            navigate({ to: '/' })
        },
    })
}

