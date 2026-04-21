import { Upload, FileText, X, Loader2, CheckCircle2, ChevronRight, ChevronLeft, Code, Database, Save, Eye } from 'lucide-react'
import { useState, useCallback } from 'react'
import toast from 'react-hot-toast'

import { Button } from '@/components/ui/button'
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import { pdfService, type ExtractedProblem, type ProblemSolution } from '@/services/pdf.service'
import { cn } from '@/lib/utils'

type WizardStep = 'upload' | 'review' | 'solutions' | 'complete'

interface PDFImportWizardProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    onSuccess?: () => void
}

const ACCEPTED_EXTENSIONS = ['.pdf']

export function PDFImportWizard({ open, onOpenChange, onSuccess }: PDFImportWizardProps) {
    const [step, setStep] = useState<WizardStep>('upload')
    const [file, setFile] = useState<File | null>(null)
    const [isUploading, setIsUploading] = useState(false)
    const [isDragOver, setIsDragOver] = useState(false)
    const [_uploadId, setUploadId] = useState<number | null>(null)
    const [problems, setProblems] = useState<ExtractedProblem[]>([])
    const [currentProblemIndex, setCurrentProblemIndex] = useState(0)
    const [solutions, setSolutions] = useState<Record<number, ProblemSolution>>({})

    const handleDragOver = useCallback((e: React.DragEvent) => {
        e.preventDefault()
        setIsDragOver(true)
    }, [])

    const handleDragLeave = useCallback((e: React.DragEvent) => {
        e.preventDefault()
        setIsDragOver(false)
    }, [])

    const handleDrop = useCallback((e: React.DragEvent) => {
        e.preventDefault()
        setIsDragOver(false)

        const droppedFile = e.dataTransfer.files[0]
        if (droppedFile && isValidFile(droppedFile)) {
            setFile(droppedFile)
        } else {
            toast.error('Chỉ chấp nhận file PDF')
        }
    }, [])

    const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const selectedFile = e.target.files?.[0]
        if (selectedFile && isValidFile(selectedFile)) {
            setFile(selectedFile)
        } else if (selectedFile) {
            toast.error('Chỉ chấp nhận file PDF')
        }
    }

    const isValidFile = (file: File): boolean => {
        const extension = '.' + file.name.split('.').pop()?.toLowerCase()
        return ACCEPTED_EXTENSIONS.includes(extension)
    }

    const handleUpload = async () => {
        if (!file) return

        setIsUploading(true)
        try {
            const result = await pdfService.uploadPDF(file)
            setUploadId(result.id)
            toast.success('Upload thành công! Đang trích xuất câu hỏi...')

            // Poll for extraction completion
            await pollForExtraction(result.id)
        } catch (error: any) {
            toast.error(error?.message || 'Upload thất bại')
        } finally {
            setIsUploading(false)
        }
    }

    const pollForExtraction = async (id: number) => {
        const maxAttempts = 30
        let attempts = 0

        while (attempts < maxAttempts) {
            await new Promise(resolve => setTimeout(resolve, 2000))
            const status = await pdfService.getUploadStatus(id)

            if (status.status === 'completed') {
                const extractedProblems = await pdfService.getExtractedProblems(id)
                setProblems(extractedProblems)
                setStep('review')
                toast.success(`Trích xuất thành công ${extractedProblems.length} câu hỏi!`)
                return
            } else if (status.status === 'failed') {
                toast.error('Trích xuất thất bại: ' + status.error_message)
                return
            }

            attempts++
        }

        toast.error('Quá thời gian chờ trích xuất')
    }

    const handleNext = () => {
        if (step === 'review') {
            setStep('solutions')
        } else if (step === 'solutions') {
            handleSaveAll()
        }
    }

    const handleBack = () => {
        if (step === 'solutions') {
            setStep('review')
        } else if (step === 'complete') {
            setStep('solutions')
        }
    }

    const handleSaveSolution = async (problemId: number) => {
        const solution = solutions[problemId]
        if (!solution?.solution_query) {
            toast.error('Vui lòng nhập solution query')
            return
        }

        try {
            await pdfService.updateSolution(problemId, solution)
            toast.success('Đã lưu đáp án!')
        } catch (error: any) {
            toast.error(error?.message || 'Lưu đáp án thất bại')
        }
    }

    const handleSaveAll = async () => {
        const unsavedProblems = problems.filter(p => !solutions[p.id]?.solution_query)
        if (unsavedProblems.length > 0) {
            toast.error(`Còn ${unsavedProblems.length} câu chưa nhập đáp án`)
            return
        }

        try {
            for (const problem of problems) {
                await pdfService.updateSolution(problem.id, solutions[problem.id])
            }
            setStep('complete')
            toast.success('Tất cả câu hỏi đã được lưu!')
            onSuccess?.()
        } catch (error: any) {
            toast.error(error?.message || 'Lưu thất bại')
        }
    }

    const updateSolution = (problemId: number, field: keyof ProblemSolution, value: string) => {
        setSolutions(prev => ({
            ...prev,
            [problemId]: {
                ...prev[problemId],
                [field]: value
            }
        }))
    }

    const reset = () => {
        setStep('upload')
        setFile(null)
        setUploadId(null)
        setProblems([])
        setSolutions({})
        setCurrentProblemIndex(0)
    }

    const getFileIcon = () => '📕'

    const formatFileSize = (bytes: number): string => {
        if (bytes < 1024) return bytes + ' B'
        if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
        return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
    }

    return (
        <Dialog open={open} onOpenChange={(newOpen) => {
            if (!newOpen) reset()
            onOpenChange(newOpen)
        }}>
            <DialogContent className="sm:max-w-[700px] max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>
                        {step === 'upload' && 'Import đề thi từ PDF'}
                        {step === 'review' && 'Xem trước câu hỏi đã trích xuất'}
                        {step === 'solutions' && 'Nhập đáp án cho từng câu'}
                        {step === 'complete' && 'Hoàn tất!'}
                    </DialogTitle>
                    <DialogDescription>
                        {step === 'upload' && 'Tải lên file PDF để trích xuất câu hỏi (giống Codeforces - chỉ có đề bài)'}
                        {step === 'review' && 'Kiểm tra các câu hỏi đã được trích xuất từ PDF'}
                        {step === 'solutions' && 'Nhập solution query SQL cho từng câu hỏi'}
                        {step === 'complete' && 'Tất cả câu hỏi đã được lưu vào ngân hàng câu hỏi'}
                    </DialogDescription>
                </DialogHeader>

                {/* Step 1: Upload */}
                {step === 'upload' && (
                    <div className="space-y-4 py-4">
                        <div
                            onDragOver={handleDragOver}
                            onDragLeave={handleDragLeave}
                            onDrop={handleDrop}
                            className={cn(
                                "border-2 border-dashed rounded-lg p-8 text-center transition-colors cursor-pointer",
                                isDragOver ? 'border-primary bg-primary/5' : 'border-muted-foreground/25 hover:border-primary/50',
                                file ? 'bg-muted/50' : ''
                            )}
                        >
                            {file ? (
                                <div className="flex items-center justify-between gap-4">
                                    <div className="flex items-center gap-3">
                                        <span className="text-3xl">{getFileIcon()}</span>
                                        <div className="text-left">
                                            <p className="font-medium truncate max-w-[300px]">{file.name}</p>
                                            <p className="text-sm text-muted-foreground">{formatFileSize(file.size)}</p>
                                        </div>
                                    </div>
                                    <Button variant="ghost" size="icon" onClick={() => setFile(null)} disabled={isUploading}>
                                        <X className="h-4 w-4" />
                                    </Button>
                                </div>
                            ) : (
                                <>
                                    <FileText className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
                                    <p className="text-sm font-medium">Kéo thả file PDF vào đây hoặc click để chọn</p>
                                    <p className="text-xs text-muted-foreground mt-2">Chỉ hỗ trợ file PDF</p>
                                </>
                            )}
                            <Input
                                type="file"
                                accept=".pdf"
                                onChange={handleFileChange}
                                className="hidden"
                                id="file-upload"
                                disabled={isUploading}
                            />
                            {!file && <Label htmlFor="file-upload" className="absolute inset-0 cursor-pointer" />}
                        </div>

                        <div className="bg-blue-500/10 rounded-lg p-3 text-sm text-blue-600 dark:text-blue-400">
                            <p className="font-medium mb-1">Lưu ý về định dạng Codeforces:</p>
                            <ul className="list-disc list-inside space-y-1 text-xs">
                                <li>PDF chỉ chứa đề bài, không có đáp án</li>
                                <li>Hệ thống trích xuất mô tả câu hỏi</li>
                                <li>Giảng viên cần nhập solution SQL ở bước tiếp theo</li>
                            </ul>
                        </div>

                        <div className="flex gap-2">
                            <Button variant="outline" className="flex-1" onClick={() => onOpenChange(false)} disabled={isUploading}>
                                Hủy
                            </Button>
                            <Button className="flex-1" onClick={handleUpload} disabled={!file || isUploading}>
                                {isUploading ? (
                                    <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Đang xử lý...</>
                                ) : (
                                    <><Upload className="mr-2 h-4 w-4" /> Tải lên và trích xuất</>
                                )}
                            </Button>
                        </div>
                    </div>
                )}

                {/* Step 2: Review Problems */}
                {step === 'review' && (
                    <div className="space-y-4 py-4">
                        <div className="flex items-center justify-between">
                            <span className="text-sm text-muted-foreground">
                                Tìm thấy {problems.length} câu hỏi
                            </span>
                            <Button variant="outline" size="sm" onClick={() => setStep('solutions')}>
                                Tiếp tục <ChevronRight className="ml-1 h-4 w-4" />
                            </Button>
                        </div>

                        <div className="space-y-3 max-h-[400px] overflow-y-auto">
                            {problems.map((problem, idx) => (
                                <div key={problem.id} className="border rounded-lg p-4">
                                    <div className="flex items-start justify-between">
                                        <div className="flex-1">
                                            <div className="flex items-center gap-2 mb-2">
                                                <span className="bg-primary/10 text-primary px-2 py-0.5 rounded text-xs font-medium">
                                                    Câu {idx + 1}
                                                </span>
                                                <span className="text-xs text-muted-foreground capitalize">
                                                    {problem.difficulty}
                                                </span>
                                            </div>
                                            <h4 className="font-medium mb-1">{problem.title}</h4>
                                            <p className="text-sm text-muted-foreground line-clamp-3">
                                                {problem.description}
                                            </p>
                                            {problem.init_script && (
                                                <div className="mt-2 text-xs text-muted-foreground">
                                                    <span className="font-medium">Schema:</span> Có script khởi tạo
                                                </div>
                                            )}
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>

                        <div className="flex gap-2">
                            <Button variant="outline" className="flex-1" onClick={handleBack}>
                                <ChevronLeft className="mr-2 h-4 w-4" /> Quay lại
                            </Button>
                            <Button className="flex-1" onClick={handleNext}>
                                Tiếp tục <ChevronRight className="ml-2 h-4 w-4" />
                            </Button>
                        </div>
                    </div>
                )}

                {/* Step 3: Input Solutions */}
                {step === 'solutions' && problems.length > 0 && (
                    <div className="space-y-4 py-4">
                        <div className="flex items-center justify-between">
                            <div className="flex items-center gap-2">
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={() => setCurrentProblemIndex(Math.max(0, currentProblemIndex - 1))}
                                    disabled={currentProblemIndex === 0}
                                >
                                    <ChevronLeft className="h-4 w-4" />
                                </Button>
                                <span className="text-sm font-medium">
                                    Câu {currentProblemIndex + 1} / {problems.length}
                                </span>
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={() => setCurrentProblemIndex(Math.min(problems.length - 1, currentProblemIndex + 1))}
                                    disabled={currentProblemIndex === problems.length - 1}
                                >
                                    <ChevronRight className="h-4 w-4" />
                                </Button>
                            </div>
                            <div className="text-xs text-muted-foreground">
                                {Object.keys(solutions).length} / {problems.length} câu đã nhập
                            </div>
                        </div>

                        {(() => {
                            const problem = problems[currentProblemIndex]
                            const solution = solutions[problem.id] || { solution_query: '', db_type: 'postgresql' }

                            return (
                                <div className="space-y-4">
                                    <div className="border rounded-lg p-4 bg-muted/30">
                                        <div className="flex items-center gap-2 mb-2">
                                            <Eye className="h-4 w-4 text-muted-foreground" />
                                            <span className="text-sm font-medium">Đề bài</span>
                                        </div>
                                        <h4 className="font-medium mb-2">{problem.title}</h4>
                                        <p className="text-sm text-muted-foreground whitespace-pre-wrap">
                                            {problem.description}
                                        </p>
                                        {problem.init_script && (
                                            <div className="mt-3">
                                                <span className="text-xs font-medium">Schema SQL:</span>
                                                <pre className="mt-1 text-xs bg-muted p-2 rounded overflow-x-auto">
                                                    {problem.init_script}
                                                </pre>
                                            </div>
                                        )}
                                    </div>

                                    <div className="space-y-3">
                                        <div className="flex items-center gap-2">
                                            <Database className="h-4 w-4 text-muted-foreground" />
                                            <Label>Loại Database</Label>
                                        </div>
                                        <Select
                                            value={solution.db_type}
                                            onValueChange={(v) => updateSolution(problem.id, 'db_type', v)}
                                        >
                                            <SelectTrigger>
                                                <SelectValue />
                                            </SelectTrigger>
                                            <SelectContent>
                                                <SelectItem value="postgresql">PostgreSQL</SelectItem>
                                                <SelectItem value="mysql">MySQL</SelectItem>
                                                <SelectItem value="sqlserver">SQL Server</SelectItem>
                                            </SelectContent>
                                        </Select>

                                        <div className="flex items-center gap-2">
                                            <Code className="h-4 w-4 text-muted-foreground" />
                                            <Label>Solution Query (Đáp án)</Label>
                                        </div>
                                        <Textarea
                                            value={solution.solution_query}
                                            onChange={(e) => updateSolution(problem.id, 'solution_query', e.target.value)}
                                            placeholder="SELECT * FROM ..."
                                            rows={6}
                                            className="font-mono text-sm"
                                        />
                                        <p className="text-xs text-muted-foreground">
                                            Nhập câu query SQL đúng để so sánh với đáp án sinh viên
                                        </p>

                                        <Button
                                            onClick={() => handleSaveSolution(problem.id)}
                                            disabled={!solution.solution_query}
                                            className="w-full"
                                        >
                                            <Save className="mr-2 h-4 w-4" />
                                            Lưu đáp án
                                        </Button>
                                    </div>
                                </div>
                            )
                        })()}

                        <div className="flex gap-2">
                            <Button variant="outline" className="flex-1" onClick={handleBack}>
                                <ChevronLeft className="mr-2 h-4 w-4" /> Quay lại
                            </Button>
                            <Button className="flex-1" onClick={handleSaveAll}>
                                <CheckCircle2 className="mr-2 h-4 w-4" />
                                Hoàn tất & Lưu tất cả
                            </Button>
                        </div>
                    </div>
                )}

                {/* Step 4: Complete */}
                {step === 'complete' && (
                    <div className="space-y-4 py-8 text-center">
                        <div className="w-16 h-16 bg-green-100 dark:bg-green-900/20 rounded-full flex items-center justify-center mx-auto">
                            <CheckCircle2 className="h-8 w-8 text-green-600" />
                        </div>
                        <div>
                            <h3 className="text-lg font-medium mb-1">Import thành công!</h3>
                            <p className="text-sm text-muted-foreground">
                                {problems.length} câu hỏi đã được lưu vào ngân hàng câu hỏi
                            </p>
                        </div>
                        <div className="flex gap-2 justify-center">
                            <Button variant="outline" onClick={() => onOpenChange(false)}>
                                Đóng
                            </Button>
                            <Button onClick={() => {
                                reset()
                                setStep('upload')
                            }}>
                                Import file khác
                            </Button>
                        </div>
                    </div>
                )}
            </DialogContent>
        </Dialog>
    )
}
