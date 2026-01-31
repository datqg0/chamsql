import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from '@tanstack/react-router'
import toast from 'react-hot-toast'

import { authService } from '@/services/auth.service'
import { useAuthStore } from '@/stores/use-auth-store'
import type { LoginDto, RegisterDto, User } from '@/types/auth.types'

export function useLogin() {
    const setAuth = useAuthStore((state) => state.setAuth)

    return useMutation({
        mutationFn: (dto: LoginDto) => authService.login(dto),
        onSuccess: (data) => {
            if (data.data?.accessToken) {
                const userData = data.data.user
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
                setAuth(data.data.accessToken, user, userData.role)
                toast.success('Đăng nhập thành công!')
            } else {
                toast.error(data.message || 'Đăng nhập thất bại!')
            }
        },
        onError: (error: any) => {
            toast.error(error?.response?.data?.message || error?.message || 'Đăng nhập thất bại. Vui lòng thử lại!')
        },
    })
}

export function useRegister() {
    const setAuth = useAuthStore((state) => state.setAuth)

    return useMutation({
        mutationFn: (dto: RegisterDto) => authService.register(dto),
        onSuccess: (data) => {
            // Handle success - check for accessToken regardless of code
            if (data.data?.accessToken) {
                // Auto-login sau khi đăng ký thành công
                const userData = data.data.user
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
                setAuth(data.data.accessToken, user)
                toast.success('Đăng ký thành công!')
            } else {
                toast.error(data.message || 'Đăng ký thất bại!')
            }
        },
        onError: (error: any) => {
            toast.error(error?.response?.data?.message || error?.message || 'Đăng ký thất bại. Vui lòng thử lại!')
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

