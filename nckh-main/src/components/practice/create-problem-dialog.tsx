import { useState, useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Plus } from 'lucide-react'
import toast from 'react-hot-toast'
import { useQuery } from '@tanstack/react-query'

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
    FormDescription,
} from '@/components/ui/form'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Checkbox } from '@/components/ui/checkbox'
import { problemsService } from '@/services/problems.service'
import { topicsService } from '@/services/topics.service'
import { Editor } from '@/components/ui/editor'

const problemSchema = z.object({
    title: z.string().min(3, 'Tiêu đề phải có ít nhất 3 ký tự'),
    slug: z.string().min(3, 'Slug phải có ít nhất 3 ký tự'),
    description: z.string().min(10, 'Mô tả phải có ít nhất 10 ký tự'),
    difficulty: z.enum(['easy', 'medium', 'hard']),
    topicId: z.number().optional(),

    // Đã sửa: Bắt buộc nhập Init Script và Solution Query
    initScript: z.string().min(1, 'Script khởi tạo là bắt buộc'),
    solutionQuery: z.string().min(1, 'Câu truy vấn đáp án là bắt buộc'),

    supportedDatabases: z.array(z.string()).min(1, 'Chọn ít nhất 1 database'),
    orderMatters: z.boolean(),
    isPublic: z.boolean(),
})

type ProblemFormValues = z.infer<typeof problemSchema>

interface CreateProblemDialogProps {
    onSuccess?: () => void
}

export function CreateProblemDialog({ onSuccess }: CreateProblemDialogProps) {
    const [open, setOpen] = useState(false)
    const [isSubmitting, setIsSubmitting] = useState(false)

    const { data: topicsData = [] } = useQuery({
        queryKey: ['topics'],
        queryFn: () => topicsService.list(),
    })
    const topics = Array.isArray(topicsData) ? topicsData : []

    const form = useForm<ProblemFormValues>({
        resolver: zodResolver(problemSchema),
        defaultValues: {
            title: '',
            slug: '',
            description: '',
            difficulty: 'easy',
            topicId: undefined,
            initScript: '',
            solutionQuery: '', // Thêm default value
            supportedDatabases: ['postgresql'],
            orderMatters: false,
            isPublic: true,
        },
    })

    const onSubmit = async (values: ProblemFormValues) => {
        setIsSubmitting(true)
        try {
            await problemsService.create(values as any)
            toast.success('Tạo bài tập thành công!')
            setOpen(false)
            form.reset()
            onSuccess?.()
        } catch (error: any) {
            toast.error(error?.message || 'Tạo bài tập thất bại!')
        } finally {
            setIsSubmitting(false)
        }
    }

    useEffect(() => {
        const subscription = form.watch((_, { name }) => {
            if (name === 'topicId') {
                const currentTitle = form.getValues('title')
                if (currentTitle) {
                    handleTitleChange(currentTitle)
                }
            }
        })
        return () => subscription.unsubscribe()
    }, [form, topics])

    const handleTitleChange = (title: string) => {
        const selectedTopicId = form.getValues('topicId')
        const selectedTopic = topics.find(t => t.id === selectedTopicId)

        const titleSlug = title
            .toLowerCase()
            .normalize('NFD')
            .replace(/[\u0300-\u036f]/g, '')
            .replace(/đ/g, 'd')
            .replace(/[^a-z0-9]/g, '')
            .trim()

        const slug = selectedTopic && selectedTopic.slug
            ? `${selectedTopic.slug}-${titleSlug}`
            : titleSlug

        form.setValue('slug', slug)
    }

    const databases = ['postgresql', 'mysql', 'sqlite']

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger asChild>
                <Button>
                    <Plus className="h-4 w-4 mr-2" />
                    Tạo bài tập
                </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-[800px] max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>Tạo bài tập mới</DialogTitle>
                    <DialogDescription>
                        Tạo bài tập SQL cho sinh viên luyện tập
                    </DialogDescription>
                </DialogHeader>

                <Form {...form}>
                    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                        <div className="grid grid-cols-2 gap-4">
                            <FormField
                                control={form.control}
                                name="title"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Tiêu đề <span className="text-destructive">*</span></FormLabel>
                                        <FormControl>
                                            <Input
                                                placeholder="Simple SELECT"
                                                {...field}
                                                onChange={(e) => {
                                                    field.onChange(e)
                                                    handleTitleChange(e.target.value)
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
                                        <FormLabel>Slug <span className="text-destructive">*</span></FormLabel>
                                        <FormControl>
                                            <Input placeholder="simple-select" {...field} />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                            <FormField
                                control={form.control}
                                name="difficulty"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Độ khó <span className="text-destructive">*</span></FormLabel>
                                        <Select onValueChange={field.onChange} defaultValue={field.value}>
                                            <FormControl>
                                                <SelectTrigger>
                                                    <SelectValue placeholder="Chọn độ khó" />
                                                </SelectTrigger>
                                            </FormControl>
                                            <SelectContent>
                                                <SelectItem value="easy">Dễ</SelectItem>
                                                <SelectItem value="medium">Trung bình</SelectItem>
                                                <SelectItem value="hard">Khó</SelectItem>
                                            </SelectContent>
                                        </Select>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="topicId"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Chủ đề</FormLabel>
                                        <Select
                                            onValueChange={(value) => field.onChange(value ? parseInt(value) : undefined)}
                                            value={field.value?.toString()}
                                        >
                                            <FormControl>
                                                <SelectTrigger>
                                                    <SelectValue placeholder="Chọn chủ đề" />
                                                </SelectTrigger>
                                            </FormControl>
                                            <SelectContent>
                                                {topics.map((topic) => (
                                                    <SelectItem key={topic.id} value={topic.id.toString()}>
                                                        {topic.icon && <span className="mr-2">{topic.icon}</span>}
                                                        {topic.name}
                                                    </SelectItem>
                                                ))}
                                            </SelectContent>
                                        </Select>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                        </div>

                        <FormField
                            control={form.control}
                            name="description"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Mô tả <span className="text-destructive">*</span></FormLabel>
                                    <FormControl>
                                        <Editor
                                            placeholder="Select all employees from the employees table"
                                            value={field.value}
                                            onChange={field.onChange}
                                        />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        {/* Script khởi tạo */}
                        <FormField
                            control={form.control}
                            name="initScript"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Script khởi tạo (CREATE/INSERT) <span className="text-destructive">*</span></FormLabel>
                                    <FormControl>
                                        <Textarea
                                            placeholder="CREATE TABLE users (id INT, name VARCHAR(50)); INSERT INTO users VALUES (1, 'Alice');"
                                            rows={5}
                                            className="font-mono text-sm"
                                            {...field}
                                        />
                                    </FormControl>
                                    <FormDescription>
                                        SQL script chạy trước khi thực thi query của sinh viên.
                                    </FormDescription>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        {/* SQL Đáp án - quan trọng */}
                        <FormField
                            control={form.control}
                            name="solutionQuery"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>SQL Đáp án (Solution Query) <span className="text-destructive">*</span></FormLabel>
                                    <FormControl>
                                        <Textarea
                                            placeholder="SELECT * FROM users WHERE id = 1;"
                                            rows={3}
                                            className="font-mono text-sm"
                                            {...field}
                                        />
                                    </FormControl>
                                    <FormDescription>
                                        Câu SQL chuẩn để so sánh kết quả.
                                    </FormDescription>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="supportedDatabases"
                            render={() => (
                                <FormItem>
                                    <FormLabel>Databases hỗ trợ <span className="text-destructive">*</span></FormLabel>
                                    <div className="flex gap-4">
                                        {databases.map((db) => (
                                            <FormField
                                                key={db}
                                                control={form.control}
                                                name="supportedDatabases"
                                                render={({ field }) => (
                                                    <FormItem className="flex items-center space-x-2 space-y-0">
                                                        <FormControl>
                                                            <Checkbox
                                                                checked={field.value?.includes(db)}
                                                                onCheckedChange={(checked) => {
                                                                    return checked
                                                                        ? field.onChange([...field.value, db])
                                                                        : field.onChange(
                                                                            field.value?.filter(
                                                                                (value) => value !== db
                                                                            )
                                                                        )
                                                                }}
                                                            />
                                                        </FormControl>
                                                        <FormLabel className="font-normal cursor-pointer capitalize">
                                                            {db}
                                                        </FormLabel>
                                                    </FormItem>
                                                )}
                                            />
                                        ))}
                                    </div>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <div className="flex gap-6 pt-2">
                            <FormField
                                control={form.control}
                                name="orderMatters"
                                render={({ field }) => (
                                    <FormItem className="flex items-center space-x-2 space-y-0">
                                        <FormControl>
                                            <Checkbox
                                                checked={field.value}
                                                onCheckedChange={field.onChange}
                                            />
                                        </FormControl>
                                        <FormLabel className="font-normal cursor-pointer">
                                            Thứ tự kết quả quan trọng
                                        </FormLabel>
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="isPublic"
                                render={({ field }) => (
                                    <FormItem className="flex items-center space-x-2 space-y-0">
                                        <FormControl>
                                            <Checkbox
                                                checked={field.value}
                                                onCheckedChange={field.onChange}
                                            />
                                        </FormControl>
                                        <FormLabel className="font-normal cursor-pointer">
                                            Công khai bài tập
                                        </FormLabel>
                                    </FormItem>
                                )}
                            />
                        </div>

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
                                {isSubmitting ? 'Đang tạo...' : 'Tạo bài tập'}
                            </Button>
                        </div>
                    </form>
                </Form>
            </DialogContent>
        </Dialog>
    )
}