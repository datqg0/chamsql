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
  submissionId: number
  studentId: number
  studentName: string
  problemTitle: string
  score: number
  maxPoints: number
  isCorrect: boolean
  scoringMode: string
  gradedBy: number | null
  gradedByName: string | null
  gradedAt: string | null
  feedback: string
  comparisonLog: string
  submittedAt: string
  studentAnswer: string | null
  referenceAnswer: string | null
}
