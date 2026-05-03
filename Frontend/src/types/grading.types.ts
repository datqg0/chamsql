export interface GradingStats {
  examId: number
  totalSubmissions: number
  gradedCount: number
  ungradedCount: number
  gradingPercentage: number
  averageScore: number
  maxScore: number
  minScore: number
}

export interface GradingSubmission {
  id: number
  submissionId: number
  studentId: number
  studentName: string
  studentCode?: string
  problemId: number
  problemTitle: string
  score: number
  maxScore: number
  status: 'pending' | 'graded' | 'error'
  executionTimeMs?: number
  isCorrect: boolean
  scoringMode: string
  gradedBy: number | null
  gradedByName: string | null
  gradedAt: string | null
  feedback: string
  comparisonLog: string
  submittedAt: string
  submittedCode: string
  errorMessage?: string
  studentAnswer: string | null
  referenceAnswer: string | null
}
