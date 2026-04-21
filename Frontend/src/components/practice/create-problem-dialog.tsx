import { zodResolver } from '@hookform/resolvers/zod'
import { useQuery } from '@tanstack/react-query'
import { Plus } from 'lucide-react'
import { useState, useEffect } from 'react'
import { useForm, useFieldArray } from 'react-hook-form'
import toast from 'react-hot-toast'
import { Trash2, AlertCircle } from 'lucide-react'
import * as z from 'zod'

import { Button } from '@/components/ui/button'
import {
    Card,
    CardContent,
    CardHeader,
    CardTitle,
} from '@/components/ui/card'
import { Checkbox } from '@/components/ui/checkbox'
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from '@/components/ui/dialog'
import { Editor } from '@/components/ui/editor'
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
    FormDescription,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import type { Problem } from '@/types/exam.types'
import { problemsService } from '@/services/problems.service'
import { topicsService } from '@/services/topics.service'

const testCaseSchema = z.object({
    name: z.string().min(1, 'Tên test case là bắt buộc'),
    description: z.string().optional(),
    initScript: z.string().min(1, 'Script khởi tạo là bắt buộc'),
    solutionQuery: z.string().min(1, 'Câu truy vấn đáp án là bắt buộc'),
    weight: z.number().min(1, 'Trọng số phải ít nhất là 1'),
    isHidden: z.boolean(),
})

const problemSchema = z.object({
    title: z.string().min(3, 'Tiêu đề phải có ít nhất 3 ký tự'),
    slug: z.string().min(3, 'Slug phải có ít nhất 3 ký tự'),
    description: z.string().min(10, 'Mô tả phải có ít nhất 10 ký tự'),
    difficulty: z.enum(['easy', 'medium', 'hard']),
    topicId: z.number().optional(),

    initScript: z.string().optional(),
    solutionQuery: z.string().optional(),

    supportedDatabases: z.array(z.string()).min(1, 'Chọn ít nhất 1 database'),
    orderMatters: z.boolean(),
    isPublic: z.boolean(),
    testCases: z.array(testCaseSchema).optional(),
})

type ProblemFormValues = z.infer<typeof problemSchema>

interface CreateProblemDialogProps {
    onSuccess?: () => void
    problem?: Problem // Add problem prop for editing
}

export function CreateProblemDialog({ onSuccess, problem }: CreateProblemDialogProps) {
    const isEdit = !!problem
    const [open, setOpen] = useState(false)
    const [isSubmitting, setIsSubmitting] = useState(false)

    const { data: topicsData = [] } = useQuery({
        queryKey: ['topics'],
        queryFn: () => topicsService.list(),
    })
    const topics = Array.isArray(topicsData) ? topicsData : []

    const { data: fullProblem, isLoading: _isLoadingFullProblem } = useQuery({
        queryKey: ['problem', problem?.slug],
        queryFn: () => problemsService.getBySlug(problem!.slug),
        enabled: isEdit && open,
    })

    const form = useForm<ProblemFormValues>({
        resolver: zodResolver(problemSchema),
        defaultValues: {
            title: problem?.title || '',
            slug: problem?.slug || '',
            description: problem?.description || '',
            difficulty: (problem?.difficulty as any) || 'easy',
            topicId: problem?.topicId,
            initScript: problem?.initScript || '',
            solutionQuery: problem?.solutionQuery || '',
            supportedDatabases: problem?.supportedDatabases || ['postgresql'],
            orderMatters: problem?.orderMatters || false,
            isPublic: problem?.isPublic ?? true,
            testCases: problem?.testCases || [],
        },
    })

    // Reset form when problem changes or full data loads
    useEffect(() => {
        if (isEdit && fullProblem && open) {
            form.reset({
                title: fullProblem.title,
                slug: fullProblem.slug,
                description: fullProblem.description,
                difficulty: fullProblem.difficulty as any,
                topicId: fullProblem.topicId,
                initScript: fullProblem.initScript || '',
                solutionQuery: fullProblem.solutionQuery || '',
                supportedDatabases: fullProblem.supportedDatabases && fullProblem.supportedDatabases.length > 0 ? fullProblem.supportedDatabases : ['postgresql'],
                orderMatters: fullProblem.orderMatters || false,
                isPublic: fullProblem.isPublic,
                testCases: fullProblem.testCases || [],
            })
        } else if (!isEdit && open) {
            form.reset({
                title: '',
                slug: '',
                description: '',
                difficulty: 'easy',
                topicId: undefined,
                initScript: '',
                solutionQuery: '',
                supportedDatabases: ['postgresql'],
                orderMatters: false,
                isPublic: true,
                testCases: [],
            })
        }
    }, [fullProblem, isEdit, open, form])

    const { fields, append, remove } = useFieldArray({
        control: form.control,
        name: "testCases"
    })

    const onSubmit = async (values: ProblemFormValues) => {
        setIsSubmitting(true)
        try {
            if (isEdit && problem) {
                await problemsService.update(problem.id, values as any)
                toast.success('Cập nhật bài tập thành công!')
            } else {
                await problemsService.create(values as any)
                toast.success('Tạo bài tập thành công!')
            }
            setOpen(false)
            if (!isEdit) form.reset()
            onSuccess?.()
        } catch (error: any) {
            toast.error(error?.message || (isEdit ? 'Cập nhật bài tập thất bại!' : 'Tạo bài tập thất bại!'))
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
                {isEdit ? (
                    <Button variant="ghost" size="icon" className="h-8 w-8 text-primary hover:text-primary hover:bg-primary/10">
                        <Plus className="h-4 w-4 rotate-45 hidden" /> {/* Dummy to keep imports */}
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            width="24"
                            height="24"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            strokeWidth="2"
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            className="h-4 w-4"
                        >
                            <path d="M17 3a2.85 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z" />
                            <path d="m15 5 4 4" />
                        </svg>
                    </Button>
                ) : (
                    <Button>
                        <Plus className="h-4 w-4 mr-2" />
                        Tạo bài tập
                    </Button>
                )}
            </DialogTrigger>
            <DialogContent className="sm:max-w-[800px] max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>{isEdit ? 'Chỉnh sửa bài tập' : 'Tạo bài tập mới'}</DialogTitle>
                    <DialogDescription>
                        {isEdit ? 'Cập nhật thông tin bài tập SQL' : 'Tạo bài tập SQL cho sinh viên luyện tập'}
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
                                    <FormLabel>Script khởi tạo mặc định (CREATE/INSERT)</FormLabel>
                                    <FormControl>
                                        <Textarea
                                            placeholder="CREATE TABLE users (id INT, name VARCHAR(50)); INSERT INTO users VALUES (1, 'Alice');"
                                            rows={5}
                                            className="font-mono text-sm"
                                            {...field}
                                        />
                                    </FormControl>
                                    <FormDescription>
                                        SQL script chạy trước khi thực thi query mẫu. Sẽ được dùng làm mặc định nếu không có test case nào.
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
                                    <FormLabel>SQL Đáp án mẫu (Solution Query)</FormLabel>
                                    <FormControl>
                                        <Textarea
                                            placeholder="SELECT * FROM users WHERE id = 1;"
                                            rows={3}
                                            className="font-mono text-sm"
                                            {...field}
                                        />
                                    </FormControl>
                                    <FormDescription>
                                        Câu SQL chuẩn để so sánh kết quả. Sẽ được dùng làm mặc định nếu không có test case nào.
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
                                    <FormLabel>Database hỗ trợ <span className="text-destructive">*</span></FormLabel>
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
                                            xếp theo thứ tự kết quả
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

                        {/* Test Cases Section */}
                        <div className="space-y-4 pt-4 border-t">
                            <div className="flex items-center justify-between">
                                <div>
                                    <h3 className="text-lg font-medium">Test Cases</h3>
                                    <p className="text-sm text-muted-foreground">Thêm các bộ test để chấm điểm chi tiết.</p>
                                </div>
                                <Button
                                    type="button"
                                    variant="outline"
                                    size="sm"
                                    onClick={() => append({
                                        name: `Test Case ${fields.length + 1}`,
                                        initScript: '',
                                        solutionQuery: '',
                                        weight: 1,
                                        isHidden: false
                                    })}
                                >
                                    <Plus className="h-4 w-4 mr-2" />
                                    Thêm Test Case
                                </Button>
                            </div>

                            {fields.length === 0 && (
                                <div className="p-4 border border-dashed rounded-md text-center text-muted-foreground bg-muted/5">
                                    <AlertCircle className="h-8 w-8 mx-auto mb-2 opacity-20" />
                                    <p className="text-sm">Chưa có test case nào. Hệ thống sẽ dùng Script khởi tạo & SQL Đáp án ở trên làm test case mặc định.</p>
                                </div>
                            )}

                            <div className="space-y-4">
                                {fields.map((field, index) => (
                                    <Card key={field.id} className="relative overflow-hidden">
                                        <CardHeader className="py-3 px-4 bg-muted/30 flex flex-row items-center justify-between space-y-0">
                                            <CardTitle className="text-sm font-medium">
                                                #{index + 1}: {form.watch(`testCases.${index}.name`) || 'Test Case'}
                                            </CardTitle>
                                            <Button
                                                type="button"
                                                variant="ghost"
                                                size="icon"
                                                className="h-8 w-8 text-destructive hover:text-destructive hover:bg-destructive/10"
                                                onClick={() => remove(index)}
                                            >
                                                <Trash2 className="h-4 w-4" />
                                            </Button>
                                        </CardHeader>
                                        <CardContent className="p-4 space-y-4">
                                            <div className="grid grid-cols-2 gap-4">
                                                <FormField
                                                    control={form.control}
                                                    name={`testCases.${index}.name`}
                                                    render={({ field }) => (
                                                        <FormItem>
                                                            <FormLabel className="text-xs">Tên Test <span className="text-destructive">*</span></FormLabel>
                                                            <FormControl>
                                                                <Input placeholder="Sample 1" {...field} className="h-8 text-xs" />
                                                            </FormControl>
                                                            <FormMessage className="text-[10px]" />
                                                        </FormItem>
                                                    )}
                                                />
                                                <div className="flex gap-4 items-end pb-1">
                                                    <FormField
                                                        control={form.control}
                                                        name={`testCases.${index}.weight`}
                                                        render={({ field }) => (
                                                            <FormItem className="flex-1">
                                                                <FormLabel className="text-xs">Trọng số <span className="text-destructive">*</span></FormLabel>
                                                                <FormControl>
                                                                    <Input
                                                                        type="number"
                                                                        {...field}
                                                                        onChange={e => field.onChange(parseInt(e.target.value) || 1)}
                                                                        className="h-8 text-xs"
                                                                    />
                                                                </FormControl>
                                                                <FormMessage className="text-[10px]" />
                                                            </FormItem>
                                                        )}
                                                    />
                                                    <FormField
                                                        control={form.control}
                                                        name={`testCases.${index}.isHidden`}
                                                        render={({ field }) => (
                                                            <FormItem className="flex items-center space-x-2 space-y-0 mb-2">
                                                                <FormControl>
                                                                    <Checkbox
                                                                        checked={field.value}
                                                                        onCheckedChange={field.onChange}
                                                                    />
                                                                </FormControl>
                                                                <FormLabel className="text-xs font-normal cursor-pointer">
                                                                    Ẩn
                                                                </FormLabel>
                                                            </FormItem>
                                                        )}
                                                    />
                                                </div>
                                            </div>

                                            <FormField
                                                control={form.control}
                                                name={`testCases.${index}.initScript`}
                                                render={({ field }) => (
                                                    <FormItem>
                                                        <FormLabel className="text-xs">Script khởi tạo <span className="text-destructive">*</span></FormLabel>
                                                        <FormControl>
                                                            <Textarea
                                                                placeholder="INSERT INTO..."
                                                                {...field}
                                                                className="font-mono text-[10px] min-h-[60px]"
                                                            />
                                                        </FormControl>
                                                        <FormMessage className="text-[10px]" />
                                                    </FormItem>
                                                )}
                                            />

                                            <FormField
                                                control={form.control}
                                                name={`testCases.${index}.solutionQuery`}
                                                render={({ field }) => (
                                                    <FormItem>
                                                        <FormLabel className="text-xs">SQL Đáp án <span className="text-destructive">*</span></FormLabel>
                                                        <FormControl>
                                                            <Textarea
                                                                placeholder="SELECT..."
                                                                {...field}
                                                                className="font-mono text-[10px] min-h-[60px]"
                                                            />
                                                        </FormControl>
                                                        <FormMessage className="text-[10px]" />
                                                    </FormItem>
                                                )}
                                            />
                                        </CardContent>
                                    </Card>
                                ))}
                            </div>
                        </div>

                        <div className="flex justify-end gap-2 pt-6 border-t">
                            <Button
                                type="button"
                                variant="outline"
                                onClick={() => setOpen(false)}
                                disabled={isSubmitting}
                            >
                                Hủy
                            </Button>
                            <Button type="submit" disabled={isSubmitting}>
                                {isSubmitting ? (isEdit ? 'Đang cập nhật...' : 'Đang tạo...') : (isEdit ? 'Cập nhật bài tập' : 'Tạo bài tập')}
                            </Button>
                        </div>
                    </form>
                </Form>
            </DialogContent>
        </Dialog>
    )
}