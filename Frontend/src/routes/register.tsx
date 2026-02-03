import { zodResolver } from '@hookform/resolvers/zod'
import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { useEffect } from 'react'
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
import { useRegister } from '@/hooks/use-auth'
import { useAuthStore } from '@/stores/use-auth-store'

const registerSchema = z
    .object({
        email: z.string().email('Email không hợp lệ'),
        username: z.string().min(3, 'Username phải có ít nhất 3 ký tự'),
        password: z.string().min(6, 'Mật khẩu phải có ít nhất 6 ký tự'),
        confirmPassword: z.string(),
        fullName: z.string().min(2, 'Họ tên phải có ít nhất 2 ký tự'),
        studentId: z.string().min(1, 'Mã sinh viên không được để trống'),
    })
    .refine((data) => data.password === data.confirmPassword, {
        message: 'Mật khẩu xác nhận không khớp',
        path: ['confirmPassword'],
    })

type RegisterFormValues = z.infer<typeof registerSchema>

function RegisterPage() {
    const navigate = useNavigate()
    const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
    const { mutate: register, isPending, error, isSuccess } = useRegister()

    const form = useForm<RegisterFormValues>({
        resolver: zodResolver(registerSchema),
        defaultValues: {
            email: '',
            username: '',
            password: '',
            confirmPassword: '',
            fullName: '',
            studentId: '',
        },
    })

    // Redirect sau khi đăng ký thành công - delay 1.5s để toast hiện
    useEffect(() => {
        if (isSuccess && isAuthenticated) {
            const timer = setTimeout(() => {
                navigate({ to: '/dashboard' })
            }, 1500)
            return () => clearTimeout(timer)
        }
    }, [isSuccess, isAuthenticated, navigate])

    function onSubmit(values: RegisterFormValues) {
        register(
            {
                email: values.email,
                username: values.username,
                password: values.password,
                fullName: values.fullName,
                studentId: values.studentId,
            },
            {
                onError: (err: any) => {
                    console.error('Đăng ký thất bại:', err)
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
                    <CardTitle className="text-3xl font-bold">Đăng Ký</CardTitle>
                    <CardDescription>
                        Tạo tài khoản mới để sử dụng hệ thống
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    {isSuccess ? (
                        <div className="space-y-4 text-center">
                            <div className="text-green-600 dark:text-green-400 font-semibold">
                                Đăng ký thành công!
                            </div>
                            <p className="text-sm text-muted-foreground">
                                Đang chuyển hướng đến trang đăng nhập...
                            </p>
                        </div>
                    ) : (
                        <Form {...form}>
                            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                                <FormField
                                    control={form.control}
                                    name="fullName"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Họ và tên</FormLabel>
                                            <FormControl>
                                                <Input
                                                    placeholder="Nhập họ và tên"
                                                    {...field}
                                                />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />

                                <FormField
                                    control={form.control}
                                    name="username"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Username</FormLabel>
                                            <FormControl>
                                                <Input
                                                    placeholder="Nhập username"
                                                    {...field}
                                                />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />

                                <FormField
                                    control={form.control}
                                    name="studentId"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Mã sinh viên</FormLabel>
                                            <FormControl>
                                                <Input
                                                    placeholder="Nhập mã sinh viên"
                                                    {...field}
                                                />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />

                                <FormField
                                    control={form.control}
                                    name="email"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Email</FormLabel>
                                            <FormControl>
                                                <Input
                                                    placeholder="Nhập email"
                                                    type="email"
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
                                                    placeholder="Nhập mật khẩu"
                                                    type="password"
                                                    {...field}
                                                />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />

                                <FormField
                                    control={form.control}
                                    name="confirmPassword"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Xác nhận mật khẩu</FormLabel>
                                            <FormControl>
                                                <Input
                                                    placeholder="Nhập lại mật khẩu"
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
                                        {error.message || 'Đăng ký thất bại. Vui lòng thử lại!'}
                                    </div>
                                )}

                                <Button type="submit" className="w-full" disabled={isPending}>
                                    {isPending ? 'Đang đăng ký...' : 'Đăng Ký'}
                                </Button>
                            </form>
                        </Form>
                    )}

                    <div className="mt-4 text-center text-sm">
                        <span className="text-muted-foreground">Đã có tài khoản? </span>
                        <Link to="/" className="text-primary hover:underline font-medium">
                            Đăng nhập ngay
                        </Link>
                    </div>
                </CardContent>
            </Card>
        </div>
    )
}

export const Route = createFileRoute('/register')({
    component: RegisterPage,
})

