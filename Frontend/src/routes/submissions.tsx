import { createFileRoute } from '@tanstack/react-router'
import { useNavigate } from '@tanstack/react-router'
import {
    Play,
    Send,
    Loader2,
    CheckCircle2,
    Clock,
    FileText,
    Code,
    ArrowLeft,
    Calendar,
    Timer,
    Trophy,
    AlertCircle,
    RotateCw,
} from 'lucide-react'
import { useState, useEffect, useCallback } from 'react'
import toast from 'react-hot-toast'
import ReactMarkdown from 'react-markdown'
import rehypeRaw from 'rehype-raw'
import remarkGfm from 'remark-gfm'

import { SQLEditor } from '@/components/editor/sql-editor'
import { ExamAutoSubmitModal } from '@/components/exam/exam-auto-submit-modal'
import { ExamImportDialog } from '@/components/exam/exam-import-dialog'
import { ExamNavigation } from '@/components/exam/exam-navigation'
import { ExamResultsOverview } from '@/components/exam/exam-results-overview'
import { ExamTimer } from '@/components/exam/exam-timer'
import { ProblemHeader } from '@/components/exam/problem-header'
import { ProblemResultDetail } from '@/components/exam/problem-result-detail'
import { RetakeModal } from '@/components/exam/retake-modal'
import { MainLayout } from '@/components/layouts/main-layout'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
    ResizablePanelGroup,
    ResizablePanel,
    ResizableHandle,
} from '@/components/ui/resizable'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useSQLChecker } from '@/hooks/use-sql-checker'
import { examSubmissionService } from '@/services/exam-submission.service'
import { examsService } from '@/services/exams.service'
import { submissionsService } from '@/services/submissions.service'
import { useAuthStore } from '@/stores/use-auth-store'
import type { ProblemSubmissionResult } from '@/types/exam-submission.types'
import type { MyExam, Exam, Submission } from '@/types/exam.types'

function SubmissionsPage() {
    const navigate = useNavigate()
    const { isOperator, userRole } = useAuthStore()
    // Kiểm tra quyền giảng viên (Operator hoặc Lecturer)
    const isLecturer = isOperator() || userRole === 'lecturer'

    // RESTRICT: Only students and admins can access this page
    useEffect(() => {
        if (isLecturer && userRole !== 'admin') {
            toast.error('Trang này chỉ dành cho sinh viên!', { id: 'submissions-access-denied' })
            navigate({ to: '/exams' })
        }
    }, [isLecturer, navigate, userRole])

    // State
    const [myExams, setMyExams] = useState<MyExam[]>([])
    const [submissionHistory, setSubmissionHistory] = useState<Submission[]>([])
    const [isLoadingExams, setIsLoadingExams] = useState(true)
    const [isLoadingHistory, setIsLoadingHistory] = useState(false)
    const [selectedExam, setSelectedExam] = useState<MyExam | null>(null)
    const [examProblems, setExamProblems] = useState<Exam['problems']>([])
    const [currentProblemIndex, setCurrentProblemIndex] = useState(0)
    const [sqlQuery, setSqlQuery] = useState('')
    const [isRunning, setIsRunning] = useState(false)
    const [isSubmitting, setIsSubmitting] = useState(false)
    const [result, setResult] = useState<unknown>(null)
    const [mobileView, setMobileView] = useState<'problem' | 'editor'>('problem')
    const [answers, setAnswers] = useState<Record<number, string>>({})
    const [solvedProblems, setSolvedProblems] = useState<Record<number, ProblemSubmissionResult>>({})
    
    // Timer & Modal states
    const [examEndTimeMs, setExamEndTimeMs] = useState<number | null>(null)
    const [showAutoSubmitModal, setShowAutoSubmitModal] = useState(false)
    const [isAutoSubmitMode, setIsAutoSubmitMode] = useState(false)
    const [isFinalizingSubmit, setIsFinalizingSubmit] = useState(false)
    const [showRetakeModal, setShowRetakeModal] = useState(false)
    const [isRetakingExam, setIsRetakingExam] = useState(false)

    const { isValid, syntaxError } = useSQLChecker(sqlQuery, {
        debounceMs: 300,
        database: 'MySQL',
    })


    const loadMyExams = useCallback(async () => {
        setIsLoadingExams(true)
        try {
            if (isLecturer) {
                const exams = await examsService.list()
                const normalized: MyExam[] = (Array.isArray(exams) ? exams : []).map((exam) => ({
                    id: exam.id,
                    exam,
                    status: 'not_started',
                    score: 0,
                    totalPoints: exam.problemCount ?? 0,
                    attemptNumber: 0,
                }))
                setMyExams(normalized)
            } else {
                const data = await examsService.getMyExams()
                const normalized: MyExam[] = (Array.isArray(data) ? data : [])
                    .map((item: unknown) => {
                        const i = item as Record<string, unknown> & { exam?: unknown }
                        // Already in MyExam shape
                        if (i && i.exam) {
                            return i as unknown as MyExam
                        }

                        // Fallback: backend may return plain ExamResponse[]
                        const exam: Exam = {
                            id: Number(i?.id ?? 0),
                            title: String(i?.title ?? 'Kỳ thi'),
                            description: i?.description as string,
                            startTime: String(i?.startTime ?? i?.start_time ?? ''),
                            endTime: String(i?.endTime ?? i?.end_time ?? ''),
                            durationMinutes: Number(i?.durationMinutes ?? i?.duration_minutes ?? 0),
                            allowedDatabases: (i?.allowedDatabases ?? i?.allowed_databases ?? ['postgresql']) as string[],
                            allowAiAssistance: Boolean(i?.allowAiAssistance ?? i?.allow_ai_assistance),
                            shuffleProblems: Boolean(i?.shuffleProblems ?? i?.shuffle_problems),
                            showResultImmediately: Boolean(i?.showResultImmediately ?? i?.show_result_immediately),
                            maxAttempts: Number(i?.maxAttempts ?? i?.max_attempts ?? 1),
                            isPublic: Boolean(i?.isPublic ?? i?.is_public),
                            status: i?.status as string,
                            problemCount: i?.problemCount as number ?? i?.problem_count as number,
                        }

                        const rawStatus = String(i?.status ?? '').toLowerCase()
                        let status: MyExam['status'] = 'not_started'
                        if (rawStatus === 'in_progress' || rawStatus === 'ongoing') {
                            status = 'in_progress'
                        }
                        if (rawStatus === 'finished' || rawStatus === 'submitted' || rawStatus === 'graded' || rawStatus === 'completed') {
                            status = 'finished'
                        }

                        return {
                            id: exam.id,
                            exam,
                            status,
                            startedAt: i?.startedAt ?? i?.started_at,
                            finishedAt: i?.finishedAt ?? i?.finished_at,
                            score: i?.score,
                            totalPoints: i?.totalPoints ?? i?.total_points,
                            attemptNumber: i?.attemptNumber ?? i?.attempt_number,
                        } as unknown as MyExam
                    })
                    .filter((x) => x.exam?.id > 0)

                setMyExams(normalized)
            }
        } catch (error) {
            console.error('Error loading exams:', error)
            toast.error('Không thể tải danh sách đề thi')
            setMyExams([])
        } finally {
            setIsLoadingExams(false)
        }
    }, [isLecturer])

    const loadSubmissionHistory = useCallback(async () => {
        setIsLoadingHistory(true)
        try {
            const response = await submissionsService.list({ page: 1, pageSize: 20 })
            if (Array.isArray(response.data)) {
                setSubmissionHistory(response.data)
            } else {
                setSubmissionHistory([])
            }
        } catch (error) {
            console.error('Error loading history:', error)
        } finally {
            setIsLoadingHistory(false)
        }
    }, [])

    // Load exams from API
    useEffect(() => {
        loadMyExams()
        loadSubmissionHistory()
    }, [loadMyExams, loadSubmissionHistory])

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

    const getExamProblemKey = (problem: unknown): number => {
        return Number(
            problem?.examProblemID ??
                problem?.exam_problem_id ??
                problem?.id ??
                problem?.problemId ??
                problem?.problem_id ??
                0
        )
    }

    const handleStartExam = async (myExam: MyExam) => {
        if (!canStartExam(myExam.exam)) {
            toast.error('Chưa đến thời gian làm bài hoặc đã hết hạn')
            return
        }

        try {
            // Join exam first (ignore already joined errors)
            try {
                await examSubmissionService.joinExam(myExam.exam.id)
            } catch {
                // Ignore if already joined
            }

            // Start exam
            await examSubmissionService.startExam(myExam.exam.id)
            
            // Get time remaining
            const timeResponse = await examSubmissionService.getTimeRemaining(myExam.exam.id)
            
            // Set timer end time (current time + remaining ms)
            const endTime = Date.now() + timeResponse.timeRemainingMs
            setExamEndTimeMs(endTime)
            
            // Load exam problems
            const problemsResponse = await examSubmissionService.getExamWithProblems(myExam.exam.id)
            
            setSelectedExam(myExam)
            setExamProblems(problemsResponse.problems || [])
            setCurrentProblemIndex(0)
            setSqlQuery('')
            setResult(null)
            setAnswers({})
            setSolvedProblems({})
            
            toast.success('Đã bắt đầu bài thi!')
        } catch (error: unknown) {
            console.error('Error starting exam:', error)
            toast.error(error?.message || 'Không thể bắt đầu bài thi')
        }
    }

    const handleBackToList = () => {
        setSelectedExam(null)
        setExamProblems([])
        setCurrentProblemIndex(0)
        setSqlQuery('')
        setResult(null)
        setExamEndTimeMs(null)
        setAnswers({})
        setSolvedProblems({})
        loadMyExams() // Refresh list
    }

    const handleRunTest = async () => {
        if (!selectedExam) return

        setIsRunning(true)
        setResult(null)

        try {
            const currentProblem = examProblems[currentProblemIndex]
            const problemID = Number(
                currentProblem?.problemID ??
                    currentProblem?.problem_id ??
                    currentProblem?.problem?.id ??
                    0
            )
            if (problemID > 0) {
                const { problemsService } = await import('@/services/problems.service')
                const response = await problemsService.run(problemID, {
                    code: sqlQuery,
                    databaseType: selectedExam.exam.allowedDatabases?.[0] || 'postgresql',
                })
                setResult(response)
            }
        } catch (error: unknown) {
            setResult({ success: false, error: error?.message || 'Lỗi khi chạy truy vấn' })
        } finally {
            setIsRunning(false)
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
            const examProblemID = getExamProblemKey(currentProblem)
            const selectedDatabase = selectedExam.exam.allowedDatabases?.[0] || 'postgresql'

            if (!examProblemID) {
                toast.error('Không tìm thấy mã câu hỏi trong kỳ thi')
                return
            }

            const submitResult = await examSubmissionService.submitProblemCode(
                selectedExam.exam.id,
                examProblemID,
                {
                    code: sqlQuery,
                    databaseType: selectedDatabase,
                }
            )

            // Save submission result
            setSolvedProblems((prev) => ({
                ...prev,
                [examProblemID]: submitResult,
            }))

            // Save answer locally
            setAnswers((prev) => ({
                ...prev,
                [examProblemID]: sqlQuery,
            }))

            // Show result feedback
            if (submitResult.status === 'accepted') {
                toast.success(`✓ Câu trả lời chính xác! (+${submitResult.score} điểm)`)
            } else {
                toast.error('✗ Câu trả lời chưa chính xác. Hãy thử lại!')
            }

            // Move to next problem if available
            if (currentProblemIndex < examProblems.length - 1) {
                setCurrentProblemIndex(currentProblemIndex + 1)
                const nextProblemID = getExamProblemKey(examProblems[currentProblemIndex + 1])
                setSqlQuery(answers[nextProblemID] || '')
                setResult(null)
            }
        } catch (error: unknown) {
            console.error('Error submitting answer:', error)
            toast.error(error?.message || 'Không thể lưu câu trả lời')
        } finally {
            setIsSubmitting(false)
        }
    }

    const handleTimeExpired = useCallback(() => {
        setIsAutoSubmitMode(true)
        setShowAutoSubmitModal(true)
        toast.error('Hết giờ làm bài! Bài làm sẽ được tự động nộp.')
    }, [])

    const handleFinishExamClick = () => {
        setIsAutoSubmitMode(false)
        setShowAutoSubmitModal(true)
    }

    const handleConfirmSubmit = async () => {
        if (!selectedExam) return

        setIsFinalizingSubmit(true)

        try {
            await examSubmissionService.finishExam(selectedExam.exam.id)
            toast.success('Đã nộp bài thành công!')
            setShowAutoSubmitModal(false)
            handleBackToList()
            await loadSubmissionHistory() // Refresh history after finish
        } catch (error: unknown) {
            console.error('Error finishing exam:', error)
            toast.error(error?.message || 'Không thể nộp bài')
        } finally {
            setIsFinalizingSubmit(false)
        }
    }

    const handleRetakeExam = () => {
        setShowRetakeModal(true)
    }

    const handleConfirmRetake = async () => {
        if (!selectedExam) return

        setIsRetakingExam(true)

        try {
            // Reset exam state
            setSelectedExam(null)
            setExamProblems([])
            setCurrentProblemIndex(0)
            setSqlQuery('')
            setResult(null)
            setExamEndTimeMs(null)
            setAnswers({})
            setSolvedProblems({})
            setShowRetakeModal(false)
            
            // Start new attempt
            await handleStartExam(selectedExam)
            toast.success('Bắt đầu lần thi mới!')
        } catch (error: unknown) {
            console.error('Error retaking exam:', error)
            toast.error(error?.message || 'Không thể bắt đầu lại bài thi')
        } finally {
            setIsRetakingExam(false)
        }
    }

    const handleProblemSelect = (index: number) => {
        // Save current answer
        const currentProblem = examProblems[currentProblemIndex]
        if (currentProblem) {
            const currentProblemID = getExamProblemKey(currentProblem)
            setAnswers((prev) => ({
                ...prev,
                [currentProblemID]: sqlQuery,
            }))
        }

        // Load new problem
        setCurrentProblemIndex(index)
        const selectedProblemID = getExamProblemKey(examProblems[index])
        setSqlQuery(answers[selectedProblemID] || '')
        setResult(null)
    }

    // Exam Taking View
    if (selectedExam && selectedExam.status !== 'finished') {
        const currentProblem = examProblems[currentProblemIndex]

        return (
            <MainLayout>
                <div className="h-[calc(100vh-120px)] flex flex-col">
                    {/* Timer Header */}
                    {examEndTimeMs && (
                        <ExamTimer 
                            endTimeMs={examEndTimeMs}
                            onTimeExpired={handleTimeExpired}
                            examTitle={selectedExam.exam.title}
                        />
                    )}

                    {/* Header */}
                    <div className="flex items-center gap-4 mb-4 px-4 py-3">
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
                        <Button 
                            onClick={handleFinishExamClick}
                            variant="default"
                        >
                            <CheckCircle2 className="h-4 w-4 mr-2" />
                            Nộp bài
                        </Button>
                    </div>

                    {/* Desktop Layout */}
                    <div className="hidden lg:flex flex-1 gap-4 px-4 overflow-hidden">
                        {/* Problem Navigation Sidebar */}
                        <ExamNavigation 
                            problems={examProblems}
                            currentProblemIndex={currentProblemIndex}
                            solvedProblems={solvedProblems}
                            onSelectProblem={handleProblemSelect}
                        />

                        {/* Main Content */}
                        <div className="flex-1 flex flex-col gap-2 overflow-hidden">
                            <ResizablePanelGroup direction="horizontal" className="h-full rounded-lg border">
                                {/* Problem Description */}
                                <ResizablePanel defaultSize={50} minSize={30}>
                                    <Card className="h-full flex flex-col overflow-hidden rounded-none border-0">
                                        <ProblemHeader 
                                            problem={currentProblem?.problem}
                                            problemIndex={currentProblemIndex + 1}
                                            totalProblems={examProblems.length}
                                            points={currentProblem?.points || 0}
                                            submissionResult={solvedProblems[currentProblem?.id]}
                                            onPrevious={() => handleProblemSelect(Math.max(0, currentProblemIndex - 1))}
                                            onNext={() => handleProblemSelect(Math.min(examProblems.length - 1, currentProblemIndex + 1))}
                                            canGoBack={currentProblemIndex > 0}
                                            canGoNext={currentProblemIndex < examProblems.length - 1}
                                        />
                                        <CardContent className="flex-1 overflow-y-auto py-4">
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
                    </div>

                    {/* Mobile Layout */}
                    <div className="lg:hidden flex-1 overflow-hidden px-4">
                        <div className="flex gap-2 mb-4">
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

                        {mobileView === 'problem' ? (
                            <Card className="h-full flex flex-col overflow-hidden">
                                <ProblemHeader 
                                    problem={currentProblem?.problem}
                                    problemIndex={currentProblemIndex + 1}
                                    totalProblems={examProblems.length}
                                    points={currentProblem?.points || 0}
                                    submissionResult={solvedProblems[currentProblem?.id]}
                                    onPrevious={() => handleProblemSelect(Math.max(0, currentProblemIndex - 1))}
                                    onNext={() => handleProblemSelect(Math.min(examProblems.length - 1, currentProblemIndex + 1))}
                                    canGoBack={currentProblemIndex > 0}
                                    canGoNext={currentProblemIndex < examProblems.length - 1}
                                />
                                <CardContent className="flex-1 overflow-y-auto py-4">
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

                {/* Auto-Submit Modal */}
                <ExamAutoSubmitModal 
                    open={showAutoSubmitModal}
                    onConfirm={handleConfirmSubmit}
                    onCancel={() => setShowAutoSubmitModal(false)}
                    examTitle={selectedExam.exam.title}
                    type={isAutoSubmitMode ? 'auto-submit' : 'manual-submit'}
                    isSubmitting={isFinalizingSubmit}
                />
            </MainLayout>
        )
    }

    // View Result View (for finished exams)
    if (selectedExam && selectedExam.status === 'finished') {
        const remainingAttempts = (selectedExam.exam.maxAttempts || 1) - (selectedExam.attemptNumber || 1)
        const canRetake = remainingAttempts > 0

        return (
            <MainLayout>
                <div className="space-y-6 px-4 py-4">
                    {/* Header */}
                    <div className="flex items-center justify-between gap-4">
                        <div className="flex-1">
                            <div className="flex items-center gap-2 mb-2">
                                <h1 className="text-3xl font-bold">{selectedExam.exam.title}</h1>
                                <Badge variant="secondary">Kết quả bài thi</Badge>
                            </div>
                            <p className="text-muted-foreground">
                                Lần thi thứ {selectedExam.attemptNumber || 1} / {selectedExam.exam.maxAttempts || 1}
                            </p>
                        </div>
                        <Button 
                            variant="ghost" 
                            size="sm" 
                            onClick={handleBackToList}
                        >
                            <ArrowLeft className="h-4 w-4 mr-1" />
                            Quay lại
                        </Button>
                    </div>

                    {/* Results Overview */}
                    <ExamResultsOverview 
                        totalScore={selectedExam.score || 0}
                        maxScore={selectedExam.totalPoints || 0}
                        submittedAt={selectedExam.finishedAt}
                        attemptNumber={selectedExam.attemptNumber || 1}
                    />

                    {/* Retake Button */}
                    {canRetake && (
                        <div className="flex gap-2">
                            <Button 
                                onClick={handleRetakeExam}
                                variant="default"
                                className="gap-2"
                            >
                                <RotateCw className="h-4 w-4" />
                                Làm lại ({remainingAttempts} lần còn lại)
                            </Button>
                        </div>
                    )}

                    {/* Problem Details */}
                    <div className="space-y-3">
                        <h2 className="text-xl font-bold">Chi tiết bài làm</h2>
                        {examProblems.length === 0 ? (
                            <Card>
                                <CardContent className="py-8 text-center">
                                    <p className="text-muted-foreground">Không có dữ liệu chi tiết</p>
                                </CardContent>
                            </Card>
                        ) : (
                            <div className="space-y-3">
                                {examProblems.map((problem, index) => {
                                    const submission = solvedProblems[problem.id]
                                    return submission ? (
                                        <ProblemResultDetail
                                            key={problem.id}
                                            problemIndex={index + 1}
                                            problemTitle={problem.problem?.title || problem.title || `Câu ${index + 1}`}
                                            submission={submission}
                                        />
                                    ) : (
                                        <Card key={problem.id} className="p-4 bg-gray-50 dark:bg-gray-900">
                                            <div className="flex items-center gap-3">
                                                <AlertCircle className="h-5 w-5 text-orange-600" />
                                                <div>
                                                    <p className="font-medium">Câu {index + 1}: {problem.problem?.title || `Câu ${index + 1}`}</p>
                                                    <p className="text-sm text-muted-foreground">Không có kết quả</p>
                                                </div>
                                            </div>
                                        </Card>
                                    )
                                })}
                            </div>
                        )}
                    </div>
                </div>

                {/* Retake Modal */}
                <RetakeModal 
                    open={showRetakeModal}
                    onConfirm={handleConfirmRetake}
                    onCancel={() => setShowRetakeModal(false)}
                    examTitle={selectedExam.exam.title}
                    currentScore={selectedExam.score || 0}
                    maxScore={selectedExam.totalPoints || 0}
                    remainingAttempts={remainingAttempts}
                    maxAttempts={selectedExam.exam.maxAttempts || 1}
                    isRetaking={isRetakingExam}
                />
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
                                                    {isLecturer ? (
                                                        <Button
                                                            variant="outline"
                                                            onClick={() => navigate({ to: '/exams' })}
                                                        >
                                                            Quản lý kỳ thi
                                                        </Button>
                                                    ) : (
                                                        <Button
                                                            onClick={() => handleStartExam(myExam)}
                                                            disabled={!canStartExam(myExam.exam)}
                                                        >
                                                            {canStartExam(myExam.exam) ? 'Bắt đầu làm bài' : 'Chưa mở'}
                                                        </Button>
                                                    )}
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
                                                            {sub.executionTime !== undefined && (
                                                                <span className="flex items-center gap-1 text-muted-foreground font-medium">
                                                                    <Clock className="h-4 w-4" />
                                                                    {sub.executionTime}ms
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
        </MainLayout>
    )
}

export const Route = createFileRoute('/submissions')({
    component: SubmissionsPage,
})
