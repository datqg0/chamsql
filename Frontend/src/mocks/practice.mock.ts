// Mock data cho trang Luyện tập sinh viên
export const USE_PRACTICE_MOCK = true

export type Difficulty = 'easy' | 'medium' | 'hard'
export type PracticeStatus = 'not_started' | 'in_progress' | 'completed'

export interface PracticeExercise {
    id: number
    title: string
    description: string
    difficulty: Difficulty
    category: string
    tags: string[]
    completedCount: number
    totalAttempts: number
    acceptanceRate: number
    points: number
    estimatedTime: string
    hint?: string
    sampleInput?: string
    expectedOutput?: string
    userStatus: PracticeStatus
    lastAttempt?: string
}

export interface ExerciseDetail extends PracticeExercise {
    problemStatement: string
    schemaDescription: string
    examples: {
        input: string
        output: string
        explanation?: string
    }[]
    constraints: string[]
    starterCode: string
}

// Mock danh sách bài luyện tập
export const MOCK_PRACTICE_EXERCISES: PracticeExercise[] = [
    {
        id: 1,
        title: 'Lấy tất cả người dùng',
        description: 'Viết câu truy vấn để lấy tất cả thông tin từ bảng users',
        difficulty: 'easy',
        category: 'SELECT Cơ bản',
        tags: ['SELECT', 'FROM'],
        completedCount: 1234,
        totalAttempts: 1500,
        acceptanceRate: 82.3,
        points: 10,
        estimatedTime: '5 phút',
        userStatus: 'completed',
        lastAttempt: '2026-01-15',
    },
    {
        id: 2,
        title: 'Lọc theo điều kiện',
        description: 'Tìm các user có tuổi lớn hơn 25 và đang active',
        difficulty: 'easy',
        category: 'WHERE',
        tags: ['SELECT', 'WHERE', 'AND'],
        completedCount: 890,
        totalAttempts: 1200,
        acceptanceRate: 74.2,
        points: 15,
        estimatedTime: '8 phút',
        userStatus: 'in_progress',
        lastAttempt: '2026-01-17',
    },
    {
        id: 3,
        title: 'JOIN hai bảng',
        description: 'Kết hợp thông tin từ bảng orders và customers',
        difficulty: 'medium',
        category: 'JOIN',
        tags: ['JOIN', 'INNER JOIN', 'ON'],
        completedCount: 567,
        totalAttempts: 980,
        acceptanceRate: 57.9,
        points: 20,
        estimatedTime: '12 phút',
        userStatus: 'not_started',
    },
    {
        id: 4,
        title: 'LEFT JOIN với NULL',
        description: 'Tìm customers chưa có đơn hàng nào',
        difficulty: 'medium',
        category: 'JOIN',
        tags: ['LEFT JOIN', 'IS NULL'],
        completedCount: 432,
        totalAttempts: 850,
        acceptanceRate: 50.8,
        points: 25,
        estimatedTime: '15 phút',
        userStatus: 'not_started',
    },
    {
        id: 5,
        title: 'Aggregate - Tính tổng',
        description: 'Tính tổng doanh thu theo từng tháng',
        difficulty: 'medium',
        category: 'AGGREGATE',
        tags: ['SUM', 'GROUP BY', 'MONTH'],
        completedCount: 345,
        totalAttempts: 720,
        acceptanceRate: 47.9,
        points: 25,
        estimatedTime: '15 phút',
        userStatus: 'not_started',
    },
    {
        id: 6,
        title: 'Subquery trong WHERE',
        description: 'Tìm sản phẩm có giá cao hơn giá trung bình',
        difficulty: 'hard',
        category: 'SUBQUERY',
        tags: ['SUBQUERY', 'AVG', 'WHERE'],
        completedCount: 234,
        totalAttempts: 650,
        acceptanceRate: 36.0,
        points: 30,
        estimatedTime: '20 phút',
        userStatus: 'not_started',
    },
    {
        id: 7,
        title: 'HAVING với Aggregate',
        description: 'Tìm các nhóm có tổng giá trị đơn hàng > 1000',
        difficulty: 'medium',
        category: 'AGGREGATE',
        tags: ['GROUP BY', 'HAVING', 'SUM'],
        completedCount: 289,
        totalAttempts: 580,
        acceptanceRate: 49.8,
        points: 25,
        estimatedTime: '15 phút',
        userStatus: 'not_started',
    },
    {
        id: 8,
        title: 'Window Functions cơ bản',
        description: 'Sử dụng ROW_NUMBER để đánh số thứ tự',
        difficulty: 'hard',
        category: 'WINDOW',
        tags: ['ROW_NUMBER', 'OVER', 'PARTITION BY'],
        completedCount: 156,
        totalAttempts: 480,
        acceptanceRate: 32.5,
        points: 35,
        estimatedTime: '25 phút',
        userStatus: 'not_started',
    },
    {
        id: 9,
        title: 'CASE WHEN',
        description: 'Phân loại khách hàng theo mức chi tiêu',
        difficulty: 'medium',
        category: 'CONDITIONAL',
        tags: ['CASE', 'WHEN', 'THEN'],
        completedCount: 412,
        totalAttempts: 730,
        acceptanceRate: 56.4,
        points: 20,
        estimatedTime: '12 phút',
        userStatus: 'not_started',
    },
    {
        id: 10,
        title: 'CTE - Common Table Expression',
        description: 'Sử dụng WITH để tạo bảng tạm',
        difficulty: 'hard',
        category: 'CTE',
        tags: ['WITH', 'CTE', 'Recursive'],
        completedCount: 98,
        totalAttempts: 320,
        acceptanceRate: 30.6,
        points: 40,
        estimatedTime: '30 phút',
        userStatus: 'not_started',
    },
]

// Mock chi tiết bài tập
export const MOCK_EXERCISE_DETAILS: Record<number, ExerciseDetail> = {
    1: {
        ...MOCK_PRACTICE_EXERCISES[0],
        problemStatement: `
# Lấy tất cả người dùng

Viết một câu truy vấn SQL để lấy **tất cả thông tin** từ bảng \`users\`.

## Mô tả bảng

Bảng \`users\` có cấu trúc như sau:

| Cột | Kiểu dữ liệu | Mô tả |
|-----|--------------|-------|
| id | INT | Mã người dùng (Primary Key) |
| name | VARCHAR(100) | Tên người dùng |
| email | VARCHAR(150) | Email người dùng |
| age | INT | Tuổi |
| created_at | DATETIME | Ngày tạo tài khoản |
        `.trim(),
        schemaDescription: 'Bảng users chứa thông tin cơ bản của người dùng trong hệ thống.',
        examples: [
            {
                input: 'Bảng users có 2 bản ghi',
                output: '| id | name | email | age | created_at |\n|---|---|---|---|---|\n| 1 | Nguyễn Văn A | a@test.com | 25 | 2024-01-01 |',
                explanation: 'Trả về tất cả các cột và bản ghi từ bảng users',
            },
        ],
        constraints: [
            'Kết quả phải bao gồm tất cả các cột',
            'Không cần sắp xếp kết quả',
        ],
        starterCode: '',
    },
    2: {
        ...MOCK_PRACTICE_EXERCISES[1],
        problemStatement: `
# Lọc người dùng theo điều kiện

Viết câu truy vấn SQL để tìm tất cả người dùng có:
- Tuổi **lớn hơn 25**
- Trạng thái **active = 1**

## Mô tả bảng

Bảng \`users\`:

| Cột | Kiểu dữ liệu | Mô tả |
|-----|--------------|-------|
| id | INT | Mã người dùng |
| name | VARCHAR(100) | Tên |
| age | INT | Tuổi |
| active | TINYINT | Trạng thái (1: active, 0: inactive) |
        `.trim(),
        schemaDescription: 'Bảng users với trường active để đánh dấu trạng thái hoạt động.',
        examples: [
            {
                input: 'users: [(1, "A", 30, 1), (2, "B", 22, 1), (3, "C", 28, 0)]',
                output: '| id | name | age | active |\n|---|---|---|---|\n| 1 | A | 30 | 1 |',
                explanation: 'Chỉ user A thỏa mãn cả 2 điều kiện: age > 25 VÀ active = 1',
            },
        ],
        constraints: [
            'Phải sử dụng mệnh đề WHERE',
            'Kết hợp 2 điều kiện bằng AND',
        ],
        starterCode: '',
    },
    3: {
        ...MOCK_PRACTICE_EXERCISES[2],
        problemStatement: `
# JOIN hai bảng

Viết câu truy vấn SQL để kết hợp thông tin từ bảng \`orders\` và \`customers\`.

Kết quả cần bao gồm:
- Tất cả thông tin đơn hàng
- Tên khách hàng

## Mô tả bảng

**Bảng \`orders\`:**
| Cột | Kiểu | Mô tả |
|-----|------|-------|
| id | INT | Mã đơn hàng |
| customer_id | INT | FK tới customers |
| total | DECIMAL | Tổng tiền |
| order_date | DATE | Ngày đặt |

**Bảng \`customers\`:**
| Cột | Kiểu | Mô tả |
|-----|------|-------|
| id | INT | Mã khách hàng |
| name | VARCHAR | Tên khách hàng |
        `.trim(),
        schemaDescription: 'Hai bảng có mối quan hệ 1-n thông qua customer_id.',
        examples: [
            {
                input: 'orders: [(1, 101, 500)], customers: [(101, "Nguyễn A")]',
                output: '| id | customer_id | total | name |\n|---|---|---|---|\n| 1 | 101 | 500 | Nguyễn A |',
            },
        ],
        constraints: [
            'Sử dụng INNER JOIN hoặc JOIN',
            'Kết nối qua customer_id',
        ],
        starterCode: '',
    },
}

// API mock functions
export function getMockPracticeExercises(filters?: {
    difficulty?: Difficulty
    category?: string
    status?: PracticeStatus
    search?: string
}): Promise<PracticeExercise[]> {
    return new Promise((resolve) => {
        setTimeout(() => {
            let result = [...MOCK_PRACTICE_EXERCISES]

            if (filters?.difficulty) {
                result = result.filter((e) => e.difficulty === filters.difficulty)
            }
            if (filters?.category) {
                result = result.filter((e) => e.category === filters.category)
            }
            if (filters?.status) {
                result = result.filter((e) => e.userStatus === filters.status)
            }
            if (filters?.search) {
                const search = filters.search.toLowerCase()
                result = result.filter(
                    (e) =>
                        e.title.toLowerCase().includes(search) ||
                        e.description.toLowerCase().includes(search) ||
                        e.tags.some((t) => t.toLowerCase().includes(search))
                )
            }

            resolve(result)
        }, 300)
    })
}

export function getMockExerciseDetail(id: number): Promise<ExerciseDetail | null> {
    return new Promise((resolve) => {
        setTimeout(() => {
            resolve(MOCK_EXERCISE_DETAILS[id] || null)
        }, 200)
    })
}

export function getMockPracticeStats(): Promise<{
    totalExercises: number
    completed: number
    inProgress: number
    totalPoints: number
    earnedPoints: number
}> {
    return new Promise((resolve) => {
        setTimeout(() => {
            const completed = MOCK_PRACTICE_EXERCISES.filter((e) => e.userStatus === 'completed')
            const inProgress = MOCK_PRACTICE_EXERCISES.filter((e) => e.userStatus === 'in_progress')

            resolve({
                totalExercises: MOCK_PRACTICE_EXERCISES.length,
                completed: completed.length,
                inProgress: inProgress.length,
                totalPoints: MOCK_PRACTICE_EXERCISES.reduce((sum, e) => sum + e.points, 0),
                earnedPoints: completed.reduce((sum, e) => sum + e.points, 0),
            })
        }, 200)
    })
}

// Categories for filter
export const PRACTICE_CATEGORIES = [
    'SELECT Cơ bản',
    'WHERE',
    'JOIN',
    'AGGREGATE',
    'SUBQUERY',
    'WINDOW',
    'CONDITIONAL',
    'CTE',
]
