import { createFileRoute } from '@tanstack/react-router'

import { MainLayout } from '@/components/layouts/main-layout'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { useAuthStore } from '@/stores/use-auth-store'
import { Users, GraduationCap, Shield } from 'lucide-react'

const ROLES = [
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

function RolesPage() {
    const { isOperator } = useAuthStore()

    if (!isOperator()) {
        return (
            <MainLayout>
                <div className="text-center py-12">
                    <p className="text-muted-foreground">Bạn không có quyền truy cập trang này</p>
                </div>
            </MainLayout>
        )
    }

    return (
        <MainLayout>
            <div className="space-y-6">
                <div>
                    <h1 className="text-3xl font-bold">Quản lý Vai trò</h1>
                    <p className="text-muted-foreground">Các vai trò trong hệ thống</p>
                </div>

                <div className="grid gap-4 md:grid-cols-3">
                    {ROLES.map((role) => {
                        const Icon = role.icon
                        return (
                            <Card key={role.id} className="hover:shadow-lg transition-shadow">
                                <CardHeader>
                                    <div className={`w-12 h-12 rounded-lg ${role.bgColor} flex items-center justify-center mb-4`}>
                                        <Icon className={`h-6 w-6 ${role.color}`} />
                                    </div>
                                    <CardTitle>{role.name}</CardTitle>
                                    <CardDescription>{role.description}</CardDescription>
                                </CardHeader>
                                <CardContent>
                                    <span className={`text-xs px-2 py-1 rounded ${role.bgColor} ${role.color}`}>
                                        {role.id}
                                    </span>
                                </CardContent>
                            </Card>
                        )
                    })}
                </div>
            </div>
        </MainLayout>
    )
}

export const Route = createFileRoute('/roles')({
    component: RolesPage,
})


