import { useQuery, useQueryClient } from '@tanstack/react-query'
import { createFileRoute } from '@tanstack/react-router'
import { useNavigate } from '@tanstack/react-router'
import {
    Calendar,
    Clock,
    Users,
    FileText,
    Loader2,
    ArrowLeft,
    Settings,
} from 'lucide-react'
import { useState } from 'react'
import toast from 'react-hot-toast'

import { PDFImportWizard } from '@/components/exam/pdf-import-wizard'
import { AddParticipantsDialog } from '@/components/exams/add-participants-dialog'
import { AddProblemsDialog } from '@/components/exams/add-problems-dialog'
import { CreateExamDialog } from '@/components/exams/create-exam-dialog'
import { MainLayout } from '@/components/layouts/main-layout'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { examsService } from '@/services/exams.service'
import { useAuthStore } from '@/stores/use-auth-store'
import type { Exam } from '@/types/exam.types'

function ExamsPage() {
    const navigate = useNavigate()
    const queryClient = useQueryClient()
    const { isOperator, userRole } = useAuthStore()
    const isLecturer = isOperator() || userRole === 'lecturer'

    const [selectedExam, setSelectedExam] = useState<Exam | null>(null)
    const [showPDFImport, setShowPDFImport] = useState(false)

    const { data: selectedExamDetails } = useQuery({
        queryKey: ['exam', selectedExam?.id],
        queryFn: () => examsService.getById(selectedExam!.id),
        enabled: !!selectedExam,
    })

    const { data: selectedExamParticipants = [] } = useQuery({
        queryKey: ['exam-participants', selectedExam?.id],
        queryFn: () => examsService.listParticipants(selectedExam!.id),
        enabled: !!selectedExam,
    })

    // Fetch exams list
    const { data: exams = [], isLoading } = useQuery({
        queryKey: ['exams'],
        queryFn: () => examsService.list(),
    })

    // RESTRICT: Only lecturers can access
    if (!isLecturer) {
        toast.error('Bạn không có quyền truy cập trang này!', { id: 'exams-access-denied' })
        navigate({ to: '/practice' })
        return null
    }

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
        const examDetail = selectedExamDetails ?? selectedExam

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
                            <h1 className="text-3xl font-bold">{examDetail.title}</h1>
                            <p className="text-muted-foreground">{examDetail.description}</p>
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
                                    <p className="font-medium">{formatDateTime(examDetail.startTime)}</p>
                                </div>
                                <div>
                                    <p className="text-sm text-muted-foreground">Thời gian kết thúc</p>
                                    <p className="font-medium">{formatDateTime(examDetail.endTime)}</p>
                                </div>
                                <div>
                                    <p className="text-sm text-muted-foreground">Thời lượng</p>
                                    <p className="font-medium">{examDetail.durationMinutes} phút</p>
                                </div>
                                <div>
                                    <p className="text-sm text-muted-foreground">Số lần thi tối đa</p>
                                    <p className="font-medium">{examDetail.maxAttempts}</p>
                                </div>
                            </CardContent>
                        </Card>

                        {/* Problems Section */}
                        <Card className="lg:col-span-2">
                            <CardHeader>
                                <div className="flex items-center justify-between">
                                    <CardTitle className="flex items-center gap-2">
                                        <FileText className="h-5 w-5" />
                                        Bài tập ({examDetail.problems?.length || 0})
                                    </CardTitle>
                                    <AddProblemsDialog
                                        examId={examDetail.id}
                                        currentProblemsCount={examDetail.problems?.length || 0}
                                        onSuccess={() => {
                                            queryClient.invalidateQueries({ queryKey: ['exam', examDetail.id] })
                                            queryClient.invalidateQueries({ queryKey: ['exams'] })
                                        }}
                                    />
                                </div>
                            </CardHeader>
                            <CardContent>
                                {examDetail.problems && examDetail.problems.length > 0 ? (
                                    <div className="space-y-2">
                                        {examDetail.problems.map((examProblem) => (
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
                                        examId={examDetail.id}
                                        onSuccess={() => {
                                            queryClient.invalidateQueries({ queryKey: ['exam', examDetail.id] })
                                            queryClient.invalidateQueries({ queryKey: ['exam-participants', examDetail.id] })
                                            queryClient.invalidateQueries({ queryKey: ['exams'] })
                                        }}
                                    />
                                </div>
                            </CardHeader>
                            <CardContent>
                                {selectedExamParticipants.length > 0 ? (
                                    <div className="space-y-2">
                                        {selectedExamParticipants.map((participant) => (
                                            <div
                                                key={participant.id}
                                                className="flex items-center justify-between p-3 border rounded-lg"
                                            >
                                                <div className="flex-1 min-w-0">
                                                    <p className="font-medium truncate">{participant.fullName || `User ${participant.userId}`}</p>
                                                    <p className="text-sm text-muted-foreground truncate">{participant.email}</p>
                                                </div>
                                                <span className="text-xs text-muted-foreground">{participant.status}</span>
                                            </div>
                                        ))}
                                    </div>
                                ) : (
                                    <p className="text-center text-muted-foreground py-8">
                                        Chưa có sinh viên nào
                                    </p>
                                )}
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
                    <div className="flex gap-2">
                        <Button variant="outline" onClick={() => setShowPDFImport(true)}>
                            <FileText className="h-4 w-4 mr-2" />
                            Import PDF
                        </Button>
                        <CreateExamDialog
                            onSuccess={() => {
                                queryClient.invalidateQueries({ queryKey: ['exams'] })
                            }}
                        />
                    </div>
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

                {/* PDF Import Wizard */}
                <PDFImportWizard
                    open={showPDFImport}
                    onOpenChange={setShowPDFImport}
                    onSuccess={() => {
                        toast.success('Import câu hỏi thành công!')
                        setShowPDFImport(false)
                    }}
                />
            </div>
        </MainLayout>
    )
}

export const Route = createFileRoute('/exams')({
    component: ExamsPage,
})
