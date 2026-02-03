import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useEffect } from 'react'

import { MainLayout } from '@/components/layouts/main-layout'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useAuthStore } from '@/stores/use-auth-store'

function DashboardPage() {
    const navigate = useNavigate()
    const { user, isOperator, isAuthenticated, userRole } = useAuthStore()

    // Redirect to login if not authenticated
    useEffect(() => {
        if (!isAuthenticated) {
            navigate({ to: '/' })
        }
    }, [isAuthenticated, navigate])

    // Show loading while checking auth
    if (!isAuthenticated) {
        return (
            <MainLayout>
                <div className="flex items-center justify-center py-12">
                    <p className="text-muted-foreground">Đang kiểm tra xác thực...</p>
                </div>
            </MainLayout>
        )
    }

    const roleName = userRole || (isOperator() ? 'Admin' : 'User')

    return (
        <MainLayout>
            <div className="space-y-6">
                <div>
                    <h1 className="text-3xl font-bold">Dashboard</h1>
                    <p className="text-muted-foreground">
                        Chào mừng trở lại, {user?.name || 'Người dùng'}!
                    </p>
                </div>

                <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                    <Card>
                        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                            <CardTitle className="text-sm font-medium">Vai trò</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold capitalize">{roleName}</div>
                            <p className="text-xs text-muted-foreground">
                                {isOperator() ? 'Full quyền truy cập' : 'Quyền hạn giới hạn'}
                            </p>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                            <CardTitle className="text-sm font-medium">Email</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold">
                                {user?.email || 'Chưa có'}
                            </div>
                            <p className="text-xs text-muted-foreground">Thông tin liên hệ</p>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                            <CardTitle className="text-sm font-medium">Trạng thái</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold">
                                {user?.active === 1 ? 'Hoạt động' : 'Khóa'}
                            </div>
                            <p className="text-xs text-muted-foreground">Tài khoản</p>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                            <CardTitle className="text-sm font-medium">ID</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold">#{user?.id || 'N/A'}</div>
                            <p className="text-xs text-muted-foreground">Mã người dùng</p>
                        </CardContent>
                    </Card>
                </div>

                <Card>
                    <CardHeader>
                        <CardTitle>Thông tin tài khoản</CardTitle>
                        <CardDescription>Chi tiết tài khoản của bạn</CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="grid gap-4 md:grid-cols-2">
                            <div>
                                <p className="text-sm font-medium text-muted-foreground">Họ và tên</p>
                                <p className="text-lg">{user?.name || 'Chưa có'}</p>
                            </div>
                            <div>
                                <p className="text-sm font-medium text-muted-foreground">Email</p>
                                <p className="text-lg">{user?.email || 'Chưa có'}</p>
                            </div>
                        </div>
                    </CardContent>
                </Card>
            </div>
        </MainLayout>
    )
}

export const Route = createFileRoute('/dashboard')({
    component: DashboardPage,
})


