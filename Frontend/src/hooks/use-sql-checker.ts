import { useState, useEffect, useRef, useCallback } from 'react'

import { checkSQLSyntax, type SQLSyntaxError, type SQLDialect } from '@/lib/sql-checker'
import { analyzeSQLLogic, type SQLWarning } from '@/lib/sql-logic-analyzer'

export interface SQLCheckerResult {
    isValid: boolean
    syntaxError?: SQLSyntaxError
    warnings: SQLWarning[]
    isChecking: boolean
}

interface UseSQLCheckerOptions {
    debounceMs?: number
    database?: SQLDialect
    analyzeLogic?: boolean
    checkSyntax?: boolean
}

export function useSQLChecker(
    sql: string,
    options: UseSQLCheckerOptions = {}
): SQLCheckerResult {
    const {
        debounceMs = 300,
        database = 'MySQL',
        analyzeLogic = true,
        checkSyntax = true,
    } = options

    const [result, setResult] = useState<SQLCheckerResult>({
        isValid: true,
        warnings: [],
        isChecking: false,
    })

    const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
    const lastSqlRef = useRef<string>('')

    const performCheck = useCallback((sqlToCheck: string) => {
        if (!sqlToCheck || sqlToCheck.trim() === '') {
            setResult({
                isValid: true,
                warnings: [],
                isChecking: false,
            })
            return
        }

        setResult(prev => ({ ...prev, isChecking: true }))

        let isValid = true
        let syntaxError: SQLSyntaxError | undefined

        if (checkSyntax) {
            const syntaxResult = checkSQLSyntax(sqlToCheck, database)
            isValid = syntaxResult.isValid
            syntaxError = syntaxResult.error
        }

        let warnings: SQLWarning[] = []
        if (analyzeLogic && (isValid || !checkSyntax)) {
            warnings = analyzeSQLLogic(sqlToCheck)
        }

        setResult({
            isValid,
            syntaxError,
            warnings,
            isChecking: false,
        })
    }, [database, analyzeLogic, checkSyntax])

    useEffect(() => {
        if (sql === lastSqlRef.current) {
            return
        }
        lastSqlRef.current = sql

        if (timerRef.current) {
            clearTimeout(timerRef.current)
        }

        if (sql && sql.trim() !== '') {
            setResult(prev => ({ ...prev, isChecking: true }))
        }

        timerRef.current = setTimeout(() => {
            performCheck(sql)
        }, debounceMs)

        return () => {
            if (timerRef.current) {
                clearTimeout(timerRef.current)
            }
        }
    }, [sql, debounceMs, performCheck])

    return result
}

export function checkSQLImmediate(
    sql: string,
    options: Omit<UseSQLCheckerOptions, 'debounceMs'> = {}
): Omit<SQLCheckerResult, 'isChecking'> {
    const {
        database = 'MySQL',
        analyzeLogic = true,
        checkSyntax = true,
    } = options

    if (!sql || sql.trim() === '') {
        return { isValid: true, warnings: [] }
    }

    let isValid = true
    let syntaxError: SQLSyntaxError | undefined

    if (checkSyntax) {
        const syntaxResult = checkSQLSyntax(sql, database)
        isValid = syntaxResult.isValid
        syntaxError = syntaxResult.error
    }

    let warnings: SQLWarning[] = []
    if (analyzeLogic && (isValid || !checkSyntax)) {
        warnings = analyzeSQLLogic(sql)
    }

    return { isValid, syntaxError, warnings }
}
