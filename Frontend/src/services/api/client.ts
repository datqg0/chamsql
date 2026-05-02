import axios, { AxiosError, type InternalAxiosRequestConfig } from 'axios'

export const api = axios.create({
    baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1',
    timeout: 30000,
    headers: {
        'Content-Type': 'application/json',
    },
})

// Request interceptor
api.interceptors.request.use(
    (config) => {
        const token = localStorage.getItem('auth-storage')
            ? JSON.parse(localStorage.getItem('auth-storage') || '{}').state?.token
            : null

        if (token) {
            config.headers.Authorization = `Bearer ${token}`
        }
        return config
    },
    (error) => Promise.reject(error)
)

// Response interceptor
api.interceptors.response.use(
    (response) => {
        // Auto-unwrap { code, message, data } format
        const d = response.data
        if (d && typeof d === 'object' && 'code' in d && 'data' in d) {
            response.data = d.data
        }
        return response
    },
    async (error: AxiosError) => {
        const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean }

        // Prevent infinite loop if /auth/refresh itself returns 401
        if (originalRequest.url === '/auth/refresh') {
            localStorage.removeItem('auth-storage')
            if (window.location.pathname !== '/') {
                window.location.href = '/'
            }
            return Promise.reject(error)
        }

        // Do not attempt to refresh if the error comes from login or register
        const isAuthEndpoint = originalRequest.url?.includes('/auth/login') || originalRequest.url?.includes('/auth/register');

        // Handle 401: Unauthorized -> Try to refresh token
        if (error.response?.status === 401 && !originalRequest._retry && !isAuthEndpoint) {
            originalRequest._retry = true

            try {
                // Call refresh endpoint
                const { data } = await api.post('/auth/refresh')

                // Sau khi unwrap tại response interceptor, data chính là { accessToken, user }
                if (data?.accessToken) {
                    const { accessToken, user } = data

                    // Update auth store
                    const { setAuth } = await import('@/stores/use-auth-store').then((m) =>
                        m.useAuthStore.getState()
                    )
                    setAuth(accessToken, user)

                    // Retry original request
                    originalRequest.headers.Authorization = `Bearer ${accessToken}`
                    return api(originalRequest)
                }
            } catch (refreshError) {
                // Refresh failed -> Logout
                localStorage.removeItem('auth-storage')
                if (window.location.pathname !== '/') {
                    window.location.href = '/'
                }
                return Promise.reject(refreshError)
            }
        }

        // Xử lý error response từ backend
        const responseData = error.response?.data as { message?: string } | undefined
        const errorMessage =
            responseData?.message ||
            error.message ||
            'Có lỗi xảy ra. Vui lòng thử lại!'

        return Promise.reject({
            ...error,
            message: errorMessage,
        })
    }
)
