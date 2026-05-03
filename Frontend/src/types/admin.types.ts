export interface DashboardResponse {
  overview: OverviewStats
  gradingStats: GradingStats
  dailySubmissions: DailySubmission[]
  passRates: ProblemPassRate[]
  topProblems: TopProblem[]
}

export interface OverviewStats {
  totalUsers: number
  totalProblems: number
  totalSubmissions: number
  activeUsersWeek: number
  avgSolveTimeMs: number
  usersByRole: Record<string, number>
}

export interface GradingStats {
  totalSubmissions: number
  avgGradingTimeMs: number
  minGradingTimeMs: number
  maxGradingTimeMs: number
  totalCorrect: number
  totalUsers: number
  totalProblemsAttempted: number
  passRate: number
}

export interface DailySubmission {
  date: string
  totalSubmissions: number
  correctCount: number
  avgExecutionMs: number
}

export interface ProblemPassRate {
  id: number
  title: string
  difficulty: string
  totalSubmissions: number
  correctCount: number
  passRate: number
}

export interface TopProblem {
  id: number
  title: string
  slug: string
  difficulty: string
  submissionCount: number
  uniqueUsers: number
}

export interface SystemStats {
  totalUsers: number
  totalProblems: number
  totalExams: number
  totalSubmissions: number
  usersByRole: Record<string, number>
  recentActivity: Activity[]
}

export interface Activity {
  type: string
  message: string
  timestamp: string
}
