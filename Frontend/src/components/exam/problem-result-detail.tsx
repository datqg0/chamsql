import { ChevronDown, ChevronUp, Copy, CheckCircle2, AlertCircle, XCircle } from 'lucide-react'
import React, { useState } from 'react'
import toast from 'react-hot-toast'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import type { ProblemSubmissionResult } from '@/types/exam-submission.types'

interface ProblemResultDetailProps {
  problemIndex: number
  problemTitle: string
  submission: ProblemSubmissionResult
}

export const ProblemResultDetail: React.FC<ProblemResultDetailProps> = ({
  problemIndex,
  problemTitle,
  submission,
}) => {
  const [expanded, setExpanded] = useState(false)

  const getStatusIcon = () => {
    switch (submission.status) {
      case 'accepted':
        return <CheckCircle2 className="h-5 w-5 text-green-600" />
      case 'wrong_answer':
        return <XCircle className="h-5 w-5 text-orange-600" />
      case 'error':
        return <AlertCircle className="h-5 w-5 text-red-600" />
      default:
        return null
    }
  }

  const getStatusColor = () => {
    switch (submission.status) {
      case 'accepted':
        return 'bg-green-50 dark:bg-green-950 border-green-200'
      case 'wrong_answer':
        return 'bg-orange-50 dark:bg-orange-950 border-orange-200'
      case 'error':
        return 'bg-red-50 dark:bg-red-950 border-red-200'
      default:
        return 'bg-gray-50 dark:bg-gray-950 border-gray-200'
    }
  }

  const getStatusLabel = () => {
    switch (submission.status) {
      case 'accepted':
        return 'Chính xác'
      case 'wrong_answer':
        return 'Sai'
      case 'error':
        return 'Lỗi thực thi'
      default:
        return 'Không rõ'
    }
  }

  const getStatusBadgeVariant = () => {
    switch (submission.status) {
      case 'accepted':
        return 'default'
      case 'wrong_answer':
        return 'secondary'
      case 'error':
        return 'destructive'
      default:
        return 'outline'
    }
  }

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text)
    toast.success(`Đã copy ${label}`)
  }

  return (
    <Card className={`border-2 ${getStatusColor()} overflow-hidden`}>
      <div
        className="p-4 cursor-pointer hover:opacity-90 transition-opacity"
        onClick={() => setExpanded(!expanded)}
      >
        <div className="flex items-start justify-between gap-4">
          {/* Left: Problem Info */}
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-2">
              {getStatusIcon()}
              <span className="text-sm font-medium text-muted-foreground">
                Câu {problemIndex}
              </span>
              <Badge variant={getStatusBadgeVariant()}>
                {getStatusLabel()}
              </Badge>
            </div>
            <h3 className="font-semibold text-lg mb-1 truncate">{problemTitle}</h3>
            <div className="flex items-center gap-4 text-sm text-muted-foreground">
              <span>
                Điểm: <span className="font-medium">{submission.score}/{submission.maxScore}</span>
              </span>
              <span>
                Lần thử: <span className="font-medium">{submission.attemptNumber}</span>
              </span>
              {submission.executionTimeMs !== undefined && (
                <span>
                  Thời gian: <span className="font-mono font-medium">{submission.executionTimeMs}ms</span>
                </span>
              )}
            </div>
          </div>

          {/* Right: Expand Button */}
          <Button
            variant="ghost"
            size="sm"
            className="flex-shrink-0"
          >
            {expanded ? (
              <ChevronUp className="h-4 w-4" />
            ) : (
              <ChevronDown className="h-4 w-4" />
            )}
          </Button>
        </div>
      </div>

      {/* Expanded Details */}
      {expanded && (
        <CardContent className="pt-0 border-t border-inherit space-y-4">
          {/* Feedback Message */}
          {submission.feedback && (
            <div className="p-3 bg-muted rounded-lg">
              <p className="text-sm font-medium mb-1">Phản hồi:</p>
              <p className="text-sm text-muted-foreground">{submission.feedback}</p>
            </div>
          )}

          {/* Error Message */}
          {submission.errorMessage && (
            <div className="p-3 bg-red-50 dark:bg-red-950 rounded-lg border border-red-200 dark:border-red-800">
              <p className="text-sm font-medium text-red-900 dark:text-red-100 mb-1">
                Lỗi:
              </p>
              <pre className="text-xs text-red-800 dark:text-red-200 overflow-x-auto bg-red-100/30 dark:bg-red-900/30 p-2 rounded">
                {submission.errorMessage}
              </pre>
            </div>
          )}

          {/* Actual Output */}
          {submission.actualOutput && (
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <p className="text-sm font-medium">Kết quả của bạn:</p>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => copyToClipboard(submission.actualOutput!, 'kết quả')}
                  className="h-6 px-2 text-xs"
                >
                  <Copy className="h-3 w-3 mr-1" />
                  Copy
                </Button>
              </div>
              <pre className="text-xs bg-muted p-3 rounded border overflow-x-auto max-h-48 overflow-y-auto font-mono">
                {submission.actualOutput}
              </pre>
            </div>
          )}

          {/* Expected Output */}
          {submission.expectedOutput && (
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <p className="text-sm font-medium">Kết quả đúng:</p>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => copyToClipboard(submission.expectedOutput!, 'kết quả đúng')}
                  className="h-6 px-2 text-xs"
                >
                  <Copy className="h-3 w-3 mr-1" />
                  Copy
                </Button>
              </div>
              <pre className="text-xs bg-green-50 dark:bg-green-950 p-3 rounded border border-green-200 dark:border-green-800 overflow-x-auto max-h-48 overflow-y-auto font-mono">
                {submission.expectedOutput}
              </pre>
            </div>
          )}

          {/* Comparison hint */}
          {submission.actualOutput &&
            submission.expectedOutput &&
            submission.status === 'wrong_answer' && (
            <div className="p-3 bg-blue-50 dark:bg-blue-950 rounded-lg border border-blue-200 dark:border-blue-800">
              <p className="text-sm text-blue-900 dark:text-blue-100">
                💡 Hãy so sánh kết quả của bạn với kết quả đúng để tìm ra lỗi.
              </p>
            </div>
          )}
        </CardContent>
      )}
    </Card>
  )
}

export default ProblemResultDetail
