import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import toast from 'react-hot-toast'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

function ProductCreate() {
    const navigate = useNavigate()

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault()
        // Logic tạo sản phẩm ở đây
        toast.success('Tạo sản phẩm thành công!')
        navigate({ to: '/products' })
    }

    return (
        <div className="container mx-auto p-6 max-w-2xl">
            <h1 className="text-3xl font-bold mb-6">Tạo sản phẩm mới</h1>

            <Card>
                <CardHeader>
                    <CardTitle>Thông tin sản phẩm</CardTitle>
                    <CardDescription>Điền thông tin để tạo sản phẩm mới</CardDescription>
                </CardHeader>
                <CardContent>
                    <form onSubmit={handleSubmit} className="space-y-4">
                        <div className="space-y-2">
                            <Label htmlFor="name">Tên sản phẩm</Label>
                            <Input id="name" placeholder="Nhập tên sản phẩm" required />
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="price">Giá</Label>
                            <Input id="price" type="number" placeholder="0" required />
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="description">Mô tả</Label>
                            <Input id="description" placeholder="Mô tả sản phẩm" />
                        </div>

                        <div className="flex gap-2 pt-4">
                            <Button type="submit">Tạo sản phẩm</Button>
                            <Link to="/products">
                                <Button type="button" variant="outline">
                                    Hủy
                                </Button>
                            </Link>
                        </div>
                    </form>
                </CardContent>
            </Card>
        </div>
    )
}

export const Route = createFileRoute('/products/create')({
    component: ProductCreate,
})
