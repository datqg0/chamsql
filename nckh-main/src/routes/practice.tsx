import { createFileRoute } from '@tanstack/react-router'
import { useState, useEffect } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import rehypeRaw from 'rehype-raw'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { MainLayout } from '@/components/layouts/main-layout'
import { SQLEditor } from '@/components/editor/sql-editor'
import { AIChatPanel } from '@/components/ai/ai-chat-panel'
import { CreateTopicDialog } from '@/components/practice/create-topic-dialog'
import { CreateProblemDialog } from '@/components/practice/create-problem-dialog'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import {
    ResizablePanelGroup,
    ResizablePanel,
    ResizableHandle,
} from '@/components/ui/resizable'
import { useSQLChecker } from '@/hooks/use-sql-checker'
import { useAuthStore } from '@/stores/use-auth-store'
import { topicsService } from '@/services/topics.service'
import { problemsService } from '@/services/problems.service'
import type { Topic, Problem } from '@/types/exam.types'
import {
    ArrowLeft,
    Play,
    Send,
    Loader2,
    Code,
    FileText,
    BookOpen,
    Table,
} from 'lucide-react'
import { cn } from '@/lib/utils'
function PracticePage() {
    const { userRole, isOperator } = useAuthStore()
    const queryClient = useQueryClient()
    const isLecturer = isOperator() || userRole === 'lecturer'
    // Navigation state
    const [selectedTopic, setSelectedTopic] = useState<Topic | null>(null)
    const [selectedProblem, setSelectedProblem] = useState<Problem | null>(null)
    const [difficultyFilter, setDifficultyFilter] = useState<string>('all')
    // Editor state
    const [sqlQuery, setSqlQuery] = useState('')
    const [isRunning, setIsRunning] = useState(false)
    const [isSubmitting, setIsSubmitting] = useState(false)
    const [result, setResult] = useState<any>(null)
    const [activeTab, setActiveTab] = useState<'problem' | 'result'>('problem')
    const [mobileView, setMobileView] = useState<'problem' | 'editor'>('problem')
    const { isValid, syntaxError } = useSQLChecker(sqlQuery, {
        debounceMs: 300,
        database: 'MySQL',
    })
    // Fetch topics
    const { data: topics = [], isLoading: loadingTopics } = useQuery({
        queryKey: ['topics'],
        queryFn: () => topicsService.list(),
    })
    // Fetch all problems
    const { data: allProblemsData, isLoading: loadingAllProblems } = useQuery({
        queryKey: ['all-problems', difficultyFilter],
        queryFn: async () => {
            const filters: any = {}
            if (difficultyFilter !== 'all') {
                filters.difficulty = difficultyFilter
            }
            return await problemsService.list(filters)
        },
    })
    const allProblems = Array.isArray(allProblemsData) ? allProblemsData : []
    // Fetch problems for selected topic
    const { data: topicProblemsData, isLoading: loadingTopicProblems } = useQuery({
        queryKey: ['topic-problems', selectedTopic?.id],
        queryFn: async () => {
            if (!selectedTopic?.id) return []
            const filters = { topicId: selectedTopic.id }
            return await problemsService.list(filters)
        },
        enabled: !!selectedTopic,
    })
    const topicProblems = Array.isArray(topicProblemsData) ? topicProblemsData : []
    const handleBackToTopics = () => {
        setSelectedTopic(null)
    }
    const handleBackToProblems = () => {
        setSelectedProblem(null)
        setSqlQuery('')
        setResult(null)
    }
    const handleSelectTopic = (topic: Topic) => {
        setSelectedTopic(topic)
    }
    const handleSelectProblem = (problem: Problem) => {
        setSelectedProblem(problem)
        setSqlQuery('')
        setResult(null)
        setActiveTab('problem')
    }
    const handleRunTest = async () => {
        if (!selectedProblem) return
        setIsRunning(true)
        setResult(null)
        setActiveTab('result') // Switch to result tab
        try {
            const response = await problemsService.run(selectedProblem.id, {
                code: sqlQuery,
                databaseType: selectedProblem.supportedDatabases[0] || 'postgresql',
            })
            // Parse nested response: response.data contains the actual result
            const actualData = response.data || response // Adjust based on your API response wrapper
            setResult({ type: 'run', data: actualData })
        } catch (error: any) {
            setResult({ type: 'error', error: error?.message || 'Lỗi khi chạy truy vấn' })
        } finally {
            setIsRunning(false)
        }
    }
    const handleSubmit = async () => {
        if (!isValid || !selectedProblem) {
            toast.error('Vui lòng sửa lỗi cú pháp!')
            return
        }
        setIsSubmitting(true)
        setResult(null)
        setActiveTab('result')
        try {
            const response = await problemsService.submit(selectedProblem.id, {
                code: sqlQuery,
                databaseType: selectedProblem.supportedDatabases[0] || 'postgresql',
            })

            // Extract data
            const data = response.data || response
            if (data.status === 'accepted') {
                toast.success('Chúc mừng! Bài làm của bạn đúng!')
                setResult({ type: 'submit', success: true, data: data })
            } else {
                // toast.error(data.message || 'Bài làm chưa đúng, hãy thử lại!')
                setResult({ type: 'submit', success: false, data: data })
            }
        } catch (error: any) {
            toast.error(error?.message || 'Không thể nộp bài')
            setResult({ type: 'error', error: error?.message })
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
            <Badge variant="secondary" className={config[difficulty]?.className}>
                {config[difficulty]?.label}
            </Badge>
        )
    }
    const renderTable = (columns: string[], rows: any[][]) => {
        if (!columns || !rows) return null
        return (
            <div className="overflow-x-auto rounded-md border">
                <table className="w-full text-sm border-collapse">
                    <thead className="bg-muted/50">
                        <tr>
                            {columns.map((col, idx) => (
                                <th key={idx} className="text-left p-2 font-semibold border-b border-r last:border-r-0 whitespace-nowrap">
                                    {col}
                                </th>
                            ))}
                        </tr>
                    </thead>
                    <tbody>
                        {rows.map((row, rowIdx) => (
                            <tr key={rowIdx} className="border-b last:border-b-0 hover:bg-muted/20">
                                {row.map((cell, cellIdx) => (
                                    <td key={cellIdx} className="p-2 border-r last:border-r-0 font-mono text-xs">
                                        {cell === null ? (
                                            <span className="text-muted-foreground italic">NULL</span>
                                        ) : (
                                            String(cell)
                                        )}
                                    </td>
                                ))}
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        )
    }
    const renderResultContent = () => {
        if (!result) {
            return (
                <div className="flex flex-col items-center justify-center h-full text-muted-foreground p-8">
                    <Play className="h-8 w-8 mb-2 opacity-50" />
                    <p>Chạy thử code để xem kết quả ở đây</p>
                </div>
            )
        }
        if (result.type === 'error') {
            return (
                <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-md text-red-600">
                    <h3 className="font-semibold mb-1">Lỗi thực thi:</h3>
                    <pre className="text-sm whitespace-pre-wrap">{result.error}</pre>
                </div>
            )
        }
        if (result.type === 'run') {
            const { success, columns, rows, error, executionMs, rowCount } = result.data

            if (!success) {
                return (
                    <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-md text-red-600">
                        <h3 className="font-semibold mb-1">Lỗi SQL:</h3>
                        <pre className="text-sm whitespace-pre-wrap">{error}</pre>
                    </div>
                )
            }
            return (
                <div className="space-y-4">
                    <div className="flex items-center justify-between">
                        <Badge variant="outline" className="bg-green-500/10 text-green-600 border-green-200">
                            Thành công
                        </Badge>
                        <span className="text-xs text-muted-foreground">
                            {rowCount} dòng • {executionMs}ms
                        </span>
                    </div>
                    {rows && rows.length > 0 ? (
                        renderTable(columns, rows)
                    ) : (
                        <div className="text-center p-4 text-muted-foreground italic border rounded bg-muted/20">
                            Truy vấn không trả về dữ liệu nào.
                        </div>
                    )}
                </div>
            )
        }
        if (result.type === 'submit') {
            const { isCorrect: dataIsCorrect, status, message, error, expectedOutput, actualOutput, executionMs } = result.data
            const isCorrect = result.success ?? dataIsCorrect

            // Parse outputs if they are JSON strings
            let expectedRows = []
            let actualRows = []
            try {
                if (typeof expectedOutput === 'string') expectedRows = JSON.parse(expectedOutput)
                else expectedRows = expectedOutput

                if (typeof actualOutput === 'string') actualRows = JSON.parse(actualOutput)
                else actualRows = actualOutput
            } catch (e) {
                console.error("Error parsing output JSON", e)
            }
            return (
                <div className="space-y-4">
                    <div className={`p-4 rounded-md border ${isCorrect ? 'bg-green-500/10 border-green-500/20' : 'bg-red-500/10 border-red-500/20'}`}>
                        <div className="flex items-center justify-between mb-2">
                            <h3 className={`font-bold ${isCorrect ? 'text-green-700' : 'text-red-700'}`}>
                                {isCorrect ? 'Chính xác! ' : 'Sai kết quả '}
                            </h3>
                            <Badge variant={isCorrect ? "default" : "destructive"}>{status || (isCorrect ? 'Accepted' : 'Failed')}</Badge>
                        </div>
                        <p className="text-sm font-medium">{message || error || 'Có lỗi xảy ra'}</p>
                        {executionMs && <p className="text-xs text-muted-foreground mt-2">Thời gian chạy: {executionMs}ms</p>}
                    </div>
                    {!isCorrect && (
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <h4 className="font-medium text-sm flex items-center">
                                    <span className="w-2 h-2 rounded-full bg-green-500 mr-2"></span>
                                    Kết quả mong đợi (Expected)
                                </h4>
                                <div className="max-h-[300px] overflow-auto border rounded bg-muted/10 p-2 text-xs">
                                    <pre>{JSON.stringify(expectedRows, null, 2)}</pre>
                                </div>
                            </div>
                            <div className="space-y-2">
                                <h4 className="font-medium text-sm flex items-center">
                                    <span className="w-2 h-2 rounded-full bg-red-500 mr-2"></span>
                                    Kết quả của bạn (Actual)
                                </h4>
                                <div className="max-h-[300px] overflow-auto border rounded bg-muted/10 p-2 text-xs">
                                    <pre>{JSON.stringify(actualRows, null, 2)}</pre>
                                </div>
                            </div>
                        </div>
                    )}
                </div>
            )
        }
        return null
    }
    // Problem Detail View
    if (selectedProblem) {
        return (
            <MainLayout>
                <div className="h-[calc(100vh-120px)] flex flex-col">
                    {/* Header */}
                    <div className="flex items-center gap-4 mb-4">
                        <Button variant="ghost" size="sm" onClick={handleBackToProblems}>
                            <ArrowLeft className="h-4 w-4 mr-1" />
                            Quay lại
                        </Button>
                        <div className="flex-1">
                            <div className="flex items-center gap-2">
                                <h1 className="text-xl font-bold">{selectedProblem.title}</h1>
                                {getDifficultyBadge(selectedProblem.difficulty)}
                            </div>
                        </div>
                    </div>
                    {/* Desktop Layout */}
                    <div className="hidden lg:block flex-1">
                        <ResizablePanelGroup direction="horizontal" className="h-full rounded-lg border">
                            {/* Problem Description */}
                            <ResizablePanel defaultSize={40} minSize={30}>
                                <div className="h-full flex flex-col bg-background">
                                    <div className="border-b px-4 py-3 flex items-center justify-between bg-muted/40">
                                        <div className="flex items-center gap-2 font-medium">
                                            <BookOpen className="h-4 w-4" />
                                            Đề bài
                                        </div>
                                    </div>
                                    <div className="flex-1 overflow-y-auto p-4">
                                        <div className="prose prose-sm dark:prose-invert max-w-none">
                                            <ReactMarkdown
                                                remarkPlugins={[remarkGfm]}
                                                rehypePlugins={[rehypeRaw]}
                                            >
                                                {selectedProblem.description}
                                            </ReactMarkdown>
                                        </div>
                                    </div>
                                </div>
                            </ResizablePanel>
                            <ResizableHandle withHandle />
                            {/* SQL Editor & Results */}
                            <ResizablePanel defaultSize={60} minSize={30}>
                                <ResizablePanelGroup direction="vertical">
                                    <ResizablePanel defaultSize={60} minSize={30}>
                                        <div className="h-full flex flex-col">
                                            <div className="border-b px-4 py-2 flex items-center justify-between bg-muted/40 shrink-0">
                                                <div className="flex items-center gap-2 font-medium">
                                                    <Code className="h-4 w-4" />
                                                    Viết câu truy vấn
                                                </div>
                                                <div className="flex items-center gap-2">
                                                    <Button
                                                        variant="secondary"
                                                        size="sm"
                                                        onClick={handleRunTest}
                                                        type="button"
                                                        disabled={isRunning || isSubmitting}
                                                    >
                                                        {isRunning ? <Loader2 className="h-3 w-3 animate-spin mr-1" /> : <Play className="h-3 w-3 mr-1" />}
                                                        Run
                                                    </Button>
                                                    <Button
                                                        size="sm"
                                                        onClick={handleSubmit}
                                                        type="button"
                                                        disabled={isRunning || isSubmitting || !isValid}
                                                    >
                                                        {isSubmitting ? <Loader2 className="h-3 w-3 animate-spin mr-1" /> : <Send className="h-3 w-3 mr-1" />}
                                                        Submit
                                                    </Button>
                                                </div>
                                            </div>

                                            <div className="flex-1 min-h-0 bg-background">
                                                <SQLEditor
                                                    value={sqlQuery}
                                                    onChange={(v) => setSqlQuery(v || '')}
                                                    height="100%"
                                                    syntaxError={syntaxError}
                                                />
                                            </div>
                                        </div>
                                    </ResizablePanel>

                                    <ResizableHandle withHandle />

                                    <ResizablePanel defaultSize={40} minSize={20}>
                                        <div className="h-full flex flex-col bg-muted/10">
                                            <div className="flex items-center border-b bg-muted/40 px-2 shrink-0">
                                                <button
                                                    className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${activeTab === 'result' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'}`}
                                                    onClick={() => setActiveTab('result')}
                                                >
                                                    Kết quả
                                                </button>
                                            </div>
                                            <div className="flex-1 overflow-auto p-4 bg-background">
                                                {renderResultContent()}
                                            </div>
                                        </div>
                                    </ResizablePanel>
                                </ResizablePanelGroup>
                            </ResizablePanel>
                        </ResizablePanelGroup>
                    </div>
                    {/* Mobile Layout */}
                    <div className="lg:hidden flex-1 overflow-hidden">
                        {/* Simplified mobile view... (giữ nguyên hoặc cập nhật tương tự) */}
                        <div className="p-4 text-center text-muted-foreground">
                            Vui lòng sử dụng máy tính để có trải nghiệm tốt nhất.
                        </div>
                    </div>
                </div>
                {/* AI Chat Panel */}
                <AIChatPanel
                    exerciseContext={{
                        title: selectedProblem.title,
                        description: selectedProblem.description,
                    }}
                />
            </MainLayout>
        )
    }
    // ... (Phần render Topics/Problems list giữ nguyên như cũ)
    return (
        <MainLayout>
            <div className="space-y-6">
                {/* Header */}
                <div>
                    <h1 className="text-3xl font-bold">Luyện tập SQL</h1>
                    <p className="text-muted-foreground">Chọn chủ đề hoặc bài tập để bắt đầu</p>
                </div>
                <Tabs defaultValue="topics" className="w-full">
                    <TabsList className="grid w-full grid-cols-2 max-w-md">
                        <TabsTrigger value="topics">Chủ đề</TabsTrigger>
                        <TabsTrigger value="problems">Bài tập</TabsTrigger>
                    </TabsList>

                    {/* TOPICS TAB */}
                    <TabsContent value="topics" className="space-y-4">
                        <div className="flex items-center justify-between">
                            <p className="text-sm text-muted-foreground">{topics.length} chủ đề</p>
                            {isLecturer && (
                                <CreateTopicDialog
                                    onSuccess={() => {
                                        queryClient.invalidateQueries({ queryKey: ['topics'] })
                                    }}
                                />
                            )}
                        </div>
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                            {loadingTopics ? (
                                <div className="col-span-full flex items-center justify-center py-12">
                                    <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                                </div>
                            ) : topics.length === 0 ? (
                                <div className="col-span-full text-center py-12 text-muted-foreground">
                                    Chưa có chủ đề nào
                                </div>
                            ) : (
                                topics.map((topic) => (
                                    <Card
                                        key={topic.id}
                                        className="cursor-pointer hover:shadow-md transition-shadow"
                                        onClick={() => handleSelectTopic(topic)}
                                    >
                                        <CardHeader>
                                            <CardTitle className="flex items-center gap-2">
                                                {topic.icon && <span className="text-2xl">{topic.icon}</span>}
                                                <span>{topic.name}</span>
                                            </CardTitle>
                                            <CardDescription>{topic.description}</CardDescription>
                                        </CardHeader>
                                    </Card>
                                ))
                            )}
                        </div>
                    </TabsContent>
                    {/* PROBLEMS TAB */}
                    <TabsContent value="problems" className="space-y-4">
                        <div className="flex items-center justify-between gap-4">
                            <div className="flex items-center gap-4">
                                <label className="text-sm font-medium">Độ khó:</label>
                                <Select value={difficultyFilter} onValueChange={setDifficultyFilter}>
                                    <SelectTrigger className="w-[150px]">
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="all">Tất cả</SelectItem>
                                        <SelectItem value="easy">Dễ</SelectItem>
                                        <SelectItem value="medium">Trung bình</SelectItem>
                                        <SelectItem value="hard">Khó</SelectItem>
                                    </SelectContent>
                                </Select>
                                <p className="text-sm text-muted-foreground">{allProblems.length} bài tập</p>
                            </div>
                            {isLecturer && (
                                <CreateProblemDialog
                                    onSuccess={() => {
                                        queryClient.invalidateQueries({ queryKey: ['all-problems'] })
                                        queryClient.invalidateQueries({ queryKey: ['topic-problems'] })
                                    }}
                                />
                            )}
                        </div>
                        <Card>
                            <CardContent className="pt-6">
                                {loadingAllProblems ? (
                                    <div className="flex items-center justify-center py-12">
                                        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                                    </div>
                                ) : allProblems.length === 0 ? (
                                    <div className="text-center py-12 text-muted-foreground">
                                        Chưa có bài tập nào
                                    </div>
                                ) : (
                                    <div className="space-y-2">
                                        {allProblems.map((problem) => (
                                            <div
                                                key={problem.id}
                                                onClick={() => handleSelectProblem(problem)}
                                                className="flex items-center gap-4 p-4 border rounded-lg hover:bg-muted/50 transition-colors cursor-pointer"
                                            >
                                                <div className="flex-1">
                                                    <div className="flex items-center gap-2">
                                                        <span className="font-medium">{problem.title}</span>
                                                        {getDifficultyBadge(problem.difficulty)}
                                                    </div>
                                                    <p className="text-sm text-muted-foreground mt-1 line-clamp-2">
                                                        {problem.description}
                                                    </p>
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                )}
                            </CardContent>
                        </Card>
                    </TabsContent>
                </Tabs>
            </div>
            {/* AI Chat Panel */}
            <AIChatPanel />
        </MainLayout>
    )
}
export const Route = createFileRoute('/practice')({
    component: PracticePage,
})