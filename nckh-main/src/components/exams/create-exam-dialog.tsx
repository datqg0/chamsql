import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import toast from 'react-hot-toast'

import { Button } from '@/components/ui/button'
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from '@/components/ui/dialog'
import {
    Form,
    FormControl,
    FormDescription,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import { examsService } from '@/services/exams.service'
import type { CreateExamRequest } from '@/types/exam.types'
import { Plus, Loader2 } from 'lucide-react'

const examSchema = z.object({
    title: z.string().min(3, 'Tên kỳ thi phải có ít nhất 3 ký tự'),
    description: z.string().optional(),
    startTime: z.string().min(1, 'Vui lòng chọn thời gian bắt đầu'),
    endTime: z.string().min(1, 'Vui lòng chọn thời gian kết thúc'),
    durationMinutes: z.number().min(1, 'Thời lượng phải lớn hơn 0').max(480, 'Thời lượng tối đa là 480 phút'),
    allowedDatabases: z.array(z.string()).min(1, 'Chọn ít nhất 1 database'),
    allowAiAssistance: z.boolean(),
    shuffleProblems: z.boolean(),
    showResultImmediately: z.boolean(),
    maxAttempts: z.number().min(1, 'Số lần thi phải lớn hơn 0').max(10, 'Tối đa 10 lần'),
    isPublic: z.boolean(),
})

type ExamFormValues = z.infer<typeof examSchema>

interface CreateExamDialogProps {
    onSuccess?: () => void
}

export function CreateExamDialog({ onSuccess }: CreateExamDialogProps) {
    const [open, setOpen] = useState(false)
    const [isSubmitting, setIsSubmitting] = useState(false)

    const form = useForm<ExamFormValues>({
        resolver: zodResolver(examSchema),
        defaultValues: {
            title: '',
            description: '',
            startTime: '',
            endTime: '',
            durationMinutes: 60,
            allowedDatabases: ['postgresql'],
            allowAiAssistance: false,
            shuffleProblems: true,
            showResultImmediately: true,
            maxAttempts: 3,
            isPublic: true,
        },
    })

    const onSubmit = async (values: ExamFormValues) => {
        setIsSubmitting(true)

        try {
            const examData: CreateExamRequest = {
                ...values,
                startTime: new Date(values.startTime).toISOString(),
                endTime: new Date(values.endTime).toISOString(),
            }

            await examsService.create(examData)
            toast.success('Tạo kỳ thi thành công!')
            setOpen(false)
            form.reset()
            onSuccess?.()
        } catch (error: any) {
            toast.error(error?.message || 'Tạo kỳ thi thất bại!')
        } finally {
            setIsSubmitting(false)
        }
    }

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger asChild>
                <Button>
                    <Plus className="h-4 w-4 mr-2" />
                    Tạo kỳ thi
                </Button>
            </DialogTrigger>
            <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>Tạo kỳ thi mới</DialogTitle>
                    <DialogDescription>
                        Điền thông tin để tạo kỳ thi cho sinh viên
                    </DialogDescription>
                </DialogHeader>

                <Form {...form}>
                    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                        {/* Title */}
                        <FormField
                            control={form.control}
                            name="title"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Tên kỳ thi *</FormLabel>
                                    <FormControl>
                                        <Input placeholder="SQL Basics Exam" {...field} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        {/* Description */}
                        <FormField
                            control={form.control}
                            name="description"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Mô tả</FormLabel>
                                    <FormControl>
                                        <Textarea
                                            placeholder="Bài kiểm tra kiến thức SQL cơ bản..."
                                            rows={3}
                                            {...field}
                                        />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        {/* Date & Time */}
                        <div className="grid grid-cols-2 gap-4">
                            <FormField
                                control={form.control}
                                name="startTime"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Thời gian bắt đầu *</FormLabel>
                                        <FormControl>
                                            <Input
                                                type="datetime-local"
                                                {...field}
                                            />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="endTime"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Thời gian kết thúc *</FormLabel>
                                        <FormControl>
                                            <Input
                                                type="datetime-local"
                                                {...field}
                                            />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                        </div>

                        {/* Duration & Max Attempts */}
                        <div className="grid grid-cols-2 gap-4">
                            <FormField
                                control={form.control}
                                name="durationMinutes"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Thời lượng (phút) *</FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="maxAttempts"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Số lần thi tối đa *</FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                {...field}
                                                onChange={(e) => field.onChange(parseInt(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                        </div>

                        {/* Allowed Databases */}
                        <FormField
                            control={form.control}
                            name="allowedDatabases"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Database hỗ trợ *</FormLabel>
                                    <Select
                                        value={field.value[0]}
                                        onValueChange={(value) => field.onChange([value])}
                                    >
                                        <FormControl>
                                            <SelectTrigger>
                                                <SelectValue />
                                            </SelectTrigger>
                                        </FormControl>
                                        <SelectContent>
                                            <SelectItem value="postgresql">PostgreSQL</SelectItem>
                                            <SelectItem value="mysql">MySQL</SelectItem>
                                            <SelectItem value="sqlite">SQLite</SelectItem>
                                        </SelectContent>
                                    </Select>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        {/* Switches */}
                        <div className="space-y-3 pt-2">
                            <FormField
                                control={form.control}
                                name="allowAiAssistance"
                                render={({ field }) => (
                                    <FormItem className="flex items-center justify-between rounded-lg border p-3">
                                        <div className="space-y-0.5">
                                            <FormLabel className="text-base">Cho phép AI hỗ trợ</FormLabel>
                                            <FormDescription>
                                                Sinh viên có thể sử dụng AI assistant trong khi làm bài
                                            </FormDescription>
                                        </div>
                                        <FormControl>
                                            <Switch
                                                checked={field.value}
                                                onCheckedChange={field.onChange}
                                            />
                                        </FormControl>
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="shuffleProblems"
                                render={({ field }) => (
                                    <FormItem className="flex items-center justify-between rounded-lg border p-3">
                                        <div className="space-y-0.5">
                                            <FormLabel className="text-base">Xáo trộn câu hỏi</FormLabel>
                                            <FormDescription>
                                                Thứ tự bài tập sẽ random cho mỗi sinh viên
                                            </FormDescription>
                                        </div>
                                        <FormControl>
                                            <Switch
                                                checked={field.value}
                                                onCheckedChange={field.onChange}
                                            />
                                        </FormControl>
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="showResultImmediately"
                                render={({ field }) => (
                                    <FormItem className="flex items-center justify-between rounded-lg border p-3">
                                        <div className="space-y-0.5">
                                            <FormLabel className="text-base">Hiển thị kết quả ngay</FormLabel>
                                            <FormDescription>
                                                Sinh viên xem điểm ngay sau khi nộp bài
                                            </FormDescription>
                                        </div>
                                        <FormControl>
                                            <Switch
                                                checked={field.value}
                                                onCheckedChange={field.onChange}
                                            />
                                        </FormControl>
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="isPublic"
                                render={({ field }) => (
                                    <FormItem className="flex items-center justify-between rounded-lg border p-3">
                                        <div className="space-y-0.5">
                                            <FormLabel className="text-base">Công khai</FormLabel>
                                            <FormDescription>
                                                Tất cả sinh viên đều có thể thấy kỳ thi này
                                            </FormDescription>
                                        </div>
                                        <FormControl>
                                            <Switch
                                                checked={field.value}
                                                onCheckedChange={field.onChange}
                                            />
                                        </FormControl>
                                    </FormItem>
                                )}
                            />
                        </div>

                        <DialogFooter>
                            <Button
                                type="button"
                                variant="outline"
                                onClick={() => setOpen(false)}
                                disabled={isSubmitting}
                            >
                                Hủy
                            </Button>
                            <Button type="submit" disabled={isSubmitting}>
                                {isSubmitting ? (
                                    <>
                                        <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                                        Đang tạo...
                                    </>
                                ) : (
                                    'Tạo kỳ thi'
                                )}
                            </Button>
                        </DialogFooter>
                    </form>
                </Form>
            </DialogContent>
        </Dialog>
    )
}
