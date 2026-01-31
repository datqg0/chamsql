import { createFileRoute, Link } from '@tanstack/react-router'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

function ProductsIndex() {
    return (
        <div className="container mx-auto p-6">
            <div className="flex justify-between items-center mb-6">
                <h1 className="text-3xl font-bold">Danh sách sản phẩm</h1>
                <Link to="/products/create">
                    <Button>Tạo sản phẩm mới</Button>
                </Link>
            </div>

            <div className="grid gap-4">
                <Card>
                    <CardHeader>
                        <CardTitle>Sản phẩm #1</CardTitle>
                        <CardDescription>Mô tả sản phẩm 1</CardDescription>
                    </CardHeader>
                    <CardContent className="flex gap-2">
                        <Link to="/products/$id" params={{ id: '1' }}>
                            <Button variant="outline">Xem chi tiết</Button>
                        </Link>
                        <Link to="/products/$id/edit" params={{ id: '1' }}>
                            <Button variant="outline">Sửa</Button>
                        </Link>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle>Sản phẩm #2</CardTitle>
                        <CardDescription>Mô tả sản phẩm 2</CardDescription>
                    </CardHeader>
                    <CardContent className="flex gap-2">
                        <Link to="/products/$id" params={{ id: '2' }}>
                            <Button variant="outline">Xem chi tiết</Button>
                        </Link>
                        <Link to="/products/$id/edit" params={{ id: '2' }}>
                            <Button variant="outline">Sửa</Button>
                        </Link>
                    </CardContent>
                </Card>
            </div>

            <div className="mt-6">
                <Link to="/">
                    <Button variant="ghost">← Về trang chủ</Button>
                </Link>
            </div>
        </div>
    )
}

export const Route = createFileRoute('/products/')({
    component: ProductsIndex,
})
