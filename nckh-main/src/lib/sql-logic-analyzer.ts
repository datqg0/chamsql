export type WarningType = 'warning' | 'info' | 'error' | 'optimization'

export interface SQLWarning {
    type: WarningType
    code: string
    message: string
    suggestion?: string
    line?: number
}

interface LogicRule {
    code: string
    type: WarningType
    pattern: RegExp | ((sql: string) => boolean)
    message: string
    suggestion?: string
}

const LOGIC_RULES: LogicRule[] = [
    {
        code: 'SQL001',
        type: 'warning',
        pattern: (sql: string) => {
            const s = sql.toLowerCase()
            const hasAggregate = /\b(count|sum|avg|max|min)\s*\(/i.test(s)
            const hasGroupBy = /\bgroup\s+by\b/i.test(s)
            const hasSelect = /\bselect\b/i.test(s)
            return hasSelect && hasAggregate && !hasGroupBy
        },
        message: 'Hàm aggregate (COUNT, SUM, AVG, MAX, MIN) được sử dụng mà không có GROUP BY',
        suggestion: 'Thêm GROUP BY để nhóm dữ liệu theo các cột không aggregate',
    },
    {
        code: 'SQL002',
        type: 'info',
        pattern: (sql: string) => {
            const s = sql.toLowerCase()
            return /\bjoin\b/i.test(s) &&
                !/\b(left|right|full|cross)\s+(outer\s+)?join\b/i.test(s) &&
                /\binner\s+join\b/i.test(s) === false &&
                /\bjoin\b/i.test(s)
        },
        message: 'INNER JOIN có thể làm mất các bản ghi không khớp',
        suggestion: 'Cân nhắc sử dụng LEFT JOIN nếu muốn giữ tất cả bản ghi từ bảng bên trái',
    },
    {
        code: 'SQL003',
        type: 'warning',
        pattern: (sql: string) => {
            const s = sql.toLowerCase()
            return /\blimit\b/i.test(s) && !/\border\s+by\b/i.test(s)
        },
        message: 'LIMIT được sử dụng mà không có ORDER BY',
        suggestion: 'Thêm ORDER BY để đảm bảo kết quả trả về theo thứ tự xác định',
    },
    {
        code: 'SQL004',
        type: 'optimization',
        pattern: /\bselect\s+\*/i,
        message: 'SELECT * không được khuyến khích trong production',
        suggestion: 'Chỉ định cụ thể các cột cần lấy để tối ưu performance',
    },
    {
        code: 'SQL005',
        type: 'error',
        pattern: (sql: string) => {
            const s = sql.toLowerCase().trim()
            const isDelete = /^\s*delete\s+from\b/i.test(s)
            const isUpdate = /^\s*update\b/i.test(s)
            const hasWhere = /\bwhere\b/i.test(s)
            return (isDelete || isUpdate) && !hasWhere
        },
        message: 'DELETE/UPDATE không có WHERE sẽ ảnh hưởng TẤT CẢ các bản ghi',
        suggestion: 'Thêm WHERE clause để giới hạn các bản ghi bị ảnh hưởng',
    },
    {
        code: 'SQL006',
        type: 'warning',
        pattern: /\b(=|!=|<>)\s*null\b/i,
        message: 'So sánh với NULL bằng = hoặc != sẽ không hoạt động như mong đợi',
        suggestion: 'Sử dụng IS NULL hoặc IS NOT NULL thay vì = NULL hoặc != NULL',
    },
    {
        code: 'SQL007',
        type: 'optimization',
        pattern: (sql: string) => {
            const s = sql.toLowerCase()
            return /\bwhere\b/i.test(s) && /\bor\b/i.test(s)
        },
        message: 'OR trong WHERE có thể không sử dụng được index hiệu quả',
        suggestion: 'Cân nhắc sử dụng UNION hoặc IN() để tối ưu query',
    },
    {
        code: 'SQL008',
        type: 'info',
        pattern: /\bselect\s+distinct\b/i,
        message: 'DISTINCT có thể ảnh hưởng đến performance',
        suggestion: 'Xem xét lại query để tránh duplicate thay vì dùng DISTINCT',
    },
    {
        code: 'SQL009',
        type: 'warning',
        pattern: (sql: string) => {
            const s = sql.toLowerCase()
            return /\bwhere\b.*=\s*\(\s*select\b/i.test(s)
        },
        message: 'Subquery với = có thể gây lỗi nếu trả về nhiều hơn 1 giá trị',
        suggestion: 'Sử dụng IN hoặc EXISTS thay vì = khi subquery có thể trả về nhiều giá trị',
    },
    {
        code: 'SQL010',
        type: 'info',
        pattern: (sql: string) => {
            const s = sql.toLowerCase()
            const hasMultipleTables = (s.match(/\bjoin\b/gi) || []).length >= 1 ||
                (s.match(/,/g) || []).length >= 1
            const hasAlias = /\bas\s+\w+/i.test(s) || /\bfrom\s+\w+\s+\w+/i.test(s)
            return hasMultipleTables && !hasAlias
        },
        message: 'Không sử dụng alias cho bảng khi có nhiều bảng',
        suggestion: 'Sử dụng alias (AS) cho các bảng để code dễ đọc',
    },
    {
        code: 'SQL011',
        type: 'info',
        pattern: /!=/,
        message: '!= không phải cú pháp SQL chuẩn',
        suggestion: 'Sử dụng <> để đảm bảo tương thích với tất cả database',
    },
    {
        code: 'SQL012',
        type: 'warning',
        pattern: (sql: string) => {
            const matches = sql.match(/\blike\s+['"]([^'"]+)['"]/gi)
            if (!matches) return false
            return matches.some(m => !m.includes('%') && !m.includes('_'))
        },
        message: 'LIKE được sử dụng mà không có wildcard (% hoặc _)',
        suggestion: 'Sử dụng = thay vì LIKE nếu không cần pattern matching',
    },
    {
        code: 'SQL013',
        type: 'info',
        pattern: /\border\s+by\s+\d+/i,
        message: 'ORDER BY sử dụng số thứ tự cột thay vì tên cột',
        suggestion: 'Sử dụng tên cột để code dễ bảo trì',
    },
    {
        code: 'SQL014',
        type: 'error',
        pattern: (sql: string) => {
            const s = sql.toLowerCase()
            const fromMatches = s.match(/\bfrom\s+(\w+\s*,\s*)+\w+/i)
            const hasWhere = /\bwhere\b/i.test(s)
            const hasJoin = /\bjoin\b/i.test(s)
            return !!fromMatches && !hasWhere && !hasJoin
        },
        message: 'Cartesian product - JOIN nhiều bảng mà không có điều kiện',
        suggestion: 'Thêm WHERE hoặc ON clause để chỉ định điều kiện join',
    },
]

export function analyzeSQLLogic(sql: string): SQLWarning[] {
    if (!sql || sql.trim() === '') {
        return []
    }

    const warnings: SQLWarning[] = []

    for (const rule of LOGIC_RULES) {
        let matches = false

        if (typeof rule.pattern === 'function') {
            matches = rule.pattern(sql)
        } else {
            matches = rule.pattern.test(sql)
        }

        if (matches) {
            warnings.push({
                type: rule.type,
                code: rule.code,
                message: rule.message,
                suggestion: rule.suggestion,
            })
        }
    }

    return warnings
}

export function getWarningColorClass(type: WarningType): string {
    switch (type) {
        case 'error':
            return 'text-red-500'
        case 'warning':
            return 'text-yellow-500'
        case 'info':
            return 'text-blue-500'
        case 'optimization':
            return 'text-purple-500'
        default:
            return 'text-gray-500'
    }
}
