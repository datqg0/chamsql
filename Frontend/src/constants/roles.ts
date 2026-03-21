import { Users, GraduationCap, Shield } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'

export interface Role {
    id: string
    name: string
    description: string
    icon: LucideIcon
    color: string
    bgColor: string
}

export const SYSTEM_ROLES: Role[] = [
    {
        id: 'student',
        name: 'Student',
        description: 'Sinh viên - Có thể làm bài tập, tham gia kỳ thi',
        icon: Users,
        color: 'text-blue-500',
        bgColor: 'bg-blue-500/10',
    },
    {
        id: 'lecturer',
        name: 'Lecturer',
        description: 'Giảng viên - Tạo bài tập, tạo kỳ thi, chấm điểm',
        icon: GraduationCap,
        color: 'text-green-500',
        bgColor: 'bg-green-500/10',
    },
    {
        id: 'admin',
        name: 'Admin',
        description: 'Quản trị viên - Full quyền quản lý hệ thống',
        icon: Shield,
        color: 'text-red-500',
        bgColor: 'bg-red-500/10',
    },
]
