import { Parser } from 'node-sql-parser'
import type { AST } from 'node-sql-parser'

export interface SQLSyntaxError {
    message: string
    line?: number
    column?: number
    expected?: string[]
}

export interface SQLCheckResult {
    isValid: boolean
    error?: SQLSyntaxError
    ast?: AST | AST[]
}

export type SQLDialect =
    | 'MySQL'
    | 'MariaDB'
    | 'PostgreSQL'
    | 'SQLite'
    | 'TransactSQL'
    | 'BigQuery'

const parser = new Parser()

export function checkSQLSyntax(
    sql: string,
    database: SQLDialect = 'MySQL'
): SQLCheckResult {
    if (!sql || sql.trim() === '') {
        return {
            isValid: false,
            error: {
                message: 'Truy vấn SQL trống',
            },
        }
    }

    try {
        const ast = parser.astify(sql, { database })
        return {
            isValid: true,
            ast,
        }
    } catch (error: any) {
        const errorMessage = error?.message || 'Lỗi cú pháp không xác định'

        const lineMatch = errorMessage.match(/line\s+(\d+)/i)
        const columnMatch = errorMessage.match(/column\s+(\d+)/i)
        const expectedMatch = errorMessage.match(/Expected\s+(.+)/i)

        return {
            isValid: false,
            error: {
                message: errorMessage,
                line: lineMatch ? parseInt(lineMatch[1], 10) : undefined,
                column: columnMatch ? parseInt(columnMatch[1], 10) : undefined,
                expected: expectedMatch
                    ? expectedMatch[1].split(/,\s*or\s*|,\s*/).map(s => s.trim())
                    : undefined,
            },
        }
    }
}

export function getAST(sql: string, database: SQLDialect = 'MySQL'): AST | AST[] | null {
    const result = checkSQLSyntax(sql, database)
    return result.isValid ? result.ast! : null
}

export function astToSQL(ast: AST | AST[], database: SQLDialect = 'MySQL'): string {
    try {
        return parser.sqlify(ast, { database })
    } catch {
        return ''
    }
}

export function getTableNames(sql: string, database: SQLDialect = 'MySQL'): string[] {
    try {
        const tableList = parser.tableList(sql, { database })
        return tableList.map(item => {
            const parts = item.split('::')
            return parts[parts.length - 1]
        })
    } catch {
        return []
    }
}

export function getColumnReferences(sql: string, database: SQLDialect = 'MySQL'): string[] {
    try {
        const columnList = parser.columnList(sql, { database })
        return columnList
    } catch {
        return []
    }
}
