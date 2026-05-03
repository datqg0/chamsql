import { createFileRoute } from '@tanstack/react-router'
import { Users, GraduationCap, Shield, type LucideIcon } from 'lucide-react'

import { MainLayout } from '@/components/layouts/main-layout'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { useRoles } from '@/hooks/use-roles'
import { useAuthStore } from '@/stores/use-auth-store'

// Map role IDs to icons and colors
const ROLE_DISPLAY: Record<string, { icon: LucideIcon; color: string; bgColor: string }> = {
    student: {
        icon: Users,
        color: 'text-blue-500',
        bgColor: 'bg-blue-500/10',
    },
    lecturer: {
        icon: GraduationCap,
        color: 'text-green-500',
        bgColor: 'bg-green-500/10',
    },
    admin: {
        icon: Shield,
        color: 'text-red-500',
        bgColor: 'bg-red-500/10',
    },
}

function RolesPage() {
    const { isOperator } = useAuthStore()
    const { data: roles, isLoading, isError } = useRoles()

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

                {isLoading && (
                    <div className="text-center py-12">
                        <p className="text-muted-foreground">Đang tải...</p>
                    </div>
                )}

                {isError && (
                    <div className="text-center py-12">
                        <p className="text-destructive">Có lỗi xảy ra khi tải danh sách vai trò</p>
                    </div>
                )}

                {roles && (
                    <div className="grid gap-4 md:grid-cols-3">
                        {roles.map((role) => {
                            const display = ROLE_DISPLAY[role.name.toLowerCase()] || ROLE_DISPLAY.student
                            const Icon = display.icon
                            return (
                                <Card key={role.id} className="hover:shadow-lg transition-shadow">
                                    <CardHeader>
                                        <div className={`w-12 h-12 rounded-lg ${display.bgColor} flex items-center justify-center mb-4`}>
                                            <Icon className={`h-6 w-6 ${display.color}`} />
                                        </div>
                                        <CardTitle>{role.name}</CardTitle>
                                        <CardDescription>{role.description}</CardDescription>
                                    </CardHeader>
                                    <CardContent>
                                        <span className={`text-xs px-2 py-1 rounded ${display.bgColor} ${display.color}`}>
                                            {role.id}
                                        </span>
                                    </CardContent>
                                </Card>
                            )
                        })}
                    </div>
                )}
            </div>
        </MainLayout>
    )
}

export const Route = createFileRoute('/roles')({
    component: RolesPage,
})
