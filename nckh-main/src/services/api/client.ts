import axios, { AxiosError } from 'axios'

export const api = axios.create({
    baseURL: 'http://localhost:8080/api/v1',
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
        // Nếu response có format { success: true, data: ... }
        if (response.data?.success !== undefined) {
            return response
        }
        // Nếu response trực tiếp là data
        return response
    },
    async (error: AxiosError) => {
        if (error.response?.status === 401) {
            // Clear auth storage
            localStorage.removeItem('auth-storage')
            // Redirect to login
            if (window.location.pathname !== '/') {
                window.location.href = '/'
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
