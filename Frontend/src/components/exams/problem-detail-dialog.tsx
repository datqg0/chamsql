import { useQuery } from '@tanstack/react-query'
import {
    BookOpen,
    Code,
    Database,
    Loader2,
    Eye,
} from 'lucide-react'
import ReactMarkdown from 'react-markdown'
import rehypeRaw from 'rehype-raw'
import remarkGfm from 'remark-gfm'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from '@/components/ui/dialog'
import { problemsService } from '@/services/problems.service'

interface ProblemDetailDialogProps {
    problemSlug: string
    problemTitle: string
}

export function ProblemDetailDialog({ problemSlug, problemTitle }: ProblemDetailDialogProps) {
    const { data: problem, isLoading } = useQuery({
        queryKey: ['problem', problemSlug],
        queryFn: () => problemsService.getBySlug(problemSlug),
        enabled: !!problemSlug,
    })

    const getDifficultyBadge = (difficulty: string) => {
        const config: Record<string, { label: string; className: string }> = {
            easy: { label: 'Dễ', className: 'bg-green-500/10 text-green-600' },
            medium: { label: 'Trung bình', className: 'bg-yellow-500/10 text-yellow-600' },
            hard: { label: 'Khó', className: 'bg-red-500/10 text-red-600' },
        }
        return (
            <Badge variant="secondary" className={config[difficulty]?.className}>
                {config[difficulty]?.label || difficulty}
            </Badge>
        )
    }

    return (
        <Dialog>
            <DialogTrigger asChild>
                <Button
                    variant="ghost"
                    size="sm"
                    className="text-indigo-600 hover:text-indigo-700 hover:bg-indigo-50 transition-all duration-200"
                    onClick={(e) => e.stopPropagation()}
                >
                    <Eye className="h-4 w-4 mr-1" />
                    <span className="text-xs font-medium">Xem đề</span>
                </Button>
            </DialogTrigger>
            <DialogContent className="max-w-4xl max-h-[90vh] overflow-hidden flex flex-col p-0 border-none shadow-2xl bg-slate-50/95 backdrop-blur-xl">
                {/* Header with Gradient */}
                <div className="bg-gradient-to-r from-indigo-600 to-violet-600 p-6 text-white shrink-0">
                    <DialogHeader>
                        <div className="flex items-center justify-between">
                            <div className="space-y-1">
                                <div className="flex items-center gap-3">
                                    <DialogTitle className="text-2xl font-bold text-white tracking-tight">
                                        {problemTitle}
                                    </DialogTitle>
                                    {problem && getDifficultyBadge(problem.difficulty)}
                                </div>
                                <p className="text-indigo-100/80 text-sm flex items-center gap-2">
                                    <Database className="h-3.5 w-3.5" />
                                    Chi tiết bài tập trong hệ thống
                                </p>
                            </div>
                        </div>
                    </DialogHeader>
                </div>

                {isLoading ? (
                    <div className="flex-1 flex flex-col items-center justify-center py-20 bg-white/50">
                        <Loader2 className="h-10 w-10 animate-spin text-indigo-600 mb-4" />
                        <p className="text-slate-500 font-medium animate-pulse">Đang tải dữ liệu bài tập...</p>
                    </div>
                ) : problem ? (
                    <div className="flex-1 overflow-y-auto p-6 space-y-8 custom-scrollbar">
                        {/* Description Section */}
                        <section className="space-y-3 animate-in fade-in slide-in-from-bottom-4 duration-500">
                            <div className="flex items-center gap-2 text-indigo-600 font-bold uppercase tracking-wider text-xs">
                                <BookOpen className="h-4 w-4" />
                                <h3>Mô tả đề bài</h3>
                            </div>
                            <div className="relative group">
                                <div className="absolute -inset-0.5 bg-gradient-to-r from-indigo-500 to-violet-500 rounded-xl blur opacity-10 group-hover:opacity-20 transition duration-300"></div>
                                <div className="relative prose prose-slate dark:prose-invert max-w-none p-5 bg-white rounded-xl border border-slate-200 shadow-sm leading-relaxed">
                                    <ReactMarkdown
                                        remarkPlugins={[remarkGfm]}
                                        rehypePlugins={[rehypeRaw]}
                                    >
                                        {problem.description || "*Không có mô tả cho bài tập này.*"}
                                    </ReactMarkdown>
                                </div>
                            </div>
                        </section>

                        {/* Solution Query Section */}
                        <section className="space-y-3 animate-in fade-in slide-in-from-bottom-4 duration-500 delay-150">
                            <div className="flex items-center gap-2 text-violet-600 font-bold uppercase tracking-wider text-xs">
                                <Code className="h-4 w-4" />
                                <h3>Câu truy vấn mẫu (Solution)</h3>
                            </div>
                            <div className="relative rounded-xl overflow-hidden shadow-inner bg-slate-900 border border-slate-800">
                                <div className="absolute top-0 left-0 right-0 h-8 bg-slate-800/50 flex items-center px-4 gap-1.5 border-b border-slate-700/50">
                                    <div className="w-2.5 h-2.5 rounded-full bg-red-500/80"></div>
                                    <div className="w-2.5 h-2.5 rounded-full bg-amber-500/80"></div>
                                    <div className="w-2.5 h-2.5 rounded-full bg-emerald-500/80"></div>
                                    <span className="ml-2 text-[10px] text-slate-400 font-mono uppercase tracking-widest">sql_solution.sql</span>
                                </div>
                                <div className="p-6 pt-12 min-h-[100px] flex flex-col justify-center">
                                    {problem.solutionQuery ? (
                                        <pre className="text-indigo-300 font-mono text-sm whitespace-pre-wrap leading-relaxed selection:bg-indigo-500/30">
                                            <code>{problem.solutionQuery}</code>
                                        </pre>
                                    ) : (
                                        <div className="text-center space-y-2 py-4">
                                            <p className="text-slate-500 italic text-sm">Bài tập này chưa có câu truy vấn mẫu.</p>
                                            <Button variant="outline" size="sm" className="bg-slate-800 border-slate-700 text-slate-300 hover:bg-slate-700">
                                                Cập nhật ngay
                                            </Button>
                                        </div>
                                    )}
                                </div>
                            </div>
                        </section>

                        {/* Databases & Meta */}
                        <section className="pt-4 flex flex-wrap items-center justify-between gap-4 border-t border-slate-200 animate-in fade-in slide-in-from-bottom-4 duration-500 delay-300">
                            <div className="space-y-2">
                                <span className="text-[10px] font-bold text-slate-400 uppercase tracking-widest">CSDL hỗ trợ</span>
                                <div className="flex gap-2">
                                    {problem.supportedDatabases && problem.supportedDatabases.length > 0 ? (
                                        problem.supportedDatabases.map((db) => (
                                            <Badge key={db} variant="secondary" className="bg-slate-200 text-slate-700 hover:bg-indigo-100 hover:text-indigo-700 transition-colors uppercase font-mono text-[10px]">
                                                {db}
                                            </Badge>
                                        ))
                                    ) : (
                                        <span className="text-xs text-slate-400 italic">Không giới hạn</span>
                                    )}
                                </div>
                            </div>
                            <div className="flex gap-2">
                                <Button variant="outline" size="sm" className="rounded-full px-4 h-9 text-xs font-semibold border-slate-300 hover:bg-white hover:text-indigo-600 transition-all shadow-sm">
                                    In đề bài
                                </Button>
                                <Button size="sm" className="rounded-full px-6 h-9 text-xs font-bold bg-indigo-600 hover:bg-indigo-700 shadow-md shadow-indigo-200 transition-all">
                                    Chạy thử
                                </Button>
                            </div>
                        </section>
                    </div>
                ) : (
                    <div className="flex-1 flex flex-col items-center justify-center py-20 text-slate-400 space-y-4">
                        <div className="p-4 bg-slate-100 rounded-full">
                            <BookOpen className="h-10 w-10 opacity-20" />
                        </div>
                        <p className="font-medium">Rất tiếc, không tìm thấy thông tin bài tập.</p>
                        <Button variant="link" className="text-indigo-600 font-bold underline decoration-indigo-200 underline-offset-4">Thử tải lại trang</Button>
                    </div>
                )}
            </DialogContent>
        </Dialog>
    )
}
