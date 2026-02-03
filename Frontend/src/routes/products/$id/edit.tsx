import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import toast from 'react-hot-toast'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

function ProductEdit() {
    const { id } = Route.useParams()
    const navigate = useNavigate()

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault()
        // Logic cập nhật sản phẩm ở đây
        toast.success(`Cập nhật sản phẩm #${id} thành công!`)
        navigate({ to: '/products/$id', params: { id } })
    }

    return (
        <div className="container mx-auto p-6 max-w-2xl">
            <h1 className="text-3xl font-bold mb-6">Sửa sản phẩm #{id}</h1>

            <Card>
                <CardHeader>
                    <CardTitle>Chỉnh sửa thông tin</CardTitle>
                    <CardDescription>Cập nhật thông tin sản phẩm</CardDescription>
                </CardHeader>
                <CardContent>
                    <form onSubmit={handleSubmit} className="space-y-4">
                        <div className="space-y-2">
                            <Label htmlFor="name">Tên sản phẩm</Label>
                            <Input id="name" defaultValue={`Sản phẩm ${id}`} required />
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="price">Giá</Label>
                            <Input id="price" type="number" defaultValue="100000" required />
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="description">Mô tả</Label>
                            <Input id="description" defaultValue={`Mô tả sản phẩm ${id}`} />
                        </div>

                        <div className="flex gap-2 pt-4">
                            <Button type="submit">Cập nhật</Button>
                            <Link to="/products/$id" params={{ id }}>
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

export const Route = createFileRoute('/products/$id/edit')({
    component: ProductEdit,
})
