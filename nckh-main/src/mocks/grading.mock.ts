// Mock data cho trang chấm điểm
export const USE_GRADING_MOCK = true

export interface Submission {
    id: number
    studentId: number
    studentName: string
    studentCode: string
    exerciseId: number
    exerciseName: string
    submittedAt: string
    sqlQuery: string
    status: 'pending' | 'graded' | 'error'
    score?: number
    maxScore: number
    executionTime?: string
    feedback?: string
    expectedOutput?: any[]
    actualOutput?: any[]
    isCorrect?: boolean
}

export interface Exercise {
    id: number
    title: string
    description: string
    difficulty: 'easy' | 'medium' | 'hard'
    category: string
    expectedQuery: string
    testCases: number
    maxScore: number
}

// Mock danh sách bài tập
export const MOCK_EXERCISES: Exercise[] = [
    {
        id: 1,
        title: 'Truy vấn cơ bản - SELECT',
        description: 'Lấy tất cả thông tin từ bảng users',
        difficulty: 'easy',
        category: 'SELECT',
        expectedQuery: 'SELECT * FROM users',
        testCases: 3,
        maxScore: 10,
    },
    {
        id: 2,
        title: 'Lọc dữ liệu với WHERE',
        description: 'Tìm các user có tuổi lớn hơn 25',
        difficulty: 'easy',
        category: 'WHERE',
        expectedQuery: 'SELECT * FROM users WHERE age > 25',
        testCases: 5,
        maxScore: 15,
    },
    {
        id: 3,
        title: 'JOIN hai bảng',
        description: 'Kết hợp bảng orders và customers',
        difficulty: 'medium',
        category: 'JOIN',
        expectedQuery: 'SELECT o.*, c.name FROM orders o INNER JOIN customers c ON o.customer_id = c.id',
        testCases: 4,
        maxScore: 20,
    },
    {
        id: 4,
        title: 'Aggregate Functions',
        description: 'Tính tổng doanh thu theo tháng',
        difficulty: 'medium',
        category: 'AGGREGATE',
        expectedQuery: 'SELECT MONTH(order_date), SUM(total) FROM orders GROUP BY MONTH(order_date)',
        testCases: 3,
        maxScore: 25,
    },
    {
        id: 5,
        title: 'Subquery nâng cao',
        description: 'Tìm sản phẩm có giá cao hơn giá trung bình',
        difficulty: 'hard',
        category: 'SUBQUERY',
        expectedQuery: 'SELECT * FROM products WHERE price > (SELECT AVG(price) FROM products)',
        testCases: 4,
        maxScore: 30,
    },
]

// Mock danh sách bài nộp
export const MOCK_SUBMISSIONS: Submission[] = [
    {
        id: 1,
        studentId: 101,
        studentName: 'Nguyễn Văn An',
        studentCode: 'SV001',
        exerciseId: 1,
        exerciseName: 'Truy vấn cơ bản - SELECT',
        submittedAt: '2026-01-17T10:30:00',
        sqlQuery: 'SELECT * FROM users;',
        status: 'graded',
        score: 10,
        maxScore: 10,
        executionTime: '0.012s',
        isCorrect: true,
        feedback: 'Chính xác! Câu truy vấn đúng hoàn toàn.',
        expectedOutput: [
            { id: 1, name: 'User 1', email: 'user1@test.com' },
            { id: 2, name: 'User 2', email: 'user2@test.com' },
        ],
        actualOutput: [
            { id: 1, name: 'User 1', email: 'user1@test.com' },
            { id: 2, name: 'User 2', email: 'user2@test.com' },
        ],
    },
    {
        id: 2,
        studentId: 102,
        studentName: 'Trần Thị Bình',
        studentCode: 'SV002',
        exerciseId: 2,
        exerciseName: 'Lọc dữ liệu với WHERE',
        submittedAt: '2026-01-17T11:15:00',
        sqlQuery: 'SELECT * FROM users WHERE age >= 25',
        status: 'graded',
        score: 12,
        maxScore: 15,
        executionTime: '0.018s',
        isCorrect: false,
        feedback: 'Gần đúng, nhưng điều kiện sai. Yêu cầu là > 25, không phải >= 25',
        expectedOutput: [
            { id: 2, name: 'User 2', age: 30 },
            { id: 3, name: 'User 3', age: 28 },
        ],
        actualOutput: [
            { id: 1, name: 'User 1', age: 25 },
            { id: 2, name: 'User 2', age: 30 },
            { id: 3, name: 'User 3', age: 28 },
        ],
    },
    {
        id: 3,
        studentId: 103,
        studentName: 'Lê Hoàng Cường',
        studentCode: 'SV003',
        exerciseId: 3,
        exerciseName: 'JOIN hai bảng',
        submittedAt: '2026-01-17T14:20:00',
        sqlQuery: 'SELECT * FROM orders, customers WHERE orders.customer_id = customers.id',
        status: 'pending',
        maxScore: 20,
        executionTime: '0.045s',
    },
    {
        id: 4,
        studentId: 104,
        studentName: 'Phạm Minh Đức',
        studentCode: 'SV004',
        exerciseId: 3,
        exerciseName: 'JOIN hai bảng',
        submittedAt: '2026-01-17T14:45:00',
        sqlQuery: 'SELECT o.*, c.name FROM orders o JOIN customers c ON o.customer_id = c.id',
        status: 'pending',
        maxScore: 20,
        executionTime: '0.032s',
    },
    {
        id: 5,
        studentId: 105,
        studentName: 'Hoàng Thị Em',
        studentCode: 'SV005',
        exerciseId: 4,
        exerciseName: 'Aggregate Functions',
        submittedAt: '2026-01-17T15:00:00',
        sqlQuery: 'SELECT MONTH(order_date) as month, SUM(total) as revenue FROM orders GROUP BY month',
        status: 'pending',
        maxScore: 25,
    },
    {
        id: 6,
        studentId: 101,
        studentName: 'Nguyễn Văn An',
        studentCode: 'SV001',
        exerciseId: 5,
        exerciseName: 'Subquery nâng cao',
        submittedAt: '2026-01-17T16:30:00',
        sqlQuery: 'SELECT * FROM products WHERE price > AVG(price)',
        status: 'error',
        maxScore: 30,
        feedback: 'Lỗi cú pháp: Không thể sử dụng AVG() trực tiếp trong WHERE. Cần dùng subquery.',
    },
    {
        id: 9,
        studentId: 108,
        studentName: 'Lê Văn Tự',
        studentCode: 'SV008',
        exerciseId: 1,
        exerciseName: 'Truy vấn cơ bản - SELECT',
        submittedAt: '2026-01-29T10:00:00',
        sqlQuery: 'select * from users',
        status: 'graded',
        score: 10,
        maxScore: 10,
        executionTime: '0.011s',
        isCorrect: true,
        feedback: 'Hoàn hảo! Kết quả đầu ra khớp với yêu cầu.',
        expectedOutput: [
            { id: 1, name: 'User 1', email: 'user1@test.com' },
            { id: 2, name: 'User 2', email: 'user2@test.com' },
        ],
        actualOutput: [
            { id: 1, name: 'User 1', email: 'user1@test.com' },
            { id: 2, name: 'User 2', email: 'user2@test.com' },
        ],
    },
    {
        id: 7,
        studentId: 106,
        studentName: 'Vũ Quốc Phong',
        studentCode: 'SV006',
        exerciseId: 1,
        exerciseName: 'Truy vấn cơ bản - SELECT',
        submittedAt: '2026-01-17T09:00:00',
        sqlQuery: 'SELECT * FROM users',
        status: 'graded',
        score: 10,
        maxScore: 10,
        executionTime: '0.010s',
        isCorrect: true,
        feedback: 'Hoàn hảo!',
    },
    {
        id: 8,
        studentId: 107,
        studentName: 'Đặng Thu Giang',
        studentCode: 'SV007',
        exerciseId: 2,
        exerciseName: 'Lọc dữ liệu với WHERE',
        submittedAt: '2026-01-17T10:00:00',
        sqlQuery: 'SELECT * FROM users WHERE age > 25',
        status: 'graded',
        score: 15,
        maxScore: 15,
        executionTime: '0.015s',
        isCorrect: true,
        feedback: 'Xuất sắc! Câu truy vấn chính xác.',
    },
]

// Store để quản lý state
let submissionsStore = [...MOCK_SUBMISSIONS]

export function getMockSubmissions(filters?: {
    status?: 'pending' | 'graded' | 'error'
    exerciseId?: number
    studentCode?: string
}): Promise<Submission[]> {
    return new Promise((resolve) => {
        setTimeout(() => {
            let result = [...submissionsStore]

            if (filters?.status) {
                result = result.filter((s) => s.status === filters.status)
            }
            if (filters?.exerciseId) {
                result = result.filter((s) => s.exerciseId === filters.exerciseId)
            }
            if (filters?.studentCode) {
                result = result.filter((s) =>
                    s.studentCode.toLowerCase().includes(filters.studentCode!.toLowerCase()) ||
                    s.studentName.toLowerCase().includes(filters.studentCode!.toLowerCase())
                )
            }

            // Sort by submittedAt descending
            result.sort((a, b) => new Date(b.submittedAt).getTime() - new Date(a.submittedAt).getTime())

            resolve(result)
        }, 300)
    })
}

export function getMockSubmissionById(id: number): Promise<Submission | null> {
    return new Promise((resolve) => {
        setTimeout(() => {
            const submission = submissionsStore.find((s) => s.id === id)
            resolve(submission || null)
        }, 200)
    })
}

export function mockGradeSubmission(
    submissionId: number,
    score: number,
    feedback: string
): Promise<{ success: boolean; message: string }> {
    return new Promise((resolve) => {
        setTimeout(() => {
            const index = submissionsStore.findIndex((s) => s.id === submissionId)
            if (index !== -1) {
                submissionsStore[index] = {
                    ...submissionsStore[index],
                    status: 'graded',
                    score,
                    feedback,
                    isCorrect: score === submissionsStore[index].maxScore,
                }
                resolve({ success: true, message: 'Chấm điểm thành công!' })
            } else {
                resolve({ success: false, message: 'Không tìm thấy bài nộp!' })
            }
        }, 500)
    })
}

export function mockAutoGrade(submissionId: number): Promise<{
    success: boolean
    score: number
    feedback: string
    isCorrect: boolean
}> {
    return new Promise((resolve) => {
        setTimeout(() => {
            const submission = submissionsStore.find((s) => s.id === submissionId)
            if (submission) {
                // Simulate auto-grading logic based on OUTPUT COMPARISON
                // NOTE: This is a MOCK simulation. In real application, the backend runs the query.

                // 1. Get the Exercise to check expected output (In real app, backend retrieves this)
                const exercise = MOCK_EXERCISES.find(e => e.id === submission.exerciseId);

                // 2. Simulate "Running" the student query
                // For mock purposes, we'll assume:
                // - If query contains correct keywords/structure -> returns CORRECT output
                // - If query is slightly off -> returns PARTIAL/WRONG output

                // Mock execution result
                let actualOutput: any[] = [];
                let expectedOutput: any[] = [];
                let executionTime = (Math.random() * 0.1).toFixed(3) + 's';

                // Hardcoded expectation logic for demonstration based on Exercise ID
                if (submission.exerciseId === 1) { // SELECT * FROM users
                    expectedOutput = [
                        { id: 1, name: 'User 1', email: 'user1@test.com' },
                        { id: 2, name: 'User 2', email: 'user2@test.com' },
                    ];

                    if (submission.sqlQuery.toLowerCase().includes('select * from users')) {
                        actualOutput = [...expectedOutput];
                    } else if (submission.sqlQuery.toLowerCase().includes('select id, name from users')) {
                        actualOutput = [
                            { id: 1, name: 'User 1' },
                            { id: 2, name: 'User 2' },
                        ];
                    } else {
                        actualOutput = [];
                    }
                } else {
                    // Default fallback for other exercises
                    expectedOutput = [{ message: 'Expected Result' }];
                    // Randomly decide if student output matches
                    if (Math.random() > 0.3) {
                        actualOutput = [...expectedOutput];
                    } else {
                        actualOutput = [{ message: 'Wrong Result' }];
                    }
                }

                // 3. Compare Outputs
                const isCorrect = JSON.stringify(actualOutput) === JSON.stringify(expectedOutput);

                const score = isCorrect ? submission.maxScore : 0; // Binary scoring for simplicity

                const feedback = isCorrect
                    ? 'Xuất sắc! Kết quả đầu ra của bạn hoàn toàn khớp với đáp án mong đợi.'
                    : `Kết quả không khớp. \nMong đợi: ${JSON.stringify(expectedOutput)} \nThực tế: ${JSON.stringify(actualOutput)}`;

                // Update store
                const index = submissionsStore.findIndex((s) => s.id === submissionId)
                if (index !== -1) {
                    submissionsStore[index] = {
                        ...submissionsStore[index],
                        status: 'graded',
                        score,
                        feedback,
                        isCorrect,
                        actualOutput,
                        expectedOutput,
                        executionTime,
                    }
                }

                resolve({ success: true, score, feedback, isCorrect })
            } else {
                resolve({
                    success: false,
                    score: 0,
                    feedback: 'Không tìm thấy bài nộp!',
                    isCorrect: false,
                })
            }
        }, 1000) // Simulate processing time
    })
}

// Statistics
export function getMockGradingStats(): Promise<{
    totalSubmissions: number
    pendingCount: number
    gradedCount: number
    errorCount: number
    averageScore: number
}> {
    return new Promise((resolve) => {
        setTimeout(() => {
            const graded = submissionsStore.filter((s) => s.status === 'graded')
            const avgScore =
                graded.length > 0
                    ? graded.reduce((sum, s) => sum + (s.score || 0), 0) / graded.length
                    : 0

            resolve({
                totalSubmissions: submissionsStore.length,
                pendingCount: submissionsStore.filter((s) => s.status === 'pending').length,
                gradedCount: graded.length,
                errorCount: submissionsStore.filter((s) => s.status === 'error').length,
                averageScore: Math.round(avgScore * 10) / 10,
            })
        }, 200)
    })
}

// Reset mock data
export function resetMockSubmissions() {
    submissionsStore = [...MOCK_SUBMISSIONS]
}
