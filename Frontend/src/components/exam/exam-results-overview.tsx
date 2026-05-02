import { Trophy, TrendingUp, Clock } from 'lucide-react'
import React from 'react'

import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'

interface ExamResultsOverviewProps {
  totalScore: number
  maxScore: number
  submittedAt?: string
  attemptNumber?: number
}

export const ExamResultsOverview: React.FC<ExamResultsOverviewProps> = ({
  totalScore,
  maxScore,
  submittedAt,
  attemptNumber = 1,
}) => {
  const percentage = maxScore > 0 ? Math.round((totalScore / maxScore) * 100) : 0
  
  // Determine pass/fail status (typically 50% is passing)
  const isPassing = percentage >= 50
  const statusColor = isPassing ? 'text-green-600' : 'text-red-600'
  const statusBg = isPassing ? 'bg-green-50 dark:bg-green-950' : 'bg-red-50 dark:bg-red-950'
  const statusBorder = isPassing ? 'border-green-200' : 'border-red-200'

  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
      {/* Total Score Card */}
      <Card className={`${statusBg} border-2 ${statusBorder}`}>
        <CardContent className="pt-6">
          <div className="flex items-center gap-3">
            <Trophy className={`h-8 w-8 ${statusColor}`} />
            <div>
              <p className="text-sm text-muted-foreground">Điểm số</p>
              <p className={`text-3xl font-bold ${statusColor}`}>
                {totalScore}/{maxScore}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Percentage Card */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex items-center gap-3">
            <TrendingUp className="h-8 w-8 text-blue-600" />
            <div>
              <p className="text-sm text-muted-foreground">Tỷ lệ đúng</p>
              <div className="flex items-center gap-2">
                <p className="text-3xl font-bold text-blue-600">{percentage}%</p>
                <Badge variant={isPassing ? 'secondary' : 'destructive'}>
                  {isPassing ? '✓ Đạt' : '✗ Chưa đạt'}
                </Badge>
              </div>
            </div>
          </div>

          {/* Progress bar */}
          <div className="mt-3 h-2 bg-muted rounded-full overflow-hidden">
            <div
              className={`h-full transition-all ${
                isPassing ? 'bg-green-600' : 'bg-red-600'
              }`}
              style={{ width: `${percentage}%` }}
            />
          </div>
        </CardContent>
      </Card>

      {/* Metadata Card */}
      <Card>
        <CardContent className="pt-6">
          <div className="space-y-3 text-sm">
            {submittedAt && (
              <div className="flex items-start gap-2">
                <Clock className="h-4 w-4 text-muted-foreground mt-0.5 flex-shrink-0" />
                <div>
                  <p className="text-muted-foreground">Thời gian nộp</p>
                  <p className="font-medium">
                    {new Date(submittedAt).toLocaleString('vi-VN', {
                      year: 'numeric',
                      month: '2-digit',
                      day: '2-digit',
                      hour: '2-digit',
                      minute: '2-digit',
                    })}
                  </p>
                </div>
              </div>
            )}
            <div>
              <p className="text-muted-foreground">Lần thi thứ</p>
              <p className="font-medium">{attemptNumber}</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

export default ExamResultsOverview
