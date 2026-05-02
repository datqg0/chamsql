// Types for Exam-related entities based on real API

export interface Topic {
    id: number
    name: string
    slug: string
    description?: string
    icon?: string
    sortOrder?: number
}

export interface TestCase {
    id: number
    name: string
    description?: string
    initScript: string
    solutionQuery: string
    weight: number
    isHidden: boolean
}

export interface TestCaseRequest {
    name: string
    description?: string
    initScript: string
    solutionQuery: string
    weight: number
    isHidden: boolean
}

export interface TestResult {
    testCaseId: number
    testCaseName: string
    status: 'accepted' | 'wrong_answer' | 'error' | 'timeout'
    executionMs: number
    isCorrect: boolean
    isHidden: boolean
    actualOutput?: unknown
    errorMessage?: string
}

export interface Problem {
    id: number
    title: string
    slug: string
    description: string
    difficulty: 'easy' | 'medium' | 'hard'
    initScript?: string
    solutionQuery?: string
    supportedDatabases: string[]
    orderMatters: boolean
    isPublic: boolean
    testCases?: TestCase[]
    topicId?: number
    topicName?: string
    topicSlug?: string
    createdAt?: string
    updatedAt?: string
}

export interface ProblemListResponse {
    data: Problem[]
    page: number
    pageSize: number
    total: number
}

export interface Submission {
    id: number
    problemId: number
    userId: number
    examId?: number
    examTitle?: string
    problemTitle?: string
    problemSlug?: string
    code: string
    databaseType: string
    status: 'pending' | 'accepted' | 'wrong_answer' | 'error' | 'timeout'
    isCorrect: boolean
    executionTime?: number
    score?: number
    totalTests?: number
    passedTests?: number
    errorMessage?: string
    expectedOutput?: unknown
    actualOutput?: unknown
    testResults?: TestResult[]
    submittedAt: string
    createdAt?: string
}

export interface SubmissionListResponse {
    data: Submission[]
    page: number
    pageSize: number
    total: number
}

export interface RunQueryRequest {
    code: string
    databaseType: string
}

export interface RunQueryResponse {
    success: boolean
    columns?: string[]
    rows?: unknown[]
    rowCount: number
    executionMs: number
    error?: string
    errorType?: string
}

export interface SubmitSolutionRequest {
    code: string
    databaseType: string
}

export interface SubmitSolutionResponse {
    id: number
    isCorrect: boolean
    status: 'accepted' | 'wrong_answer' | 'error' | 'timeout'
    executionMs: number
    score: number
    totalTests: number
    passedTests: number
    message?: string
    testResults?: TestResult[]
}

// Exam Types
export interface Exam {
    id: number
    title: string
    description?: string
    startTime: string
    endTime: string
    durationMinutes: number
    allowedDatabases: string[] | null
    allowAiAssistance: boolean
    shuffleProblems: boolean
    showResultImmediately: boolean
    maxAttempts: number
    isPublic: boolean
    createdBy?: number
    createdAt?: string
    updatedAt?: string
    problems?: ExamProblem[]
    status?: string
    problemCount?: number
}

export interface ExamProblem {
    id: number
    examId: number
    problemId: number
    points: number
    sortOrder: number
    problem?: Problem
}

export interface CreateExamRequest {
    title: string
    description?: string
    startTime: string
    endTime: string
    durationMinutes: number
    allowedDatabases: string[]
    allowAiAssistance: boolean
    shuffleProblems: boolean
    showResultImmediately: boolean
    maxAttempts: number
    isPublic: boolean
}

export interface AddExamProblemRequest {
    problemId: number
    points: number
    sortOrder: number
}

export interface AddParticipantsRequest {
    userIds: number[]
}

export interface ExamParticipant {
    id: number
    userId: number
    fullName: string
    email: string
    studentId?: string
    status: string
    startedAt?: string
    submittedAt?: string
    totalScore: number
}

export interface ExamSubmission {
    id: number
    examId: number
    problemId: number
    userId: number
    code: string
    databaseType: string
    status: 'pending' | 'accepted' | 'wrong_answer' | 'error'
    points?: number
    submittedAt: string
}

export interface SubmitExamAnswerRequest {
    problemId: number
    code: string
    databaseType: string
}

export interface MyExam {
    id: number
    exam: Exam
    status: 'not_started' | 'in_progress' | 'finished'
    startedAt?: string
    finishedAt?: string
    score?: number
    totalPoints?: number
    attemptNumber?: number
}

// Admin Types
export interface AdminStats {
    totalUsers: number
    totalProblems: number
    totalSubmissions: number
    totalExams: number
}

export interface ImportUserDto {
    email: string
    username: string
    fullName: string
    studentId?: string
    role: 'student' | 'lecturer'
}

export interface ImportUsersRequest {
    users: ImportUserDto[]
}
