import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import toast from 'react-hot-toast'
import { useQuery } from '@tanstack/react-query'

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
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { examsService } from '@/services/exams.service'
import { problemsService } from '@/services/problems.service'
import type { AddExamProblemRequest, Problem } from '@/types/exam.types'
import { Plus, Loader2 } from 'lucide-react'

const problemSchema = z.object({
    problemId: z.number().min(1, 'Vui lòng chọn bài tập'),
    points: z.number().min(0, 'Điểm phải lớn hơn hoặc bằng 0').max(100, 'Điểm tối đa là 100'),
    sortOrder: z.number().min(1, 'Thứ tự phải lớn hơn 0'),
})

type ProblemFormValues = z.infer<typeof problemSchema>

interface AddProblemsDialogProps {
    examId: number
    currentProblemsCount: number
    onSuccess?: () => void
}

export function AddProblemsDialog({ examId, currentProblemsCount, onSuccess }: AddProblemsDialogProps) {
    const [open, setOpen] = useState(false)
    const [isSubmitting, setIsSubmitting] = useState(false)

    // Fetch all problems
    const { data: problemsData = [], isLoading } = useQuery({
        queryKey: ['problems'],
        queryFn: () => problemsService.list({}),
    })
    const problems = Array.isArray(problemsData) ? problemsData : []

    const form = useForm<ProblemFormValues>({
        resolver: zodResolver(problemSchema),
        defaultValues: {
            problemId: 0,
            points: 10,
            sortOrder: currentProblemsCount + 1,
        },
    })

    const selectedProblemId = form.watch('problemId')
    const selectedProblem = problems.find((p) => p.id === selectedProblemId)

    const onSubmit = async (values: ProblemFormValues) => {
        setIsSubmitting(true)

        try {
            const request: AddExamProblemRequest = {
                problemId: values.problemId,
                points: values.points,
                sortOrder: values.sortOrder,
            }

            await examsService.addProblem(examId, request)
            toast.success('Thêm bài tập thành công!')
            setOpen(false)
            form.reset({
                problemId: 0,
                points: 10,
                sortOrder: currentProblemsCount + 2,
            })
            onSuccess?.()
        } catch (error: any) {
            toast.error(error?.message || 'Thêm bài tập thất bại!')
        } finally {
            setIsSubmitting(false)
        }
    }

    const getDifficultyBadge = (difficulty: Problem['difficulty']) => {
        const config = {
            easy: { label: 'Dễ', className: 'bg-green-500/10 text-green-600' },
            medium: { label: 'Trung bình', className: 'bg-yellow-500/10 text-yellow-600' },
            hard: { label: 'Khó', className: 'bg-red-500/10 text-red-600' },
        }
        return (
            <Badge variant="secondary" className={config[difficulty].className}>
                {config[difficulty].label}
            </Badge>
        )
    }

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger asChild>
                <Button size="sm">
                    <Plus className="h-4 w-4 mr-1" />
                    Thêm bài tập
                </Button>
            </DialogTrigger>
            <DialogContent className="max-w-lg">
                <DialogHeader>
                    <DialogTitle>Thêm bài tập vào kỳ thi</DialogTitle>
                    <DialogDescription>
                        Chọn bài tập và cấu hình điểm số
                    </DialogDescription>
                </DialogHeader>

                <Form {...form}>
                    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                        {/* Problem Selection */}
                        <FormField
                            control={form.control}
                            name="problemId"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Bài tập *</FormLabel>
                                    <Select
                                        value={field.value.toString()}
                                        onValueChange={(value) => field.onChange(parseInt(value))}
                                        disabled={isLoading}
                                    >
                                        <FormControl>
                                            <SelectTrigger>
                                                <SelectValue placeholder="Chọn bài tập..." />
                                            </SelectTrigger>
                                        </FormControl>
                                        <SelectContent>
                                            {problems.map((problem) => (
                                                <SelectItem key={problem.id} value={problem.id.toString()}>
                                                    <div className="flex items-center gap-2">
                                                        <span>{problem.title}</span>
                                                    </div>
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        {/* Show selected problem info */}
                        {selectedProblem && (
                            <div className="p-3 border rounded-lg bg-muted/50">
                                <div className="flex items-center gap-2 mb-2">
                                    <span className="font-medium">{selectedProblem.title}</span>
                                    {getDifficultyBadge(selectedProblem.difficulty)}
                                </div>
                                <p className="text-sm text-muted-foreground line-clamp-2">
                                    {selectedProblem.description}
                                </p>
                            </div>
                        )}

                        {/* Points & Sort Order */}
                        <div className="grid grid-cols-2 gap-4">
                            <FormField
                                control={form.control}
                                name="points"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Điểm *</FormLabel>
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
                                name="sortOrder"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Thứ tự *</FormLabel>
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

                        <DialogFooter>
                            <Button
                                type="button"
                                variant="outline"
                                onClick={() => setOpen(false)}
                                disabled={isSubmitting}
                            >
                                Hủy
                            </Button>
                            <Button type="submit" disabled={isSubmitting || !selectedProblem}>
                                {isSubmitting ? (
                                    <>
                                        <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                                        Đang thêm...
                                    </>
                                ) : (
                                    'Thêm bài tập'
                                )}
                            </Button>
                        </DialogFooter>
                    </form>
                </Form>
            </DialogContent>
        </Dialog>
    )
}
