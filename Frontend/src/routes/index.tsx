import { useEffect } from 'react'
import { zodResolver } from '@hookform/resolvers/zod'
import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { useForm } from 'react-hook-form'
import * as z from 'zod'

import { ModeToggle } from '@/components/mode-toggle'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { useLogin } from '@/hooks/use-auth'
import { useAuthStore } from '@/stores/use-auth-store'

const loginSchema = z.object({
    identifier: z.string().min(1, 'Vui lòng nhập email hoặc username'),
    password: z.string().min(6, 'Mật khẩu phải có ít nhất 6 ký tự'),
})

type LoginFormValues = z.infer<typeof loginSchema>

function LoginPage() {
    const navigate = useNavigate()
    const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
    const userRole = useAuthStore((state) => state.userRole)
    const { mutate: login, isPending, error } = useLogin()

    const form = useForm<LoginFormValues>({
        resolver: zodResolver(loginSchema),
        defaultValues: {
            identifier: '',
            password: '',
        },
    })

    // Redirect nếu đã đăng nhập - dựa theo role (cho trường hợp user đã login vào lại trang login)
    useEffect(() => {
        if (isAuthenticated && userRole) {
            let redirectPath = '/dashboard'
            if (userRole === 'student') redirectPath = '/practice'
            else if (userRole === 'lecturer') redirectPath = '/grading'
            navigate({ to: redirectPath as any })
        }
    }, [isAuthenticated, userRole, navigate])

    function onSubmit(values: LoginFormValues) {
        login(
            {
                identifier: values.identifier,
                password: values.password,
            },
            {
                onSuccess: (data) => {
                    // Redirect sau khi login thành công
                    if (data.data?.accessToken) {
                        const role = data.data.user.role
                        let redirectPath = '/dashboard'
                        if (role === 'student') redirectPath = '/practice'
                        else if (role === 'lecturer') redirectPath = '/grading'

                        // Delay nhỏ để đảm bảo state đã persist
                        setTimeout(() => {
                            navigate({ to: redirectPath as any })
                        }, 100)
                    }
                },
                onError: (err: any) => {
                    console.error('Đăng nhập thất bại:', err)
                },
            }
        )
    }

    return (
        <div className="min-h-screen flex flex-col items-center justify-center p-4 bg-gradient-to-br from-background to-muted/20">
            <div className="absolute top-4 right-4">
                <ModeToggle />
            </div>

            <Card className="w-full max-w-md shadow-lg">
                <CardHeader className="space-y-1 text-center">
                    <CardTitle className="text-3xl font-bold">Đăng Nhập</CardTitle>
                    <CardDescription>
                        Hệ Thống Chấm Thi và Gỡ Lỗi Truy Vấn SQL
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <Form {...form}>
                        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                            <FormField
                                control={form.control}
                                name="identifier"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Email hoặc Username</FormLabel>
                                        <FormControl>
                                            <Input
                                                placeholder="Nhập email hoặc username"
                                                {...field}
                                            />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="password"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Mật khẩu</FormLabel>
                                        <FormControl>
                                            <Input
                                                placeholder="••••••••"
                                                type="password"
                                                {...field}
                                            />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            {error && (
                                <div className="text-sm text-destructive bg-destructive/10 p-3 rounded-md">
                                    {error.message || 'Đăng nhập thất bại. Vui lòng thử lại!'}
                                </div>
                            )}

                            <Button type="submit" className="w-full" disabled={isPending}>
                                {isPending ? 'Đang đăng nhập...' : 'Đăng Nhập'}
                            </Button>
                        </form>
                    </Form>

                    <div className="mt-4 text-center text-sm">
                        <span className="text-muted-foreground">Chưa có tài khoản? </span>
                        <Link
                            to="/register"
                            className="text-primary hover:underline font-medium"
                        >
                            Đăng ký ngay
                        </Link>
                    </div>
                </CardContent>
            </Card>
        </div>
    )
}

export const Route = createFileRoute('/')({
    component: LoginPage,
})
