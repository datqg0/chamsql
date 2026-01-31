import { useState, useCallback } from 'react'
import { Upload, FileText, X, Loader2, CheckCircle2 } from 'lucide-react'

import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { examsService } from '@/services/exams.service'
import toast from 'react-hot-toast'

interface ExamImportDialogProps {
    onSuccess?: () => void
    trigger?: React.ReactNode
}

const ACCEPTED_FILE_TYPES = {
    'application/pdf': ['.pdf'],
    'application/vnd.openxmlformats-officedocument.wordprocessingml.document': ['.docx'],
    'application/msword': ['.doc'],
    'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet': ['.xlsx'],
    'application/vnd.ms-excel': ['.xls'],
}

const ACCEPTED_EXTENSIONS = ['.pdf', '.doc', '.docx', '.xls', '.xlsx']

export function ExamImportDialog({ onSuccess, trigger }: ExamImportDialogProps) {
    const [open, setOpen] = useState(false)
    const [file, setFile] = useState<File | null>(null)
    const [isUploading, setIsUploading] = useState(false)
    const [isDragOver, setIsDragOver] = useState(false)

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
            toast.error('Chỉ chấp nhận file PDF, Word hoặc Excel')
        }
    }, [])

    const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const selectedFile = e.target.files?.[0]
        if (selectedFile && isValidFile(selectedFile)) {
            setFile(selectedFile)
        } else if (selectedFile) {
            toast.error('Chỉ chấp nhận file PDF, Word hoặc Excel')
        }
    }

    const isValidFile = (file: File): boolean => {
        const extension = '.' + file.name.split('.').pop()?.toLowerCase()
        return ACCEPTED_EXTENSIONS.includes(extension)
    }

    const getFileIcon = (fileName: string) => {
        const ext = fileName.split('.').pop()?.toLowerCase()
        switch (ext) {
            case 'pdf':
                return '📕'
            case 'doc':
            case 'docx':
                return '📘'
            case 'xls':
            case 'xlsx':
                return '📗'
            default:
                return '📄'
        }
    }

    const formatFileSize = (bytes: number): string => {
        if (bytes < 1024) return bytes + ' B'
        if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
        return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
    }

    const handleUpload = async () => {
        if (!file) return

        setIsUploading(true)
        try {
            const result = await examsService.importExamFile(file)
            if (result.success) {
                toast.success('Import đề thi thành công!')
                setFile(null)
                setOpen(false)
                onSuccess?.()
            } else {
                toast.error(result.message || 'Có lỗi xảy ra khi import')
            }
        } catch (error: any) {
            toast.error(error?.message || 'Có lỗi xảy ra khi upload file')
        } finally {
            setIsUploading(false)
        }
    }

    const handleRemoveFile = () => {
        setFile(null)
    }

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger asChild>
                {trigger || (
                    <Button>
                        <Upload className="mr-2 h-4 w-4" />
                        Import đề thi
                    </Button>
                )}
            </DialogTrigger>
            <DialogContent className="sm:max-w-[500px]">
                <DialogHeader>
                    <DialogTitle>Import đề thi</DialogTitle>
                    <DialogDescription>
                        Tải lên file đề thi dạng PDF, Word hoặc Excel để tạo bài thi mới
                    </DialogDescription>
                </DialogHeader>

                <div className="space-y-4 py-4">
                    {/* Drop Zone */}
                    <div
                        onDragOver={handleDragOver}
                        onDragLeave={handleDragLeave}
                        onDrop={handleDrop}
                        className={`
                            border-2 border-dashed rounded-lg p-8 text-center transition-colors cursor-pointer
                            ${isDragOver
                                ? 'border-primary bg-primary/5'
                                : 'border-muted-foreground/25 hover:border-primary/50'
                            }
                            ${file ? 'bg-muted/50' : ''}
                        `}
                    >
                        {file ? (
                            <div className="flex items-center justify-between gap-4">
                                <div className="flex items-center gap-3">
                                    <span className="text-3xl">{getFileIcon(file.name)}</span>
                                    <div className="text-left">
                                        <p className="font-medium truncate max-w-[250px]">{file.name}</p>
                                        <p className="text-sm text-muted-foreground">
                                            {formatFileSize(file.size)}
                                        </p>
                                    </div>
                                </div>
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    onClick={handleRemoveFile}
                                    disabled={isUploading}
                                >
                                    <X className="h-4 w-4" />
                                </Button>
                            </div>
                        ) : (
                            <>
                                <FileText className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
                                <p className="text-sm font-medium">
                                    Kéo thả file vào đây hoặc click để chọn
                                </p>
                                <p className="text-xs text-muted-foreground mt-2">
                                    Hỗ trợ: PDF, Word (.doc, .docx), Excel (.xls, .xlsx)
                                </p>
                            </>
                        )}
                        <Input
                            type="file"
                            accept={ACCEPTED_EXTENSIONS.join(',')}
                            onChange={handleFileChange}
                            className="hidden"
                            id="file-upload"
                            disabled={isUploading}
                        />
                        {!file && (
                            <Label
                                htmlFor="file-upload"
                                className="absolute inset-0 cursor-pointer"
                            />
                        )}
                    </div>

                    {/* Info */}
                    <div className="bg-blue-500/10 rounded-lg p-3 text-sm text-blue-600 dark:text-blue-400">
                        <p className="font-medium mb-1">Lưu ý:</p>
                        <ul className="list-disc list-inside space-y-1 text-xs">
                            <li>File sẽ được xử lý bởi hệ thống để trích xuất các câu hỏi</li>
                            <li>Đề thi nên có định dạng rõ ràng với các câu hỏi được đánh số</li>
                            <li>File Excel nên có mỗi câu hỏi trên một dòng</li>
                        </ul>
                    </div>

                    {/* Actions */}
                    <div className="flex gap-2">
                        <Button
                            variant="outline"
                            className="flex-1"
                            onClick={() => setOpen(false)}
                            disabled={isUploading}
                        >
                            Hủy
                        </Button>
                        <Button
                            className="flex-1"
                            onClick={handleUpload}
                            disabled={!file || isUploading}
                        >
                            {isUploading ? (
                                <>
                                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                    Đang tải lên...
                                </>
                            ) : (
                                <>
                                    <CheckCircle2 className="mr-2 h-4 w-4" />
                                    Import đề thi
                                </>
                            )}
                        </Button>
                    </div>
                </div>
            </DialogContent>
        </Dialog>
    )
}
