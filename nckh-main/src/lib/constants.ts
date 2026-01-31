export const APP_NAME = 'React Modern Stack'
export const APP_VERSION = '1.0.0'

export const ROUTES = {
    HOME: '/',
    ABOUT: '/about',
    LOGIN: '/login',
    REGISTER: '/register',
    DASHBOARD: '/dashboard',
} as const

export const API_CONFIG = {
    TIMEOUT: 10000,
    RETRY_COUNT: 1,
    STALE_TIME: 60 * 1000, // 1 minute
} as const

export const STORAGE_KEYS = {
    AUTH_TOKEN: 'auth-token',
    THEME: 'vite-ui-theme',
    AUTH_STORAGE: 'auth-storage',
} as const
