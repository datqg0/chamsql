import { createFileRoute } from '@tanstack/react-router'
import { useState, useEffect } from 'react'
import { useNavigate } from '@tanstack/react-router'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import rehypeRaw from 'rehype-raw'
import toast from 'react-hot-toast'

import { SQLEditor } from '@/components/editor/sql-editor'
import { MainLayout } from '@/components/layouts/main-layout'
import { ExamImportDialog } from '@/components/exam/exam-import-dialog'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
    ResizablePanelGroup,
    ResizablePanel,
    ResizableHandle,
} from '@/components/ui/resizable'
import { useWebSocket } from '@/hooks/use-websocket'
import { useSQLChecker } from '@/hooks/use-sql-checker'
import { useAuthStore } from '@/stores/use-auth-store'
import { examsService } from '@/services/exams.service'
import { submissionsService } from '@/services/submissions.service'
import type { MyExam, Exam, Submission } from '@/types/exam.types'
import {
    Play,
    Send,
    Loader2,
    Wifi,
    WifiOff,
    CheckCircle2,
    Clock,
    FileText,
    Code,
    ArrowLeft,
    Calendar,
    Timer,
    Trophy,
    AlertCircle,
} from 'lucide-react'
import { cn } from '@/lib/utils'

function SubmissionsPage() {
    const navigate = useNavigate()
    const { isOperator, userRole } = useAuthStore()
    // Kiểm tra quyền giảng viên (Operator hoặc Lecturer)
    const isLecturer = isOperator() || userRole === 'lecturer'

    // RESTRICT: Only students and admins can access this page
    useEffect(() => {
        if (isLecturer && userRole !== 'admin') {
            toast.error('Trang này chỉ dành cho sinh viên!', { id: 'submissions-access-denied' })
            navigate({ to: '/exams' as any })
        }
    }, [isLecturer, navigate, userRole])

    // State
    const [myExams, setMyExams] = useState<MyExam[]>([])
    const [submissionHistory, setSubmissionHistory] = useState<Submission[]>([])
    const [isLoadingExams, setIsLoadingExams] = useState(true)
    const [isLoadingHistory, setIsLoadingHistory] = useState(false)
    const [selectedExam, setSelectedExam] = useState<MyExam | null>(null)
    const [examProblems, setExamProblems] = useState<any[]>([])
    const [currentProblemIndex, setCurrentProblemIndex] = useState(0)
    const [sqlQuery, setSqlQuery] = useState('')
    const [isRunning, setIsRunning] = useState(false)
    const [isSubmitting, setIsSubmitting] = useState(false)
    const [result, setResult] = useState<any>(null)
    const [mobileView, setMobileView] = useState<'problem' | 'editor'>('problem')
    const [answers, setAnswers] = useState<Record<number, string>>({})

    const { isValid, syntaxError } = useSQLChecker(sqlQuery, {
        debounceMs: 300,
        database: 'MySQL',
    })

    const { send, isConnected } = useWebSocket({
        onMessage: (message) => {
            if (message.type === 'sql_result') {
                setResult(message.data)
                setIsRunning(false)
            }
        },
    })

    // Load exams from API
    useEffect(() => {
        loadMyExams()
        loadSubmissionHistory()
    }, [])

    const loadMyExams = async () => {
        setIsLoadingExams(true)
        try {
            const data = await examsService.getMyExams()
            setMyExams(Array.isArray(data) ? data : [])
        } catch (error) {
            console.error('Error loading exams:', error)
            toast.error('Không thể tải danh sách đề thi')
            setMyExams([])
        } finally {
            setIsLoadingExams(false)
        }
    }

    const loadSubmissionHistory = async () => {
        setIsLoadingHistory(true)
        try {
            const response = await submissionsService.list({ page: 1, pageSize: 20 })
            const data = response as any
            // Handle various response structures
            if (data && data.data && Array.isArray(data.data.submissions)) {
                setSubmissionHistory(data.data.submissions)
            } else if (data && data.data && Array.isArray(data.data)) {
                setSubmissionHistory(data.data)
            } else if (Array.isArray(data)) {
                setSubmissionHistory(data)
            } else {
                setSubmissionHistory([])
            }
        } catch (error) {
            console.error('Error loading history:', error)
        } finally {
            setIsLoadingHistory(false)
        }
    }

    const getStatusBadge = (status: MyExam['status']) => {
        switch (status) {
            case 'not_started':
                return <Badge variant="secondary">Chưa bắt đầu</Badge>
            case 'in_progress':
                return <Badge className="bg-yellow-500">Đang làm</Badge>
            case 'finished':
                return <Badge className="bg-green-500">Đã hoàn thành</Badge>
        }
    }

    const formatDate = (dateString: string) => {
        return new Date(dateString).toLocaleString('vi-VN', {
            day: '2-digit',
            month: '2-digit',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
        })
    }

    const canStartExam = (exam: Exam) => {
        const now = new Date()
        const start = new Date(exam.startTime)
        const end = new Date(exam.endTime)
        return now >= start && now <= end
    }

    const handleStartExam = async (myExam: MyExam) => {
        if (!canStartExam(myExam.exam)) {
            toast.error('Chưa đến thời gian làm bài hoặc đã hết hạn')
            return
        }

        try {
            await examsService.start(myExam.exam.id)
            setSelectedExam(myExam)
            setExamProblems(myExam.exam.problems || [])
            setCurrentProblemIndex(0)
            setSqlQuery('')
            setResult(null)
            setAnswers({})
        } catch (error: any) {
            toast.error(error?.message || 'Không thể bắt đầu bài thi')
        }
    }

    const handleViewResult = (myExam: MyExam) => {
        setSelectedExam(myExam)
        setExamProblems(myExam.exam.problems || [])
    }

    const handleBackToList = () => {
        setSelectedExam(null)
        setExamProblems([])
        setCurrentProblemIndex(0)
        setSqlQuery('')
        setResult(null)
        loadMyExams() // Refresh list
    }

    const handleRunTest = async () => {
        if (!selectedExam) return

        setIsRunning(true)
        setResult(null)

        if (isConnected) {
            send({
                type: 'run_sql_test',
                data: { query: sqlQuery },
            })
        } else {
            // Fallback - use problems API
            try {
                const currentProblem = examProblems[currentProblemIndex]
                if (currentProblem?.problem?.id) {
                    const { problemsService } = await import('@/services/problems.service')
                    const response = await problemsService.run(currentProblem.problem.id, {
                        code: sqlQuery,
                        databaseType: selectedExam.exam.allowedDatabases?.[0] || 'postgresql',
                    })
                    setResult(response)
                }
            } catch (error: any) {
                setResult({ success: false, error: error?.message || 'Lỗi khi chạy truy vấn' })
            } finally {
                setIsRunning(false)
            }
        }
    }

    const handleSubmitAnswer = async () => {
        if (!isValid || !selectedExam) {
            toast.error('Vui lòng sửa lỗi cú pháp!')
            return
        }

        const currentProblem = examProblems[currentProblemIndex]
        if (!currentProblem) return

        setIsSubmitting(true)

        try {
            await examsService.submitAnswer(selectedExam.exam.id, {
                problemId: currentProblem.problemId || currentProblem.problem?.id,
                code: sqlQuery,
                databaseType: selectedExam.exam.allowedDatabases?.[0] || 'postgresql',
            })

            // Save answer locally
            setAnswers((prev) => ({
                ...prev,
                [currentProblem.id]: sqlQuery,
            }))

            toast.success('Đã lưu câu trả lời!')

            // Move to next problem if available
            if (currentProblemIndex < examProblems.length - 1) {
                setCurrentProblemIndex(currentProblemIndex + 1)
                setSqlQuery(answers[examProblems[currentProblemIndex + 1]?.id] || '')
                setResult(null)
            }
        } catch (error: any) {
            toast.error(error?.message || 'Không thể lưu câu trả lời')
        } finally {
            setIsSubmitting(false)
        }
    }

    const handleFinishExam = async () => {
        if (!selectedExam) return

        try {
            await examsService.finish(selectedExam.exam.id)
            toast.success('Đã nộp bài thành công!')
            toast.success('Đã nộp bài thành công!')
            handleBackToList()
            loadSubmissionHistory() // Refresh history after finish
        } catch (error: any) {
            toast.error(error?.message || 'Không thể nộp bài')
        }
    }

    const handleProblemSelect = (index: number) => {
        // Save current answer
        const currentProblem = examProblems[currentProblemIndex]
        if (currentProblem) {
            setAnswers((prev) => ({
                ...prev,
                [currentProblem.id]: sqlQuery,
            }))
        }

        // Load new problem
        setCurrentProblemIndex(index)
        setSqlQuery(answers[examProblems[index]?.id] || '')
        setResult(null)
    }

    // Exam Taking View
    if (selectedExam && selectedExam.status !== 'finished') {
        const currentProblem = examProblems[currentProblemIndex]

        return (
            <MainLayout>
                <div className="h-[calc(100vh-120px)] flex flex-col">
                    {/* Header */}
                    <div className="flex items-center gap-4 mb-4">
                        <Button variant="ghost" size="sm" onClick={handleBackToList}>
                            <ArrowLeft className="h-4 w-4 mr-1" />
                            Thoát
                        </Button>
                        <div className="flex-1">
                            <h1 className="text-xl font-bold">{selectedExam.exam.title}</h1>
                            <p className="text-sm text-muted-foreground">
                                Thời gian: {selectedExam.exam.durationMinutes} phút
                            </p>
                        </div>
                        <Button onClick={handleFinishExam}>
                            <CheckCircle2 className="h-4 w-4 mr-2" />
                            Nộp bài
                        </Button>
                    </div>

                    {/* Problem Navigation */}
                    {examProblems.length > 0 && (
                        <div className="flex gap-2 mb-4 flex-wrap">
                            {examProblems.map((problem, index) => (
                                <Button
                                    key={problem.id}
                                    variant={currentProblemIndex === index ? 'default' : 'outline'}
                                    size="sm"
                                    onClick={() => handleProblemSelect(index)}
                                    className={cn(
                                        answers[problem.id] && 'ring-2 ring-green-500'
                                    )}
                                >
                                    Câu {index + 1}
                                </Button>
                            ))}
                        </div>
                    )}

                    {/* Mobile View Toggle */}
                    <div className="lg:hidden flex gap-2 mb-4">
                        <Button
                            variant={mobileView === 'problem' ? 'default' : 'outline'}
                            size="sm"
                            onClick={() => setMobileView('problem')}
                            className="flex-1"
                        >
                            <FileText className="h-4 w-4 mr-2" />
                            Đề bài
                        </Button>
                        <Button
                            variant={mobileView === 'editor' ? 'default' : 'outline'}
                            size="sm"
                            onClick={() => setMobileView('editor')}
                            className="flex-1"
                        >
                            <Code className="h-4 w-4 mr-2" />
                            Viết code
                        </Button>
                    </div>

                    {/* Desktop Layout */}
                    <div className="hidden lg:block flex-1">
                        <ResizablePanelGroup direction="horizontal" className="h-full rounded-lg border">
                            {/* Problem Description */}
                            <ResizablePanel defaultSize={50} minSize={30}>
                                <Card className="h-full flex flex-col overflow-hidden rounded-none border-0">
                                    <CardHeader className="pb-2">
                                        <CardTitle className="text-base flex items-center justify-between">
                                            {currentProblem?.problem?.title || `Câu ${currentProblemIndex + 1}`}
                                            <Badge variant="outline">{currentProblem?.points || 0} điểm</Badge>
                                        </CardTitle>
                                    </CardHeader>
                                    <CardContent className="flex-1 overflow-y-auto">
                                        <div className="prose prose-sm dark:prose-invert max-w-none">
                                            <ReactMarkdown remarkPlugins={[remarkGfm]}>
                                                {currentProblem?.problem?.description || 'Không có mô tả'}
                                            </ReactMarkdown>
                                        </div>
                                    </CardContent>
                                </Card>
                            </ResizablePanel>

                            <ResizableHandle withHandle />

                            {/* SQL Editor */}
                            <ResizablePanel defaultSize={50} minSize={30}>
                                <Card className="h-full flex flex-col overflow-hidden rounded-none border-0">
                                    <CardHeader className="pb-2">
                                        <CardTitle className="text-base flex items-center gap-2">
                                            <Code className="h-4 w-4" />
                                            Viết câu truy vấn
                                        </CardTitle>
                                    </CardHeader>
                                    <CardContent className="flex-1 flex flex-col gap-4 overflow-y-auto">
                                        <div className="min-h-[300px]">
                                            <SQLEditor
                                                value={sqlQuery}
                                                onChange={(v) => setSqlQuery(v || '')}
                                                height="300px"
                                                syntaxError={syntaxError}
                                            />
                                        </div>

                                        {!isValid && syntaxError && (
                                            <div className="p-2 bg-red-500/10 text-red-600 rounded text-xs">
                                                {syntaxError.message}
                                            </div>
                                        )}

                                        {result && (
                                            <div className="p-3 bg-muted rounded text-sm">
                                                <p className="font-medium mb-2">Kết quả:</p>
                                                {result.success ? (
                                                    <pre className="text-xs overflow-x-auto">
                                                        {JSON.stringify(result.data, null, 2)}
                                                    </pre>
                                                ) : (
                                                    <p className="text-red-500">{result.error}</p>
                                                )}
                                            </div>
                                        )}

                                        <div className="flex gap-2">
                                            <Button
                                                variant="outline"
                                                onClick={handleRunTest}
                                                disabled={isRunning || isSubmitting}
                                                className="flex-1"
                                            >
                                                {isRunning ? (
                                                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                                                ) : (
                                                    <Play className="h-4 w-4 mr-2" />
                                                )}
                                                Chạy thử
                                            </Button>
                                            <Button
                                                onClick={handleSubmitAnswer}
                                                disabled={isRunning || isSubmitting || !isValid}
                                                className="flex-1"
                                            >
                                                {isSubmitting ? (
                                                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                                                ) : (
                                                    <Send className="h-4 w-4 mr-2" />
                                                )}
                                                Lưu câu trả lời
                                            </Button>
                                        </div>
                                    </CardContent>
                                </Card>
                            </ResizablePanel>
                        </ResizablePanelGroup>
                    </div>

                    {/* Mobile Layout */}
                    <div className="lg:hidden flex-1 overflow-hidden">
                        {mobileView === 'problem' ? (
                            <Card className="h-full flex flex-col overflow-hidden">
                                <CardHeader className="pb-2">
                                    <CardTitle className="text-base flex items-center justify-between">
                                        {currentProblem?.problem?.title || `Câu ${currentProblemIndex + 1}`}
                                        <Badge variant="outline">{currentProblem?.points || 0} điểm</Badge>
                                    </CardTitle>
                                </CardHeader>
                                <CardContent className="flex-1 overflow-y-auto">
                                    <div className="prose prose-sm dark:prose-invert max-w-none">
                                        <ReactMarkdown
                                            remarkPlugins={[remarkGfm]}
                                            rehypePlugins={[rehypeRaw]}
                                        >
                                            {currentProblem?.problem?.description || 'Không có mô tả'}
                                        </ReactMarkdown>
                                    </div>
                                </CardContent>
                            </Card>
                        ) : (
                            <Card className="h-full flex flex-col overflow-hidden">
                                <CardHeader className="pb-2">
                                    <CardTitle className="text-base">Viết câu truy vấn</CardTitle>
                                </CardHeader>
                                <CardContent className="flex-1 flex flex-col gap-3 overflow-y-auto">
                                    <div className="min-h-[200px]">
                                        <SQLEditor
                                            value={sqlQuery}
                                            onChange={(v) => setSqlQuery(v || '')}
                                            height="200px"
                                            syntaxError={syntaxError}
                                        />
                                    </div>
                                    <div className="flex gap-2">
                                        <Button
                                            variant="outline"
                                            onClick={handleRunTest}
                                            disabled={isRunning || isSubmitting}
                                            className="flex-1"
                                            size="sm"
                                        >
                                            {isRunning ? <Loader2 className="h-4 w-4 animate-spin" /> : <Play className="h-4 w-4" />}
                                            <span className="ml-1">Chạy</span>
                                        </Button>
                                        <Button
                                            onClick={handleSubmitAnswer}
                                            disabled={isRunning || isSubmitting || !isValid}
                                            className="flex-1"
                                            size="sm"
                                        >
                                            {isSubmitting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Send className="h-4 w-4" />}
                                            <span className="ml-1">Lưu</span>
                                        </Button>
                                    </div>
                                </CardContent>
                            </Card>
                        )}
                    </div>
                </div>
            </MainLayout>
        )
    }

    // View Result View (for finished exams)
    if (selectedExam && selectedExam.status === 'finished') {
        return (
            <MainLayout>
                <div className="space-y-6">
                    <div className="flex items-center gap-4">
                        <Button variant="ghost" size="sm" onClick={handleBackToList}>
                            <ArrowLeft className="h-4 w-4 mr-1" />
                            Quay lại
                        </Button>
                        <div className="flex-1">
                            <h1 className="text-2xl font-bold">{selectedExam.exam.title}</h1>
                            <p className="text-muted-foreground">Kết quả bài thi</p>
                        </div>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                        <Card>
                            <CardContent className="pt-6">
                                <div className="flex items-center gap-3">
                                    <Trophy className="h-8 w-8 text-yellow-500" />
                                    <div>
                                        <p className="text-3xl font-bold">
                                            {selectedExam.score ?? 0}/{selectedExam.totalPoints ?? 0}
                                        </p>
                                        <p className="text-sm text-muted-foreground">Điểm số</p>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                        <Card>
                            <CardContent className="pt-6">
                                <div className="flex items-center gap-3">
                                    <CheckCircle2 className="h-8 w-8 text-green-500" />
                                    <div>
                                        <p className="text-3xl font-bold">
                                            {selectedExam.totalPoints
                                                ? Math.round((selectedExam.score! / selectedExam.totalPoints) * 100)
                                                : 0}%
                                        </p>
                                        <p className="text-sm text-muted-foreground">Tỷ lệ đúng</p>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                        <Card>
                            <CardContent className="pt-6">
                                <div className="flex items-center gap-3">
                                    <Clock className="h-8 w-8 text-blue-500" />
                                    <div>
                                        <p className="text-lg font-bold">
                                            {selectedExam.finishedAt ? formatDate(selectedExam.finishedAt) : 'N/A'}
                                        </p>
                                        <p className="text-sm text-muted-foreground">Thời gian nộp</p>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                    </div>

                    <Card>
                        <CardHeader>
                            <CardTitle>Chi tiết bài làm</CardTitle>
                        </CardHeader>
                        <CardContent>
                            {examProblems.length === 0 ? (
                                <p className="text-center py-8 text-muted-foreground">
                                    Không có dữ liệu chi tiết
                                </p>
                            ) : (
                                <div className="space-y-4">
                                    {examProblems.map((problem, index) => (
                                        <div key={problem.id} className="border rounded-lg p-4">
                                            <div className="flex items-center justify-between mb-2">
                                                <span className="font-medium">
                                                    {problem.problem?.title || `Câu ${index + 1}`}
                                                </span>
                                                <Badge variant="outline">
                                                    {problem.points || 0} điểm
                                                </Badge>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </CardContent>
                    </Card>
                </div>
            </MainLayout>
        )
    }

    // Main Exam List View
    return (
        <MainLayout>
            <div className="space-y-6">
                {/* Header */}
                <div className="flex items-center justify-between">
                    <div>
                        <h1 className="text-3xl font-bold">Thi trực tuyến</h1>
                        <p className="text-muted-foreground">
                            {isLecturer
                                ? 'Quản lý và import đề thi'
                                : 'Xem danh sách đề thi và kết quả của bạn'}
                        </p>
                    </div>
                    <div className="flex items-center gap-2">
                        {isConnected ? (
                            <span className="flex items-center gap-2 text-green-600 dark:text-green-400">
                                <Wifi className="h-4 w-4" />
                                <span className="text-sm">Đã kết nối</span>
                            </span>
                        ) : (
                            <span className="flex items-center gap-2 text-muted-foreground">
                                <WifiOff className="h-4 w-4" />
                                <span className="text-sm">Chưa kết nối</span>
                            </span>
                        )}
                        {isLecturer && (
                            <ExamImportDialog
                                onSuccess={() => {
                                    toast.success('Đề thi đã được import!')
                                    loadMyExams()
                                }}
                            />
                        )}
                    </div>
                </div>

                {/* Loading State */}
                {isLoadingExams && (
                    <div className="flex items-center justify-center py-12">
                        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                    </div>
                )}

                {/* Tabs */}
                {!isLoadingExams && (
                    <Tabs defaultValue="available" className="w-full">
                        <TabsList>
                            <TabsTrigger value="available">Đề thi sắp tới</TabsTrigger>
                            <TabsTrigger value="history">Lịch sử nộp bài</TabsTrigger>
                        </TabsList>

                        <TabsContent value="available" className="mt-4">
                            <div className="grid gap-4">
                                {myExams
                                    .filter((e) => e.status !== 'finished')
                                    .map((myExam) => (
                                        <Card key={myExam.id} className="hover:shadow-md transition-shadow">
                                            <CardContent className="pt-6">
                                                <div className="flex items-start justify-between gap-4">
                                                    <div className="flex-1">
                                                        <div className="flex items-center gap-2 mb-2">
                                                            <h3 className="text-lg font-semibold">
                                                                {myExam.exam.title}
                                                            </h3>
                                                            {getStatusBadge(myExam.status)}
                                                        </div>
                                                        <p className="text-sm text-muted-foreground mb-3">
                                                            {myExam.exam.description}
                                                        </p>
                                                        <div className="flex flex-wrap gap-4 text-sm text-muted-foreground">
                                                            <span className="flex items-center gap-1">
                                                                <Calendar className="h-4 w-4" />
                                                                {formatDate(myExam.exam.startTime)}
                                                            </span>
                                                            <span className="flex items-center gap-1">
                                                                <Timer className="h-4 w-4" />
                                                                {myExam.exam.durationMinutes} phút
                                                            </span>
                                                            <span className="flex items-center gap-1">
                                                                <AlertCircle className="h-4 w-4" />
                                                                Tối đa {myExam.exam.maxAttempts} lần
                                                            </span>
                                                        </div>
                                                    </div>
                                                    <Button
                                                        onClick={() => handleStartExam(myExam)}
                                                        disabled={!canStartExam(myExam.exam)}
                                                    >
                                                        {canStartExam(myExam.exam) ? 'Bắt đầu làm bài' : 'Chưa mở'}
                                                    </Button>
                                                </div>
                                            </CardContent>
                                        </Card>
                                    ))}

                                {myExams.filter((e) => e.status !== 'finished').length === 0 && (
                                    <div className="text-center py-12 text-muted-foreground">
                                        <FileText className="h-12 w-12 mx-auto mb-4 opacity-50" />
                                        <p>Không có đề thi nào sắp tới</p>
                                    </div>
                                )}
                            </div>
                        </TabsContent>



                        <TabsContent value="history" className="mt-4">
                            {isLoadingHistory ? (
                                <div className="flex items-center justify-center py-8">
                                    <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                                </div>
                            ) : (
                                <div className="grid gap-4">
                                    {submissionHistory.map((sub) => (
                                        <Card key={sub.id} className="hover:shadow-md transition-shadow">
                                            <CardContent className="pt-6">
                                                <div className="flex items-start justify-between gap-4">
                                                    <div className="flex-1">
                                                        <div className="flex items-center gap-2 mb-2">
                                                            <h3 className="text-lg font-semibold">
                                                                {sub.examTitle || `Bài thi #${sub.examId}`}
                                                            </h3>
                                                            {sub.score !== undefined && (
                                                                <Badge variant={sub.score >= 50 ? 'secondary' : 'destructive'}>
                                                                    {sub.score >= 50 ? 'Đạt' : 'Chưa đạt'}
                                                                </Badge>
                                                            )}
                                                        </div>
                                                        <div className="flex flex-wrap gap-4 text-sm mt-3">
                                                            <span className="flex items-center gap-1 text-green-600 font-medium">
                                                                <Trophy className="h-4 w-4" />
                                                                Điểm: {sub.score}
                                                            </span>
                                                            {sub.createdAt && (
                                                                <span className="flex items-center gap-1 text-muted-foreground">
                                                                    <Clock className="h-4 w-4" />
                                                                    Nộp lúc: {formatDate(sub.createdAt)}
                                                                </span>
                                                            )}
                                                        </div>
                                                    </div>
                                                </div>
                                            </CardContent>
                                        </Card>
                                    ))}

                                    {submissionHistory.length === 0 && (
                                        <div className="text-center py-12 text-muted-foreground">
                                            <FileText className="h-12 w-12 mx-auto mb-4 opacity-50" />
                                            <p>Bạn chưa có lịch sử nộp bài nào</p>
                                        </div>
                                    )}
                                </div>
                            )}
                        </TabsContent>
                    </Tabs>
                )}
            </div>
        </MainLayout >
    )
}

export const Route = createFileRoute('/submissions')({
    component: SubmissionsPage,
})
