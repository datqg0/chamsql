import { createFileRoute } from '@tanstack/react-router'
import { useState, useEffect } from 'react'
import toast from 'react-hot-toast'

import { MainLayout } from '@/components/layouts/main-layout'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import { useAuthStore } from '@/stores/use-auth-store'
import {
    USE_GRADING_MOCK,
    getMockSubmissions,
    getMockGradingStats,
    mockGradeSubmission,
    mockAutoGrade,
    MOCK_EXERCISES,
    type Submission,
} from '@/mocks/grading.mock'
import {
    FileText,
    Clock,
    CheckCircle2,
    XCircle,
    AlertTriangle,
    Search,
    Filter,
    Eye,
    Wand2,
    Loader2,
    User,
    Calendar,
    Code,
    BarChart3,
} from 'lucide-react'

function GradingPage() {
    const { isOperator, userRole } = useAuthStore()
    const canAccess = isOperator() || userRole === 'lecturer'
    const [submissions, setSubmissions] = useState<Submission[]>([])
    const [stats, setStats] = useState({
        totalSubmissions: 0,
        pendingCount: 0,
        gradedCount: 0,
        errorCount: 0,
        averageScore: 0,
    })
    const [isLoading, setIsLoading] = useState(true)
    const [filterStatus, setFilterStatus] = useState<string>('all')
    const [filterExercise, setFilterExercise] = useState<string>('all')
    const [searchQuery, setSearchQuery] = useState('')

    // Dialog states
    const [selectedSubmission, setSelectedSubmission] = useState<Submission | null>(null)
    const [isDetailOpen, setIsDetailOpen] = useState(false)
    const [isGrading, setIsGrading] = useState(false)
    const [gradeScore, setGradeScore] = useState('')
    const [gradeFeedback, setGradeFeedback] = useState('')

    // Load data
    useEffect(() => {
        loadData()
    }, [filterStatus, filterExercise, searchQuery])

    const loadData = async () => {
        if (!USE_GRADING_MOCK) return

        setIsLoading(true)
        try {
            const filters: any = {}
            if (filterStatus !== 'all') {
                filters.status = filterStatus
            }
            if (filterExercise !== 'all') {
                filters.exerciseId = parseInt(filterExercise)
            }
            if (searchQuery) {
                filters.studentCode = searchQuery
            }

            const [submissionsData, statsData] = await Promise.all([
                getMockSubmissions(filters),
                getMockGradingStats(),
            ])

            setSubmissions(submissionsData)
            setStats(statsData)
        } finally {
            setIsLoading(false)
        }
    }

    const handleViewDetail = (submission: Submission) => {
        setSelectedSubmission(submission)
        setGradeScore(submission.score?.toString() || '')
        setGradeFeedback(submission.feedback || '')
        setIsDetailOpen(true)
    }

    const handleManualGrade = async () => {
        if (!selectedSubmission) return

        const score = parseFloat(gradeScore)
        if (isNaN(score) || score < 0 || score > selectedSubmission.maxScore) {
            toast.error(`Điểm phải từ 0 đến ${selectedSubmission.maxScore}`)
            return
        }

        setIsGrading(true)
        try {
            const result = await mockGradeSubmission(
                selectedSubmission.id,
                score,
                gradeFeedback
            )
            if (result.success) {
                toast.success(result.message)
                setIsDetailOpen(false)
                loadData()
            } else {
                toast.error(result.message)
            }
        } finally {
            setIsGrading(false)
        }
    }

    const handleAutoGrade = async (submission: Submission) => {
        setIsGrading(true)
        try {
            const result = await mockAutoGrade(submission.id)
            if (result.success) {
                toast.success(`Chấm tự động: ${result.score}/${submission.maxScore} điểm`)
                loadData()
            }
        } finally {
            setIsGrading(false)
        }
    }

    const getStatusBadge = (status: Submission['status']) => {
        switch (status) {
            case 'pending':
                return <Badge variant="secondary" className="bg-yellow-500/10 text-yellow-600"><Clock className="w-3 h-3 mr-1" />Chờ chấm</Badge>
            case 'graded':
                return <Badge variant="secondary" className="bg-green-500/10 text-green-600"><CheckCircle2 className="w-3 h-3 mr-1" />Đã chấm</Badge>
            case 'error':
                return <Badge variant="destructive"><XCircle className="w-3 h-3 mr-1" />Lỗi</Badge>
            default:
                return null
        }
    }

    const formatDate = (dateStr: string) => {
        return new Date(dateStr).toLocaleString('vi-VN', {
            day: '2-digit',
            month: '2-digit',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
        })
    }

    if (!canAccess) {
        return (
            <MainLayout>
                <div className="text-center py-12">
                    <p className="text-muted-foreground">Bạn không có quyền truy cập trang này</p>
                </div>
            </MainLayout>
        )
    }

    return (
        <MainLayout>
            <div className="space-y-6">
                {/* Header */}
                <div>
                    <h1 className="text-3xl font-bold">Chấm bài</h1>
                    <p className="text-muted-foreground">Xem và chấm điểm các bài nộp của sinh viên</p>
                </div>

                {/* Statistics */}
                <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
                    <Card>
                        <CardContent className="pt-4">
                            <div className="flex items-center gap-2">
                                <FileText className="h-5 w-5 text-primary" />
                                <div>
                                    <p className="text-2xl font-bold">{stats.totalSubmissions}</p>
                                    <p className="text-xs text-muted-foreground">Tổng bài nộp</p>
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                    <Card>
                        <CardContent className="pt-4">
                            <div className="flex items-center gap-2">
                                <Clock className="h-5 w-5 text-yellow-500" />
                                <div>
                                    <p className="text-2xl font-bold">{stats.pendingCount}</p>
                                    <p className="text-xs text-muted-foreground">Chờ chấm</p>
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                    <Card>
                        <CardContent className="pt-4">
                            <div className="flex items-center gap-2">
                                <CheckCircle2 className="h-5 w-5 text-green-500" />
                                <div>
                                    <p className="text-2xl font-bold">{stats.gradedCount}</p>
                                    <p className="text-xs text-muted-foreground">Đã chấm</p>
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                    <Card>
                        <CardContent className="pt-4">
                            <div className="flex items-center gap-2">
                                <AlertTriangle className="h-5 w-5 text-red-500" />
                                <div>
                                    <p className="text-2xl font-bold">{stats.errorCount}</p>
                                    <p className="text-xs text-muted-foreground">Lỗi</p>
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                    <Card>
                        <CardContent className="pt-4">
                            <div className="flex items-center gap-2">
                                <BarChart3 className="h-5 w-5 text-blue-500" />
                                <div>
                                    <p className="text-2xl font-bold">{stats.averageScore}</p>
                                    <p className="text-xs text-muted-foreground">Điểm TB</p>
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                </div>

                {/* Filters */}
                <Card>
                    <CardHeader className="pb-3">
                        <CardTitle className="text-base flex items-center gap-2">
                            <Filter className="h-4 w-4" />
                            Bộ lọc
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="flex flex-wrap gap-4">
                            <div className="flex-1 min-w-[200px]">
                                <div className="relative">
                                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                                    <Input
                                        placeholder="Tìm theo mã SV hoặc tên..."
                                        value={searchQuery}
                                        onChange={(e) => setSearchQuery(e.target.value)}
                                        className="pl-9"
                                    />
                                </div>
                            </div>
                            <Select value={filterStatus} onValueChange={setFilterStatus}>
                                <SelectTrigger className="w-[160px]">
                                    <SelectValue placeholder="Trạng thái" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">Tất cả</SelectItem>
                                    <SelectItem value="pending">Chờ chấm</SelectItem>
                                    <SelectItem value="graded">Đã chấm</SelectItem>
                                    <SelectItem value="error">Lỗi</SelectItem>
                                </SelectContent>
                            </Select>
                            <Select value={filterExercise} onValueChange={setFilterExercise}>
                                <SelectTrigger className="w-[200px]">
                                    <SelectValue placeholder="Bài tập" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">Tất cả bài tập</SelectItem>
                                    {MOCK_EXERCISES.map((ex) => (
                                        <SelectItem key={ex.id} value={ex.id.toString()}>
                                            {ex.title}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                    </CardContent>
                </Card>

                {/* Submissions List */}
                <Card>
                    <CardHeader>
                        <CardTitle>Danh sách bài nộp</CardTitle>
                        <CardDescription>
                            {submissions.length} bài nộp
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        {isLoading ? (
                            <div className="flex items-center justify-center py-12">
                                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                            </div>
                        ) : submissions.length === 0 ? (
                            <div className="text-center py-12 text-muted-foreground">
                                Không có bài nộp nào
                            </div>
                        ) : (
                            <div className="space-y-3">
                                {submissions.map((submission) => (
                                    <div
                                        key={submission.id}
                                        className="flex items-center justify-between p-4 border rounded-lg hover:bg-muted/50 transition-colors"
                                    >
                                        <div className="flex-1 space-y-1">
                                            <div className="flex items-center gap-3">
                                                <span className="font-medium">{submission.studentName}</span>
                                                <Badge variant="outline">{submission.studentCode}</Badge>
                                                {getStatusBadge(submission.status)}
                                            </div>
                                            <div className="flex items-center gap-4 text-sm text-muted-foreground">
                                                <span className="flex items-center gap-1">
                                                    <FileText className="h-3 w-3" />
                                                    {submission.exerciseName}
                                                </span>
                                                <span className="flex items-center gap-1">
                                                    <Calendar className="h-3 w-3" />
                                                    {formatDate(submission.submittedAt)}
                                                </span>
                                                {submission.score !== undefined && (
                                                    <span className="font-medium text-foreground">
                                                        {submission.score}/{submission.maxScore} điểm
                                                    </span>
                                                )}
                                            </div>
                                        </div>
                                        <div className="flex gap-2">
                                            {submission.status === 'pending' && (
                                                <Button
                                                    variant="outline"
                                                    size="sm"
                                                    onClick={() => handleAutoGrade(submission)}
                                                    disabled={isGrading}
                                                >
                                                    <Wand2 className="h-4 w-4 mr-1" />
                                                    Chấm tự động
                                                </Button>
                                            )}
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                onClick={() => handleViewDetail(submission)}
                                            >
                                                <Eye className="h-4 w-4 mr-1" />
                                                Chi tiết
                                            </Button>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </CardContent>
                </Card>
            </div>

            {/* Detail Dialog */}
            <Dialog open={isDetailOpen} onOpenChange={setIsDetailOpen}>
                <DialogContent className="max-w-3xl max-h-[90vh] overflow-y-auto">
                    <DialogHeader>
                        <DialogTitle>Chi tiết bài nộp</DialogTitle>
                        <DialogDescription>
                            {selectedSubmission?.studentName} - {selectedSubmission?.exerciseName}
                        </DialogDescription>
                    </DialogHeader>

                    {selectedSubmission && (
                        <div className="space-y-4">
                            {/* Student Info */}
                            <div className="grid grid-cols-2 gap-4 p-4 bg-muted/50 rounded-lg">
                                <div className="flex items-center gap-2">
                                    <User className="h-4 w-4 text-muted-foreground" />
                                    <span className="text-sm">
                                        <strong>Sinh viên:</strong> {selectedSubmission.studentName} ({selectedSubmission.studentCode})
                                    </span>
                                </div>
                                <div className="flex items-center gap-2">
                                    <Calendar className="h-4 w-4 text-muted-foreground" />
                                    <span className="text-sm">
                                        <strong>Nộp lúc:</strong> {formatDate(selectedSubmission.submittedAt)}
                                    </span>
                                </div>
                                <div className="flex items-center gap-2">
                                    <Clock className="h-4 w-4 text-muted-foreground" />
                                    <span className="text-sm">
                                        <strong>Thời gian:</strong> {selectedSubmission.executionTime || 'N/A'}
                                    </span>
                                </div>
                                <div className="flex items-center gap-2">
                                    {getStatusBadge(selectedSubmission.status)}
                                </div>
                            </div>

                            {/* SQL Query */}
                            <div>
                                <label className="text-sm font-medium flex items-center gap-2 mb-2">
                                    <Code className="h-4 w-4" />
                                    Câu truy vấn SQL
                                </label>
                                <pre className="p-4 bg-muted rounded-lg text-sm overflow-x-auto">
                                    {selectedSubmission.sqlQuery}
                                </pre>
                            </div>

                            {/* Grading Form */}
                            <div className="grid grid-cols-2 gap-4">
                                <div>
                                    <label className="text-sm font-medium">
                                        Điểm (tối đa {selectedSubmission.maxScore})
                                    </label>
                                    <Input
                                        type="number"
                                        min="0"
                                        max={selectedSubmission.maxScore}
                                        value={gradeScore}
                                        onChange={(e) => setGradeScore(e.target.value)}
                                        placeholder="Nhập điểm..."
                                        className="mt-1"
                                    />
                                </div>
                                <div className="flex items-end">
                                    {selectedSubmission.isCorrect !== undefined && (
                                        <Badge
                                            variant={selectedSubmission.isCorrect ? 'default' : 'destructive'}
                                            className="mb-2"
                                        >
                                            {selectedSubmission.isCorrect ? '✓ Đúng' : '✗ Sai'}
                                        </Badge>
                                    )}
                                </div>
                            </div>

                            <div>
                                <label className="text-sm font-medium">Nhận xét</label>
                                <Textarea
                                    value={gradeFeedback}
                                    onChange={(e) => setGradeFeedback(e.target.value)}
                                    placeholder="Nhập nhận xét cho sinh viên..."
                                    className="mt-1"
                                    rows={3}
                                />
                            </div>
                        </div>
                    )}

                    <DialogFooter>
                        <Button variant="outline" onClick={() => setIsDetailOpen(false)}>
                            Đóng
                        </Button>
                        {selectedSubmission?.status === 'pending' && (
                            <Button
                                onClick={() => handleAutoGrade(selectedSubmission)}
                                disabled={isGrading}
                                variant="secondary"
                            >
                                {isGrading ? (
                                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                                ) : (
                                    <Wand2 className="h-4 w-4 mr-2" />
                                )}
                                Chấm tự động
                            </Button>
                        )}
                        <Button onClick={handleManualGrade} disabled={isGrading}>
                            {isGrading ? (
                                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                            ) : (
                                <CheckCircle2 className="h-4 w-4 mr-2" />
                            )}
                            Lưu điểm
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </MainLayout>
    )
}

export const Route = createFileRoute('/grading')({
    component: GradingPage,
})
