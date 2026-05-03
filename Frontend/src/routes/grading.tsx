import { createFileRoute } from '@tanstack/react-router'
import {
    FileText,
    Clock,
    CheckCircle2,
    TrendingUp,
    Filter,
    Search,
    Calendar,
    XCircle,
    Wand2,
    Eye,
    Code,
    Loader2,
    User,
} from 'lucide-react'
import { useState, useEffect, useCallback } from 'react'
import toast from 'react-hot-toast'
import {
    BarChart,
    Bar,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    ResponsiveContainer,
    Cell,
} from 'recharts'

import { MainLayout } from '@/components/layouts/main-layout'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { examsService } from '@/services/exams.service'
import { gradingService, type Submission } from '@/services/grading.service'
import { useAuthStore } from '@/stores/use-auth-store'
import type { Exam } from '@/types/exam.types'

function GradingPage() {
    const { isOperator, userRole } = useAuthStore()
    const canAccess = isOperator() || userRole === 'lecturer'
    const [exams, setExams] = useState<Exam[]>([])
    const [selectedExamId, setSelectedExamId] = useState<number | null>(null)
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
    const [searchQuery, setSearchQuery] = useState('')

    // Dialog states
    const [selectedSubmission, setSelectedSubmission] = useState<Submission | null>(null)
    const [isDetailOpen, setIsDetailOpen] = useState(false)
    const [isGrading, setIsGrading] = useState(false)
    const [gradeScore, setGradeScore] = useState('')
    const [gradeFeedback, setGradeFeedback] = useState('')

    const loadExams = useCallback(async () => {
        try {
            const examsData = await examsService.list()
            setExams(examsData)
            if (examsData.length > 0) {
                setSelectedExamId(examsData[0].id)
            }
        } catch {
            toast.error('Không thể tải danh sách kỳ thi')
        }
    }, [])

    const loadData = useCallback(async () => {
        if (!selectedExamId) return

        setIsLoading(true)
        try {
            const [submissionsData, statsData] = await Promise.all([
                gradingService.listUngradedSubmissions(selectedExamId),
                gradingService.getGradingStats(selectedExamId),
            ])

            // Filter by status if needed
            let filtered = submissionsData
            if (filterStatus !== 'all') {
                filtered = filtered.filter(s => s.status === filterStatus)
            }
            if (searchQuery) {
                filtered = filtered.filter(s => 
                    s.studentCode?.toLowerCase().includes(searchQuery.toLowerCase()) ||
                    s.studentName?.toLowerCase().includes(searchQuery.toLowerCase())
                )
            }

            setSubmissions(filtered)
            setStats(statsData)
        } catch {
            toast.error('Không thể tải dữ liệu bài nộp')
        } finally {
            setIsLoading(false)
        }
    }, [selectedExamId, filterStatus, searchQuery])

    const executionTimeDistribution = (() => {
        const buckets = [
            { range: '0-50ms', min: 0, max: 50, count: 0, color: '#22c55e', label: 'Rất nhanh' },
            { range: '51-200ms', min: 51, max: 200, count: 0, color: '#3b82f6', label: 'Nhanh' },
            { range: '201-500ms', min: 201, max: 500, count: 0, color: '#eab308', label: 'Trung bình' },
            { range: '>500ms', min: 501, max: Infinity, count: 0, color: '#ef4444', label: 'Chậm' },
        ]

        submissions.forEach((sub) => {
            if (sub.executionTimeMs !== undefined) {
                const time = sub.executionTimeMs
                const bucket = buckets.find((b) => time >= b.min && time <= b.max)
                if (bucket) bucket.count++
            }
        })

        return buckets
    })()

    // Load exams on mount
    useEffect(() => {
        loadExams()
    }, [loadExams])

    // Load submissions when exam selected
    useEffect(() => {
        if (selectedExamId) {
            loadData()
        }
    }, [selectedExamId, filterStatus, searchQuery, loadData])

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
            await gradingService.gradeSubmission(
                selectedSubmission.id,
                score,
                gradeFeedback
            )
            toast.success('Đã lưu điểm thành công')
            setIsDetailOpen(false)
            loadData()
        } catch (error: unknown) {
            toast.error((error as Error)?.message || 'Chấm điểm thất bại')
        } finally {
            setIsGrading(false)
        }
    }

    const handleAutoGrade = async (submission: Submission) => {
        setIsGrading(true)
        try {
            const result = await gradingService.autoGrade(submission.id)
            toast.success(`Chấm tự động: ${result.score}/${submission.maxScore} điểm`)
            loadData()
        } catch (error: unknown) {
            toast.error((error as Error)?.message || 'Chấm tự động thất bại')
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

                {/* Statistics & Analytics */}
                <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
                    {/* Performance Distribution Chart */}
                    <Card className="lg:col-span-1 border-primary/10 shadow-sm">
                        <CardHeader className="pb-2 pt-4">
                            <CardTitle className="text-sm font-semibold flex items-center gap-2">
                                <TrendingUp className="h-4 w-4 text-primary" />
                                Phân phối Hiệu năng
                            </CardTitle>
                        </CardHeader>
                        <CardContent className="pt-0">
                            <div className="h-[180px] w-full mt-2">
                                <ResponsiveContainer width="100%" height="100%">
                                    <BarChart data={executionTimeDistribution}>
                                        <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#e5e7eb" />
                                        <XAxis 
                                            dataKey="range" 
                                            fontSize={10} 
                                            tickLine={false} 
                                            axisLine={false} 
                                            tick={{ fill: '#6b7280' }}
                                        />
                                        <YAxis 
                                            fontSize={10} 
                                            tickLine={false} 
                                            axisLine={false} 
                                            allowDecimals={false}
                                            tick={{ fill: '#6b7280' }}
                                        />
                                        <Tooltip 
                                            contentStyle={{ fontSize: '12px', borderRadius: '8px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
                                            cursor={{ fill: 'rgba(0,0,0,0.05)' }}
                                        />
                                        <Bar dataKey="count" radius={[4, 4, 0, 0]}>
                                            {executionTimeDistribution.map((entry, index) => (
                                                <Cell key={`cell-${index}`} fill={entry.color} />
                                            ))}
                                        </Bar>
                                    </BarChart>
                                </ResponsiveContainer>
                            </div>
                        </CardContent>
                    </Card>

                    {/* Statistics Cards */}
                    <div className="lg:col-span-3 grid grid-cols-2 md:grid-cols-4 gap-4">
                        <Card className="bg-primary/5 border-primary/10">
                            <CardContent className="pt-6">
                                <div className="flex flex-col items-center text-center gap-2">
                                    <div className="p-2 bg-primary/10 rounded-full">
                                        <FileText className="h-5 w-5 text-primary" />
                                    </div>
                                    <div>
                                        <p className="text-2xl font-bold">{stats.totalSubmissions}</p>
                                        <p className="text-[10px] uppercase tracking-wider font-semibold text-muted-foreground">Tổng bài nộp</p>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                        <Card className="bg-yellow-500/5 border-yellow-500/10">
                            <CardContent className="pt-6">
                                <div className="flex flex-col items-center text-center gap-2">
                                    <div className="p-2 bg-yellow-500/10 rounded-full">
                                        <Clock className="h-5 w-5 text-yellow-600" />
                                    </div>
                                    <div>
                                        <p className="text-2xl font-bold text-yellow-600">{stats.pendingCount}</p>
                                        <p className="text-[10px] uppercase tracking-wider font-semibold text-muted-foreground">Chờ chấm</p>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                        <Card className="bg-green-500/5 border-green-500/10">
                            <CardContent className="pt-6">
                                <div className="flex flex-col items-center text-center gap-2">
                                    <div className="p-2 bg-green-500/10 rounded-full">
                                        <CheckCircle2 className="h-5 w-5 text-green-600" />
                                    </div>
                                    <div>
                                        <p className="text-2xl font-bold text-green-600">{stats.gradedCount}</p>
                                        <p className="text-[10px] uppercase tracking-wider font-semibold text-muted-foreground">Đã chấm</p>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                        <Card className="bg-blue-500/5 border-blue-500/10">
                            <CardContent className="pt-6">
                                <div className="flex flex-col items-center text-center gap-2">
                                    <div className="p-2 bg-blue-500/10 rounded-full">
                                        <TrendingUp className="h-5 w-5 text-blue-600" />
                                    </div>
                                    <div>
                                        <p className="text-2xl font-bold text-blue-600">{stats.averageScore.toFixed(1)}</p>
                                        <p className="text-[10px] uppercase tracking-wider font-semibold text-muted-foreground">Điểm trung bình</p>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                    </div>
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
                            <Select value={selectedExamId?.toString() || ''} onValueChange={(v) => setSelectedExamId(parseInt(v))}>
                                <SelectTrigger className="w-[250px]">
                                    <SelectValue placeholder="Chọn kỳ thi" />
                                </SelectTrigger>
                                <SelectContent>
                                    {exams.map((exam) => (
                                        <SelectItem key={exam.id} value={exam.id.toString()}>
                                            {exam.title}
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
                                                    {submission.problemTitle}
                                                </span>
                                                <span className="flex items-center gap-1">
                                                    <Calendar className="h-3 w-3" />
                                                    {formatDate(submission.submittedAt)}
                                                </span>
                                                {submission.executionTimeMs !== undefined && (
                                                    <span className="flex items-center gap-1">
                                                        <Clock className="h-3 w-3" />
                                                        {submission.executionTimeMs}ms
                                                    </span>
                                                )}
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
                            {selectedSubmission?.studentName} - {selectedSubmission?.problemTitle}
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
                                        <strong>Thời gian:</strong> {selectedSubmission.executionTimeMs ? `${selectedSubmission.executionTimeMs}ms` : 'N/A'}
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
                                    {selectedSubmission.submittedCode}
                                </pre>
                            </div>

                            {/* Error Message */}
                            {selectedSubmission.errorMessage && (
                                <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-lg">
                                    <label className="text-sm font-medium text-red-600">Lỗi:</label>
                                    <p className="text-sm text-red-600 mt-1">{selectedSubmission.errorMessage}</p>
                                </div>
                            )}

                            {/* Grading Form */}
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
