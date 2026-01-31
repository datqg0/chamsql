// Basic User interface matching API response
export interface User {
    id: number
    email: string
    username: string
    fullName: string
    role: string // String role 'admin', 'student', 'lecturer'
    studentId?: string
    isActive: boolean // Active status
    createdAt?: string // Creation timestamp
    avatar?: string
    phone?: string

    // Legacy fields for backward compatibility during migration (optional)
    name?: string
    roleId?: number
    active?: number
}

// DTOs
export interface LoginDto {
    identifier: string
    password: string
}

export interface RegisterDto {
    email: string
    username: string
    password: string
    fullName: string
    studentId: string
}

// Auth Response
export interface AuthUserData {
    id: number
    email: string
    username: string
    fullName: string
    role: string
    studentId: string
}

export interface AuthResponse {
    code: number
    message: string
    data: {
        accessToken: string
        refreshToken: string
        expiresIn: number
        user: AuthUserData
    }
}

