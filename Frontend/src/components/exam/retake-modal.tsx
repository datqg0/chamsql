import React from 'react'
import { AlertCircle, RotateCw, Loader2 } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

interface RetakeModalProps {
  open: boolean
  onConfirm: () => void | Promise<void>
  onCancel: () => void
  examTitle: string
  currentScore: number
  maxScore: number
  remainingAttempts: number
  maxAttempts: number
  isRetaking?: boolean
}

/**
 * Modal to confirm re-attempting an exam
 * Shows current score and remaining attempts
 */
export const RetakeModal: React.FC<RetakeModalProps> = ({
  open,
  onConfirm,
  onCancel,
  examTitle,
  currentScore,
  maxScore,
  remainingAttempts,
  maxAttempts,
  isRetaking = false,
}) => {
  const percentage = maxScore > 0 ? Math.round((currentScore / maxScore) * 100) : 0
  const isPassing = percentage >= 50

  const handleConfirm = async () => {
    if (!isRetaking) {
      await onConfirm()
    }
  }

  return (
    <Dialog open={open}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <div className="flex items-center gap-2">
            <RotateCw className="h-6 w-6 text-blue-600" />
            <DialogTitle>Làm lại bài thi</DialogTitle>
          </div>
        </DialogHeader>

        <DialogDescription className="space-y-4">
          {/* Exam Info */}
          <div className="p-3 bg-muted rounded-lg">
            <p className="text-sm text-muted-foreground">Bài thi:</p>
            <p className="font-semibold text-foreground">{examTitle}</p>
          </div>

          {/* Current Score */}
          <div className="grid grid-cols-2 gap-2">
            <div className="p-3 bg-blue-50 dark:bg-blue-950 rounded-lg border border-blue-200 dark:border-blue-800">
              <p className="text-xs text-muted-foreground mb-1">Điểm hiện tại</p>
              <p className="text-lg font-bold text-blue-600">
                {currentScore}/{maxScore}
              </p>
              <p className="text-xs text-muted-foreground mt-1">{percentage}%</p>
            </div>

            <div className={`p-3 rounded-lg border-2 ${
              isPassing
                ? 'bg-green-50 dark:bg-green-950 border-green-200 dark:border-green-800'
                : 'bg-orange-50 dark:bg-orange-950 border-orange-200 dark:border-orange-800'
            }`}>
              <p className="text-xs text-muted-foreground mb-1">Trạng thái</p>
              <Badge variant={isPassing ? 'secondary' : 'destructive'}>
                {isPassing ? '✓ Đạt' : '✗ Chưa đạt'}
              </Badge>
            </div>
          </div>

          {/* Remaining Attempts */}
          <div className="p-3 bg-muted rounded-lg">
            <div className="flex items-center justify-between mb-2">
              <p className="text-sm font-medium">Lần thi còn lại</p>
              <Badge variant="outline">
                {remainingAttempts}/{maxAttempts}
              </Badge>
            </div>
            <div className="flex gap-1">
              {Array.from({ length: maxAttempts }).map((_, i) => (
                <div
                  key={i}
                  className={`h-2 flex-1 rounded-full ${
                    i < remainingAttempts
                      ? 'bg-green-500'
                      : 'bg-red-500/30'
                  }`}
                />
              ))}
            </div>
          </div>

          {/* Warning Message */}
          <div className="p-3 bg-blue-50 dark:bg-blue-950 rounded-lg border border-blue-200 dark:border-blue-800 flex gap-3">
            <AlertCircle className="h-5 w-5 text-blue-600 flex-shrink-0 mt-0.5" />
            <div className="text-sm text-blue-900 dark:text-blue-100">
              <p className="font-medium mb-1">Lưu ý:</p>
              <ul className="list-disc list-inside space-y-1 text-xs">
                <li>Điểm mới sẽ thay thế điểm cũ</li>
                <li>Bạn có <strong>{remainingAttempts - 1}</strong> lần thi còn lại sau lần này</li>
                {remainingAttempts === 1 && (
                  <li className="text-orange-600 dark:text-orange-400 font-medium">
                    Đây là lần thi cuối cùng!
                  </li>
                )}
              </ul>
            </div>
          </div>
        </DialogDescription>

        {/* Actions */}
        <div className="flex gap-3 justify-end pt-4">
          <Button 
            variant="outline"
            disabled={isRetaking} 
            onClick={onCancel}
          >
            Hủy
          </Button>
          <Button
            onClick={handleConfirm}
            disabled={isRetaking}
            className="gap-2"
          >
            {isRetaking && <Loader2 className="h-4 w-4 animate-spin" />}
            {isRetaking ? 'Đang bắt đầu...' : 'Làm lại'}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}

export default RetakeModal
