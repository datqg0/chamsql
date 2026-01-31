import { createFileRoute } from '@tanstack/react-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import toast from 'react-hot-toast'
import { Trash2, RotateCcw } from 'lucide-react'

import { MainLayout } from '@/components/layouts/main-layout'
import { AddUserDialog } from '@/components/user/add-user-dialog'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { userService, type User } from '@/services/user.service'
import { useAuthStore } from '@/stores/use-auth-store'

function UsersPage() {
    const { isOperator } = useAuthStore()
    const queryClient = useQueryClient()
    const [page] = useState(1)

    const { data, isLoading, error } = useQuery({
        queryKey: ['users', page],
        queryFn: () => userService.getList({ page, pageSize: 20 }),
    })

    const updateRoleMutation = useMutation({
        mutationFn: ({ userId, role }: { userId: number; role: string }) =>
            userService.updateRole(userId, role),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['users'] })
            toast.success('Cập nhật role thành công!')
        },
        onError: () => {
            toast.error('Cập nhật role thất bại!')
        },
    })

    const toggleActiveMutation = useMutation({
        mutationFn: (userId: number) => userService.toggleActive(userId),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['users'] })
            toast.success('Cập nhật trạng thái thành công!')
        },
        onError: () => {
            toast.error('Cập nhật trạng thái thất bại!')
        },
    })

    if (!isOperator()) {
        return (
            <MainLayout>
                <div className="text-center py-12">
                    <p className="text-muted-foreground">Bạn không có quyền truy cập trang này</p>
                </div>
            </MainLayout>
        )
    }

    const handleRoleChange = (userId: number, newRole: string) => {
        updateRoleMutation.mutate({ userId, role: newRole })
    }

    const handleToggleActive = (userId: number) => {
        toggleActiveMutation.mutate(userId)
    }

    return (
        <MainLayout>
            <div className="space-y-6">
                <div className="flex items-center justify-between">
                    <div>
                        <h1 className="text-3xl font-bold">Quản lý Người dùng</h1>
                        <p className="text-muted-foreground">Danh sách tất cả người dùng trong hệ thống</p>
                    </div>
                    <AddUserDialog
                        onSuccess={() => {
                            queryClient.invalidateQueries({ queryKey: ['users'] })
                        }}
                    />
                </div>

                <Card>
                    <CardHeader>
                        <CardTitle>Danh sách người dùng</CardTitle>
                    </CardHeader>
                    <CardContent>
                        {isLoading && <p>Đang tải...</p>}
                        {error && <p className="text-destructive">Có lỗi xảy ra</p>}
                        {data && data.data && data.data.users && (
                            <div className="overflow-x-auto">
                                {data.data.users.length === 0 ? (
                                    <p className="text-center py-8 text-muted-foreground">
                                        Không tìm thấy người dùng nào
                                    </p>
                                ) : (
                                    <table className="w-full border-collapse">
                                        <thead>
                                            <tr className="border-b">
                                                <th className="text-left p-3 font-semibold">STT</th>
                                                <th className="text-left p-3 font-semibold">Email</th>
                                                <th className="text-left p-3 font-semibold">Username</th>
                                                <th className="text-left p-3 font-semibold">Họ Tên</th>
                                                <th className="text-left p-3 font-semibold">MSSV</th>
                                                <th className="text-left p-3 font-semibold">Role</th>
                                                <th className="text-left p-3 font-semibold">Trạng thái</th>
                                                <th className="text-left p-3 font-semibold">Ngày tạo</th>
                                                <th className="text-left p-3 font-semibold">Thao tác</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {data.data.users.map((user: User, index: number) => (
                                                <tr
                                                    key={user.id}
                                                    className="border-b hover:bg-muted/50 transition-colors"
                                                >
                                                    <td className="p-3">{index + 1}</td>
                                                    <td className="p-3">
                                                        <div className="max-w-[200px] truncate" title={user.email}>
                                                            {user.email}
                                                        </div>
                                                    </td>
                                                    <td className="p-3">{user.username}</td>
                                                    <td className="p-3">
                                                        <div className="max-w-[150px] truncate" title={user.fullName}>
                                                            {user.fullName}
                                                        </div>
                                                    </td>
                                                    <td className="p-3">{user.studentId || '-'}</td>
                                                    <td className="p-3">
                                                        <Select
                                                            defaultValue={user.role || 'student'}
                                                            onValueChange={(value) => handleRoleChange(user.id, value)}
                                                            disabled={updateRoleMutation.isPending || !user.isActive}
                                                        >
                                                            <SelectTrigger className="w-32">
                                                                <SelectValue />
                                                            </SelectTrigger>
                                                            <SelectContent>
                                                                <SelectItem value="student">Student</SelectItem>
                                                                <SelectItem value="lecturer">Lecturer</SelectItem>
                                                                <SelectItem value="admin">Admin</SelectItem>
                                                            </SelectContent>
                                                        </Select>
                                                    </td>
                                                    <td className="p-3">
                                                        <Badge variant={user.isActive ? 'default' : 'destructive'}>
                                                            {user.isActive ? 'Hoạt động' : 'Vô hiệu hóa'}
                                                        </Badge>
                                                    </td>
                                                    <td className="p-3">
                                                        <div className="text-sm text-muted-foreground">
                                                            {new Date(user.createdAt || '').toLocaleDateString('vi-VN')}
                                                        </div>
                                                    </td>
                                                    <td className="p-3">
                                                        <div className="flex items-center gap-2">
                                                            {user.isActive ? (
                                                                <Button
                                                                    variant="ghost"
                                                                    size="sm"
                                                                    onClick={() => handleToggleActive(user.id)}
                                                                    disabled={toggleActiveMutation.isPending}
                                                                    title="Vô hiệu hóa tài khoản"
                                                                >
                                                                    <Trash2 className="h-4 w-4 text-destructive" />
                                                                </Button>
                                                            ) : (
                                                                <Button
                                                                    variant="ghost"
                                                                    size="sm"
                                                                    onClick={() => handleToggleActive(user.id)}
                                                                    disabled={toggleActiveMutation.isPending}
                                                                    title="Kích hoạt lại tài khoản"
                                                                >
                                                                    <RotateCcw className="h-4 w-4 text-green-600" />
                                                                </Button>
                                                            )}
                                                        </div>
                                                    </td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                )}
                            </div>
                        )}
                        {data && (!data.data || !data.data.users) && (
                            <p className="text-center py-8 text-muted-foreground">
                                Dữ liệu không hợp lệ
                            </p>
                        )}
                    </CardContent>
                </Card>
            </div>
        </MainLayout>
    )
}

export const Route = createFileRoute('/users')({
    component: UsersPage,
})

