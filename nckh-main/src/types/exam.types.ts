// Types for Exam-related entities based on real API

export interface Topic {
    id: number
    name: string
    slug: string
    description?: string
    icon?: string
    sortOrder?: number
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
    testCases?: string
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
    code: string
    databaseType: string
    status: 'pending' | 'accepted' | 'wrong_answer' | 'error'
    executionTime?: number
    createdAt: string
    examId?: number
    examTitle?: string
    score?: number
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
    data?: any[]
    error?: string
    executionTime?: string
}

export interface SubmitSolutionRequest {
    code: string
    databaseType: string
}

export interface SubmitSolutionResponse {
    success: boolean
    status: 'accepted' | 'wrong_answer' | 'error'
    message?: string
    testCases?: {
        passed: number
        total: number
    }
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
