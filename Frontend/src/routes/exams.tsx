import { createFileRoute } from '@tanstack/react-router'
import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { useNavigate } from '@tanstack/react-router'

import { MainLayout } from '@/components/layouts/main-layout'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { useAuthStore } from '@/stores/use-auth-store'
import { examsService } from '@/services/exams.service'
import type { Exam } from '@/types/exam.types'
import {
    Calendar,
    Clock,
    Users,
    FileText,
    Loader2,
    ArrowLeft,
    Settings,
} from 'lucide-react'
import { CreateExamDialog } from '@/components/exams/create-exam-dialog'
import { AddProblemsDialog } from '@/components/exams/add-problems-dialog'
import { AddParticipantsDialog } from '@/components/exams/add-participants-dialog'

function ExamsPage() {
    const navigate = useNavigate()
    const queryClient = useQueryClient()
    const { isOperator, userRole } = useAuthStore()
    const isLecturer = isOperator() || userRole === 'lecturer'

    // RESTRICT: Only lecturers can access
    if (!isLecturer) {
        toast.error('Bạn không có quyền truy cập trang này!', { id: 'exams-access-denied' })
        navigate({ to: '/practice' as any })
        return null
    }

    const [selectedExam, setSelectedExam] = useState<Exam | null>(null)

    // Fetch exams list
    const { data: exams = [], isLoading, refetch } = useQuery({
        queryKey: ['exams'],
        queryFn: () => examsService.list(),
    })

    const handleBackToList = () => {
        setSelectedExam(null)
    }

    const handleSelectExam = (exam: Exam) => {
        setSelectedExam(exam)
    }

    const formatDateTime = (dateString: string) => {
        return new Date(dateString).toLocaleString('vi-VN', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
        })
    }

    // Exam Detail View
    if (selectedExam) {
        return (
            <MainLayout>
                <div className="space-y-6">
                    {/* Header */}
                    <div className="flex items-center gap-4">
                        <Button variant="ghost" size="sm" onClick={handleBackToList}>
                            <ArrowLeft className="h-4 w-4 mr-1" />
                            Quay lại
                        </Button>
                        <div className="flex-1">
                            <h1 className="text-3xl font-bold">{selectedExam.title}</h1>
                            <p className="text-muted-foreground">{selectedExam.description}</p>
                        </div>
                    </div>

                    <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                        {/* Exam Info */}
                        <Card className="lg:col-span-3">
                            <CardHeader>
                                <CardTitle className="flex items-center gap-2">
                                    <Settings className="h-5 w-5" />
                                    Thông tin kỳ thi
                                </CardTitle>
                            </CardHeader>
                            <CardContent className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                                <div>
                                    <p className="text-sm text-muted-foreground">Thời gian bắt đầu</p>
                                    <p className="font-medium">{formatDateTime(selectedExam.startTime)}</p>
                                </div>
                                <div>
                                    <p className="text-sm text-muted-foreground">Thời gian kết thúc</p>
                                    <p className="font-medium">{formatDateTime(selectedExam.endTime)}</p>
                                </div>
                                <div>
                                    <p className="text-sm text-muted-foreground">Thời lượng</p>
                                    <p className="font-medium">{selectedExam.durationMinutes} phút</p>
                                </div>
                                <div>
                                    <p className="text-sm text-muted-foreground">Số lần thi tối đa</p>
                                    <p className="font-medium">{selectedExam.maxAttempts}</p>
                                </div>
                            </CardContent>
                        </Card>

                        {/* Problems Section */}
                        <Card className="lg:col-span-2">
                            <CardHeader>
                                <div className="flex items-center justify-between">
                                    <CardTitle className="flex items-center gap-2">
                                        <FileText className="h-5 w-5" />
                                        Bài tập ({selectedExam.problems?.length || 0})
                                    </CardTitle>
                                    <AddProblemsDialog
                                        examId={selectedExam.id}
                                        currentProblemsCount={selectedExam.problems?.length || 0}
                                        onSuccess={() => refetch()}
                                    />
                                </div>
                            </CardHeader>
                            <CardContent>
                                {selectedExam.problems && selectedExam.problems.length > 0 ? (
                                    <div className="space-y-2">
                                        {selectedExam.problems.map((examProblem) => (
                                            <div
                                                key={examProblem.id}
                                                className="flex items-center justify-between p-3 border rounded-lg"
                                            >
                                                <div className="flex-1">
                                                    <p className="font-medium">
                                                        {examProblem.problem?.title || `Problem ${examProblem.problemId}`}
                                                    </p>
                                                    <p className="text-sm text-muted-foreground">
                                                        {examProblem.points} điểm
                                                    </p>
                                                </div>
                                                <Button variant="ghost" size="sm" className="text-red-600">
                                                    Xóa
                                                </Button>
                                            </div>
                                        ))}
                                    </div>
                                ) : (
                                    <p className="text-center text-muted-foreground py-8">
                                        Chưa có bài tập nào. Nhấn "Thêm bài tập" để bắt đầu.
                                    </p>
                                )}
                            </CardContent>
                        </Card>

                        {/* Participants Section */}
                        <Card>
                            <CardHeader>
                                <div className="flex items-center justify-between">
                                    <CardTitle className="flex items-center gap-2">
                                        <Users className="h-5 w-5" />
                                        Sinh viên
                                    </CardTitle>
                                    <AddParticipantsDialog
                                        examId={selectedExam.id}
                                        onSuccess={() => refetch()}
                                    />
                                </div>
                            </CardHeader>
                            <CardContent>
                                <p className="text-center text-muted-foreground py-8">
                                    Chưa có sinh viên nào
                                </p>
                            </CardContent>
                        </Card>
                    </div>
                </div>
            </MainLayout>
        )
    }

    // Exams List View
    return (
        <MainLayout>
            <div className="space-y-6">
                {/* Header */}
                <div className="flex items-center justify-between">
                    <div>
                        <h1 className="text-3xl font-bold">Quản lý kỳ thi</h1>
                        <p className="text-muted-foreground">
                            Tạo và quản lý kỳ thi cho sinh viên
                        </p>
                    </div>
                    <CreateExamDialog
                        onSuccess={() => {
                            queryClient.invalidateQueries({ queryKey: ['exams'] })
                        }}
                    />
                </div>

                {/* Exams Grid */}
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {isLoading ? (
                        <div className="col-span-full flex items-center justify-center py-12">
                            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                        </div>
                    ) : exams.length === 0 ? (
                        <div className="col-span-full text-center py-12">
                            <p className="text-muted-foreground mb-4">
                                Chưa có kỳ thi nào. Hãy tạo kỳ thi đầu tiên!
                            </p>
                            <CreateExamDialog
                                onSuccess={() => {
                                    queryClient.invalidateQueries({ queryKey: ['exams'] })
                                }}
                            />
                        </div>
                    ) : (
                        exams.map((exam) => (
                            <Card
                                key={exam.id}
                                className="cursor-pointer hover:shadow-md transition-shadow"
                                onClick={() => handleSelectExam(exam)}
                            >
                                <CardHeader>
                                    <div className="flex items-start justify-between">
                                        <CardTitle className="text-lg">{exam.title}</CardTitle>
                                        {exam.isPublic && (
                                            <Badge variant="secondary">Public</Badge>
                                        )}
                                    </div>
                                    {exam.description && (
                                        <CardDescription className="line-clamp-2">
                                            {exam.description}
                                        </CardDescription>
                                    )}
                                </CardHeader>
                                <CardContent className="space-y-2">
                                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                                        <Calendar className="h-4 w-4" />
                                        <span>{formatDateTime(exam.startTime)}</span>
                                    </div>
                                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                                        <Clock className="h-4 w-4" />
                                        <span>{exam.durationMinutes} phút</span>
                                    </div>
                                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                                        <FileText className="h-4 w-4" />
                                        <span>{exam.problemCount ?? exam.problems?.length ?? 0} bài tập</span>
                                    </div>
                                </CardContent>
                            </Card>
                        ))
                    )}
                </div>
            </div>
        </MainLayout>
    )
}

export const Route = createFileRoute('/exams')({
    component: ExamsPage,
})
