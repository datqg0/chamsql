import Editor, { type OnMount } from '@monaco-editor/react'
import { useTheme } from '@/components/theme-provider'
import type { editor } from 'monaco-editor'
import { useEffect, useRef } from 'react'
import type { SQLSyntaxError } from '@/lib/sql-checker'

// SQL Keywords for autocomplete
const SQL_KEYWORDS = [
    'SELECT', 'FROM', 'WHERE', 'AND', 'OR', 'NOT', 'IN', 'IS', 'NULL',
    'INSERT', 'INTO', 'VALUES', 'UPDATE', 'SET', 'DELETE',
    'CREATE', 'TABLE', 'DATABASE', 'INDEX', 'VIEW', 'DROP', 'ALTER',
    'ADD', 'COLUMN', 'PRIMARY', 'KEY', 'FOREIGN', 'REFERENCES',
    'JOIN', 'INNER', 'LEFT', 'RIGHT', 'OUTER', 'FULL', 'CROSS', 'ON',
    'GROUP', 'BY', 'HAVING', 'ORDER', 'ASC', 'DESC', 'LIMIT', 'OFFSET',
    'DISTINCT', 'AS', 'CASE', 'WHEN', 'THEN', 'ELSE', 'END',
    'UNION', 'ALL', 'INTERSECT', 'EXCEPT', 'EXISTS',
    'LIKE', 'BETWEEN', 'TRUE', 'FALSE',
    'COUNT', 'SUM', 'AVG', 'MIN', 'MAX',
    'CONSTRAINT', 'UNIQUE', 'CHECK', 'DEFAULT',
    'TRUNCATE', 'CASCADE', 'RESTRICT',
    'COMMIT', 'ROLLBACK', 'TRANSACTION', 'BEGIN',
    'GRANT', 'REVOKE', 'PRIVILEGES',
]

// SQL Functions for autocomplete
const SQL_FUNCTIONS = [
    // Aggregate functions
    { name: 'COUNT', snippet: 'COUNT(${1:column})', description: 'Returns the number of rows' },
    { name: 'SUM', snippet: 'SUM(${1:column})', description: 'Returns the sum of values' },
    { name: 'AVG', snippet: 'AVG(${1:column})', description: 'Returns the average value' },
    { name: 'MIN', snippet: 'MIN(${1:column})', description: 'Returns the minimum value' },
    { name: 'MAX', snippet: 'MAX(${1:column})', description: 'Returns the maximum value' },
    // String functions
    { name: 'CONCAT', snippet: 'CONCAT(${1:str1}, ${2:str2})', description: 'Concatenates strings' },
    { name: 'SUBSTRING', snippet: 'SUBSTRING(${1:str}, ${2:start}, ${3:length})', description: 'Extracts a substring' },
    { name: 'UPPER', snippet: 'UPPER(${1:str})', description: 'Converts to uppercase' },
    { name: 'LOWER', snippet: 'LOWER(${1:str})', description: 'Converts to lowercase' },
    { name: 'TRIM', snippet: 'TRIM(${1:str})', description: 'Removes leading and trailing spaces' },
    { name: 'LENGTH', snippet: 'LENGTH(${1:str})', description: 'Returns the length of a string' },
    { name: 'REPLACE', snippet: 'REPLACE(${1:str}, ${2:from}, ${3:to})', description: 'Replaces occurrences' },
    // Date functions
    { name: 'NOW', snippet: 'NOW()', description: 'Returns current date and time' },
    { name: 'CURDATE', snippet: 'CURDATE()', description: 'Returns current date' },
    { name: 'CURTIME', snippet: 'CURTIME()', description: 'Returns current time' },
    { name: 'DATE', snippet: 'DATE(${1:datetime})', description: 'Extracts date part' },
    { name: 'YEAR', snippet: 'YEAR(${1:date})', description: 'Returns the year' },
    { name: 'MONTH', snippet: 'MONTH(${1:date})', description: 'Returns the month' },
    { name: 'DAY', snippet: 'DAY(${1:date})', description: 'Returns the day' },
    { name: 'DATEDIFF', snippet: 'DATEDIFF(${1:date1}, ${2:date2})', description: 'Returns difference in days' },
    // Other functions
    { name: 'COALESCE', snippet: 'COALESCE(${1:value1}, ${2:value2})', description: 'Returns first non-null value' },
    { name: 'NULLIF', snippet: 'NULLIF(${1:expr1}, ${2:expr2})', description: 'Returns NULL if expressions are equal' },
    { name: 'CAST', snippet: 'CAST(${1:value} AS ${2:type})', description: 'Converts value to type' },
    { name: 'CONVERT', snippet: 'CONVERT(${1:value}, ${2:type})', description: 'Converts value to type' },
    { name: 'IFNULL', snippet: 'IFNULL(${1:expr}, ${2:alt_value})', description: 'Returns alt_value if expr is NULL' },
    { name: 'IF', snippet: 'IF(${1:condition}, ${2:true_val}, ${3:false_val})', description: 'Conditional expression' },
]

// SQL Snippets for autocomplete
const SQL_SNIPPETS = [
    { name: 'SELECT * FROM', snippet: 'SELECT * FROM ${1:table_name}', description: 'Select all columns from table' },
    { name: 'SELECT columns FROM', snippet: 'SELECT ${1:columns} FROM ${2:table_name}', description: 'Select specific columns' },
    { name: 'SELECT WHERE', snippet: 'SELECT ${1:columns} FROM ${2:table_name} WHERE ${3:condition}', description: 'Select with condition' },
    { name: 'INSERT INTO', snippet: 'INSERT INTO ${1:table_name} (${2:columns}) VALUES (${3:values})', description: 'Insert new row' },
    { name: 'UPDATE SET', snippet: 'UPDATE ${1:table_name} SET ${2:column} = ${3:value} WHERE ${4:condition}', description: 'Update rows' },
    { name: 'DELETE FROM', snippet: 'DELETE FROM ${1:table_name} WHERE ${2:condition}', description: 'Delete rows' },
    { name: 'CREATE TABLE', snippet: 'CREATE TABLE ${1:table_name} (\n\t${2:column_name} ${3:data_type}\n)', description: 'Create new table' },
    { name: 'ALTER TABLE ADD', snippet: 'ALTER TABLE ${1:table_name} ADD ${2:column_name} ${3:data_type}', description: 'Add column to table' },
    { name: 'DROP TABLE', snippet: 'DROP TABLE ${1:table_name}', description: 'Delete table' },
    { name: 'JOIN', snippet: '${1:LEFT} JOIN ${2:table_name} ON ${3:condition}', description: 'Join tables' },
    { name: 'GROUP BY', snippet: 'GROUP BY ${1:column}', description: 'Group results' },
    { name: 'ORDER BY', snippet: 'ORDER BY ${1:column} ${2:ASC}', description: 'Sort results' },
    { name: 'INNER JOIN', snippet: 'INNER JOIN ${1:table_name} ON ${2:condition}', description: 'Inner join tables' },
    { name: 'LEFT JOIN', snippet: 'LEFT JOIN ${1:table_name} ON ${2:condition}', description: 'Left join tables' },
    { name: 'RIGHT JOIN', snippet: 'RIGHT JOIN ${1:table_name} ON ${2:condition}', description: 'Right join tables' },
]

interface SQLEditorProps {
    value: string
    onChange?: (value: string | undefined) => void
    height?: string
    readOnly?: boolean
    language?: string
    syntaxError?: SQLSyntaxError
}

export function SQLEditor({
    value,
    onChange,
    height = '500px',
    readOnly = false,
    language = 'sql',
    syntaxError,
}: SQLEditorProps) {
    const { theme } = useTheme()
    const editorRef = useRef<editor.IStandaloneCodeEditor | null>(null)
    const monacoRef = useRef<typeof import('monaco-editor') | null>(null)
    const completionProviderRef = useRef<{ dispose: () => void } | null>(null)

    const handleEditorDidMount: OnMount = (editor, monaco) => {
        editorRef.current = editor
        monacoRef.current = monaco

        // Dispose previous provider if exists
        if (completionProviderRef.current) {
            completionProviderRef.current.dispose()
        }

        // Register SQL autocomplete provider
        completionProviderRef.current = monaco.languages.registerCompletionItemProvider('sql', {
            provideCompletionItems: (model: import('monaco-editor').editor.ITextModel, position: import('monaco-editor').Position) => {
                const word = model.getWordUntilPosition(position)
                const range = {
                    startLineNumber: position.lineNumber,
                    endLineNumber: position.lineNumber,
                    startColumn: word.startColumn,
                    endColumn: word.endColumn,
                }

                const suggestions: import('monaco-editor').languages.CompletionItem[] = []

                // Add SQL keywords
                SQL_KEYWORDS.forEach((keyword) => {
                    suggestions.push({
                        label: keyword,
                        kind: monaco.languages.CompletionItemKind.Keyword,
                        insertText: keyword,
                        range: range,
                        detail: 'SQL Keyword',
                    })
                    // Also add lowercase version
                    suggestions.push({
                        label: keyword.toLowerCase(),
                        kind: monaco.languages.CompletionItemKind.Keyword,
                        insertText: keyword.toLowerCase(),
                        range: range,
                        detail: 'SQL Keyword',
                    })
                })

                // Add SQL functions with snippets
                SQL_FUNCTIONS.forEach((func) => {
                    suggestions.push({
                        label: func.name,
                        kind: monaco.languages.CompletionItemKind.Function,
                        insertText: func.snippet,
                        insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
                        range: range,
                        detail: 'SQL Function',
                        documentation: func.description,
                    })
                    // Also add lowercase version
                    suggestions.push({
                        label: func.name.toLowerCase(),
                        kind: monaco.languages.CompletionItemKind.Function,
                        insertText: func.snippet.toLowerCase(),
                        insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
                        range: range,
                        detail: 'SQL Function',
                        documentation: func.description,
                    })
                })

                // Add SQL snippets
                SQL_SNIPPETS.forEach((snippet) => {
                    suggestions.push({
                        label: snippet.name,
                        kind: monaco.languages.CompletionItemKind.Snippet,
                        insertText: snippet.snippet,
                        insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
                        range: range,
                        detail: 'SQL Snippet',
                        documentation: snippet.description,
                    })
                })

                return { suggestions }
            },
            triggerCharacters: [' ', '.', '('],
        })

        editor.updateOptions({
            minimap: { enabled: true },
            fontSize: 14,
            wordWrap: 'on',
            automaticLayout: true,
            scrollBeyondLastLine: false,
            formatOnPaste: true,
            formatOnType: true,
            quickSuggestions: true,
            suggestOnTriggerCharacters: true,
        })
    }

    useEffect(() => {
        if (!editorRef.current || !monacoRef.current) return

        const model = editorRef.current.getModel()
        if (!model) return

        const monaco = monacoRef.current

        if (syntaxError) {
            const line = syntaxError.line || 1
            const column = syntaxError.column || 1
            const lineContent = model.getLineContent(Math.min(line, model.getLineCount()))
            const endColumn = lineContent.length + 1

            const markers: editor.IMarkerData[] = [
                {
                    severity: monaco.MarkerSeverity.Error,
                    message: syntaxError.message,
                    startLineNumber: line,
                    startColumn: column,
                    endLineNumber: line,
                    endColumn: endColumn,
                },
            ]

            monaco.editor.setModelMarkers(model, 'sql-checker', markers)
        } else {
            monaco.editor.setModelMarkers(model, 'sql-checker', [])
        }
    }, [syntaxError])

    useEffect(() => {
        return () => {
            if (editorRef.current && monacoRef.current) {
                const model = editorRef.current.getModel()
                if (model) {
                    monacoRef.current.editor.setModelMarkers(model, 'sql-checker', [])
                }
            }
        }
    }, [])

    return (
        <div className="border rounded-lg overflow-hidden" style={{ height }}>
            <Editor
                height="100%"
                language={language}
                value={value}
                onChange={onChange}
                theme={theme === 'dark' ? 'vs-dark' : 'light'}
                onMount={handleEditorDidMount}
                options={{
                    readOnly,
                    minimap: { enabled: true },
                    fontSize: 14,
                    wordWrap: 'on',
                    automaticLayout: true,
                    scrollBeyondLastLine: false,
                    formatOnPaste: true,
                    formatOnType: true,
                    tabSize: 2,
                    insertSpaces: true,
                    renderWhitespace: 'selection',
                    lineNumbers: 'on',
                    folding: true,
                    bracketPairColorization: {
                        enabled: true,
                    },
                }}
            />
        </div>
    )
}
