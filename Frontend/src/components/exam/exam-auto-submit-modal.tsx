import { AlertCircle, Clock, CheckCircle2, Loader2 } from 'lucide-react'
import React from 'react'

import { Button } from '@/components/ui/button'
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog'

interface ExamAutoSubmitModalProps {
    open: boolean
    onConfirm: () => void | Promise<void>
    onCancel?: () => void
    examTitle?: string
    timeRemaining?: string
    isSubmitting?: boolean
    type?: 'auto-submit' | 'manual-submit'
}

/**
 * Modal to confirm exam submission
 * Shows when:
 * 1. Time expires (auto-submit)
 * 2. Student clicks "Nộp bài" button (manual submit)
 */
export const ExamAutoSubmitModal: React.FC<ExamAutoSubmitModalProps> = ({
    open,
    onConfirm,
    onCancel,
    examTitle = 'Bài thi',
    timeRemaining,
    isSubmitting = false,
    type = 'manual-submit',
}) => {
    const isAutoSubmit = type === 'auto-submit'

    const handleConfirm = async () => {
        if (!isSubmitting) {
            await onConfirm()
        }
    }

    return (
        <Dialog open={open}>
            <DialogContent className="max-w-md">
                <DialogHeader>
                    <div className="flex items-center gap-2">
                        {isAutoSubmit ? (
                            <AlertCircle className="h-6 w-6 text-red-500" />
                        ) : (
                            <CheckCircle2 className="h-6 w-6 text-blue-500" />
                        )}
                        <DialogTitle>
                            {isAutoSubmit ? 'Hết giờ làm bài' : 'Xác nhận nộp bài'}
                        </DialogTitle>
                    </div>
                </DialogHeader>

                <DialogDescription className="space-y-4">
                    {/* Exam Title */}
                    <div className="p-3 bg-muted rounded-lg">
                        <p className="text-sm text-muted-foreground">Bài thi:</p>
                        <p className="font-semibold text-foreground">{examTitle}</p>
                    </div>

                    {/* Auto-Submit Message */}
                    {isAutoSubmit && (
                        <div className="p-3 bg-red-50 dark:bg-red-950 rounded-lg border border-red-200 dark:border-red-800">
                            <div className="flex items-start gap-2">
                                <Clock className="h-5 w-5 text-red-600 flex-shrink-0 mt-0.5" />
                                <div>
                                    <p className="font-medium text-red-900 dark:text-red-100">
                                        Thời gian làm bài của bạn đã hết!
                                    </p>
                                    <p className="text-sm text-red-800 dark:text-red-200 mt-1">
                                        Bài làm của bạn sẽ được tự động nộp ngay bây giờ.
                                    </p>
                                </div>
                            </div>
                        </div>
                    )}

                    {/* Manual Submit Message */}
                    {!isAutoSubmit && (
                        <div className="p-3 bg-blue-50 dark:bg-blue-950 rounded-lg border border-blue-200 dark:border-blue-800">
                            <p className="text-sm text-blue-900 dark:text-blue-100">
                                Bạn chắc chắn muốn nộp bài? Bạn sẽ không thể chỉnh sửa sau khi nộp.
                            </p>
                        </div>
                    )}

                    {/* Time Remaining (if available) */}
                    {timeRemaining && !isAutoSubmit && (
                        <div className="flex items-center gap-2 text-sm">
                            <Clock className="h-4 w-4 text-muted-foreground" />
                            <span className="text-muted-foreground">
                                Thời gian còn lại: <span className="font-mono font-semibold">{timeRemaining}</span>
                            </span>
                        </div>
                    )}

                    {/* Warning for unsaved answers */}
                    {!isAutoSubmit && (
                        <p className="text-xs text-muted-foreground">
                            Tất cả câu trả lời đã được lưu và sẽ được chấm điểm tự động.
                        </p>
                     )}
                </DialogDescription>

                {/* Actions */}
                <div className="flex gap-3 justify-end pt-4">
                    {!isAutoSubmit && (
                        <Button 
                            variant="outline"
                            disabled={isSubmitting} 
                            onClick={onCancel}
                        >
                            Hủy
                        </Button>
                    )}
                    <Button
                        onClick={handleConfirm}
                        disabled={isSubmitting}
                        variant={isAutoSubmit ? 'destructive' : 'default'}
                        className="gap-2"
                    >
                        {isSubmitting && <Loader2 className="h-4 w-4 animate-spin" />}
                        {isAutoSubmit
                            ? 'Nộp bài ngay'
                            : isSubmitting
                              ? 'Đang nộp...'
                              : 'Nộp bài'}
                    </Button>
                </div>
            </DialogContent>
        </Dialog>
    )
}

export default ExamAutoSubmitModal
