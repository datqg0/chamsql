import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Plus } from 'lucide-react'
import toast from 'react-hot-toast'

import { Button } from '@/components/ui/button'
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from '@/components/ui/dialog'
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { topicsService } from '@/services/topics.service'

const topicSchema = z.object({
    name: z.string().min(3, 'Tên chủ đề phải có ít nhất 3 ký tự'),
    slug: z.string().min(3, 'Slug phải có ít nhất 3 ký tự'),
    description: z.string().optional(),
    icon: z.string().optional(),
    sortOrder: z.number().int().min(0).optional(),
})

type TopicFormValues = z.infer<typeof topicSchema>

interface CreateTopicDialogProps {
    onSuccess?: () => void
}

export function CreateTopicDialog({ onSuccess }: CreateTopicDialogProps) {
    const [open, setOpen] = useState(false)
    const [isSubmitting, setIsSubmitting] = useState(false)

    const form = useForm<TopicFormValues>({
        resolver: zodResolver(topicSchema),
        defaultValues: {
            name: '',
            slug: '',
            description: '',
            icon: '📊',
            sortOrder: 1,
        },
    })

    const onSubmit = async (values: TopicFormValues) => {
        setIsSubmitting(true)
        try {
            await topicsService.create(values)
            toast.success('Tạo chủ đề thành công!')
            setOpen(false)
            form.reset()
            onSuccess?.()
        } catch (error: any) {
            toast.error(error?.message || 'Tạo chủ đề thất bại!')
        } finally {
            setIsSubmitting(false)
        }
    }

    // Auto-generate slug from name
    const handleNameChange = (name: string) => {
        const slug = name
            .toLowerCase()
            .normalize('NFD')
            .replace(/[\u0300-\u036f]/g, '')
            .replace(/đ/g, 'd')
            .replace(/[^a-z0-9]/g, '')
            .trim()
        form.setValue('slug', slug)
    }

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger asChild>
                <Button>
                    <Plus className="h-4 w-4 mr-2" />
                    Tạo chủ đề
                </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-[500px]">
                <DialogHeader>
                    <DialogTitle>Tạo chủ đề mới</DialogTitle>
                    <DialogDescription>
                        Tạo chủ đề để nhóm các bài tập SQL liên quan
                    </DialogDescription>
                </DialogHeader>

                <Form {...form}>
                    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                        <FormField
                            control={form.control}
                            name="name"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Tên chủ đề</FormLabel>
                                    <FormControl>
                                        <Input
                                            placeholder="Basic SELECT"
                                            {...field}
                                            onChange={(e) => {
                                                field.onChange(e)
                                                handleNameChange(e.target.value)
                                            }}
                                        />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="slug"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Slug</FormLabel>
                                    <FormControl>
                                        <Input placeholder="basic-select" {...field} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="description"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Mô tả (tùy chọn)</FormLabel>
                                    <FormControl>
                                        <Textarea
                                            placeholder="Learn basic SELECT queries"
                                            {...field}
                                        />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="icon"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Icon (tùy chọn)</FormLabel>
                                    <FormControl>
                                        <Input placeholder="📊" {...field} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="sortOrder"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Thứ tự sắp xếp</FormLabel>
                                    <FormControl>
                                        <Input
                                            type="number"
                                            {...field}
                                            onChange={(e) => field.onChange(parseInt(e.target.value) || 1)}
                                        />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <div className="flex justify-end gap-2 pt-4">
                            <Button
                                type="button"
                                variant="outline"
                                onClick={() => setOpen(false)}
                                disabled={isSubmitting}
                            >
                                Hủy
                            </Button>
                            <Button type="submit" disabled={isSubmitting}>
                                {isSubmitting ? 'Đang tạo...' : 'Tạo chủ đề'}
                            </Button>
                        </div>
                    </form>
                </Form>
            </DialogContent>
        </Dialog>
    )
}
