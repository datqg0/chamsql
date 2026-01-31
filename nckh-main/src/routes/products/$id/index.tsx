import { createFileRoute, Link } from '@tanstack/react-router'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

function ProductDetail() {
    const { id } = Route.useParams()

    return (
        <div className="container mx-auto p-6 max-w-2xl">
            <h1 className="text-3xl font-bold mb-6">Chi tiết sản phẩm #{id}</h1>

            <Card>
                <CardHeader>
                    <CardTitle>Sản phẩm #{id}</CardTitle>
                    <CardDescription>Thông tin chi tiết sản phẩm</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div>
                        <p className="text-sm text-muted-foreground">Tên sản phẩm</p>
                        <p className="font-medium">Sản phẩm số {id}</p>
                    </div>

                    <div>
                        <p className="text-sm text-muted-foreground">Giá</p>
                        <p className="font-medium">100,000 VNĐ</p>
                    </div>

                    <div>
                        <p className="text-sm text-muted-foreground">Mô tả</p>
                        <p className="font-medium">Đây là mô tả của sản phẩm {id}</p>
                    </div>

                    <div className="flex gap-2 pt-4">
                        <Link to="/products/$id/edit" params={{ id }}>
                            <Button>Sửa sản phẩm</Button>
                        </Link>
                        <Link to="/products">
                            <Button variant="outline">← Quay lại danh sách</Button>
                        </Link>
                    </div>
                </CardContent>
            </Card>
        </div>
    )
}

export const Route = createFileRoute('/products/$id/')({
    component: ProductDetail,
})
