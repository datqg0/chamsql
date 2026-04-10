package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"backend/db"
	"backend/sql/models"
)

type CodeExecutor interface {
	ExecuteCode(ctx context.Context, code, initScript, solutionQuery string, databaseType string, timeout time.Duration) (*ExecutionResult, error)
}

type ExecutionResult struct {
	Success        bool
	Output         []map[string]interface{}
	ExpectedOutput []map[string]interface{}
	IsCorrect      bool
	Score          float64
	ErrorMessage   string
	ExecutionTime  int32
}

type codeExecutor struct {
	db      *db.Database
	queries *models.Queries
}

func NewCodeExecutor(database *db.Database) CodeExecutor {
	return &codeExecutor{
		db:      database,
		queries: models.New(database.GetPool()),
	}
}

func (ce *codeExecutor) ExecuteCode(ctx context.Context, code, initScript, solutionQuery string, databaseType string, timeout time.Duration) (*ExecutionResult, error) {
	if code == "" {
		return &ExecutionResult{
			Success:      false,
			ErrorMessage: "code cannot be empty",
		}, nil
	}

	startTime := time.Now()
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result := &ExecutionResult{
		Success: false,
	}

	initQueries := []string{}
	if initScript != "" {
		initQueries = ce.parseSQL(initScript)
	}

	studentQueries := ce.parseSQL(code)

	expectedQueries := []string{}
	if solutionQuery != "" {
		expectedQueries = ce.parseSQL(solutionQuery)
	}

	pool := ce.db.GetPool()

	tx, err := pool.Begin(ctxWithTimeout)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("failed to start transaction: %v", err)
		result.ExecutionTime = int32(time.Since(startTime).Milliseconds())
		return result, nil
	}
	defer tx.Rollback(ctxWithTimeout)

	for _, query := range initQueries {
		if query == "" {
			continue
		}
		_, err := tx.Exec(ctxWithTimeout, query)
		if err != nil {
			result.ErrorMessage = fmt.Sprintf("init error: %v", err)
			result.ExecutionTime = int32(time.Since(startTime).Milliseconds())
			return result, nil
		}
	}

	studentOutput := []map[string]interface{}{}

	for _, query := range studentQueries {
		if query == "" {
			continue
		}

		if ce.isModifyingQuery(query) {
			commandTag, err := tx.Exec(ctxWithTimeout, query)
			if err != nil {
				result.ErrorMessage = fmt.Sprintf("execution error: %v", err)
				result.ExecutionTime = int32(time.Since(startTime).Milliseconds())
				return result, nil
			}
			_ = commandTag
		} else {
			rows, err := tx.Query(ctxWithTimeout, query)
			if err != nil {
				result.ErrorMessage = fmt.Sprintf("query error: %v", err)
				result.ExecutionTime = int32(time.Since(startTime).Milliseconds())
				return result, nil
			}

			for rows.Next() {
				values, err := rows.Values()
				if err != nil {
					rows.Close()
					result.ErrorMessage = fmt.Sprintf("row scan error: %v", err)
					result.ExecutionTime = int32(time.Since(startTime).Milliseconds())
					return result, nil
				}

				rowMap := make(map[string]interface{})
				for i, col := range rows.FieldDescriptions() {
					rowMap[string(col.Name)] = values[i]
				}
				studentOutput = append(studentOutput, rowMap)
			}
			rows.Close()
		}
	}

	result.Output = studentOutput
	result.Success = true

	if len(expectedQueries) > 0 {
		expectedOutput := []map[string]interface{}{}

		for _, query := range expectedQueries {
			if query == "" {
				continue
			}

			if !ce.isModifyingQuery(query) {
				rows, err := tx.Query(ctxWithTimeout, query)
				if err != nil {
					result.ErrorMessage = fmt.Sprintf("expected query error: %v", err)
					result.ExecutionTime = int32(time.Since(startTime).Milliseconds())
					return result, nil
				}

				for rows.Next() {
					values, err := rows.Values()
					if err != nil {
						rows.Close()
						break
					}

					rowMap := make(map[string]interface{})
					for i, col := range rows.FieldDescriptions() {
						rowMap[string(col.Name)] = values[i]
					}
					expectedOutput = append(expectedOutput, rowMap)
				}
				rows.Close()
			}
		}

		result.ExpectedOutput = expectedOutput
		result.IsCorrect = ce.compareOutputs(studentOutput, expectedOutput)

		if result.IsCorrect {
			result.Score = 100.0
		} else {
			result.Score = 0.0
		}
	} else {
		result.IsCorrect = true
		result.Score = 100.0
	}

	result.ExecutionTime = int32(time.Since(startTime).Milliseconds())
	return result, nil
}

func (ce *codeExecutor) parseSQL(sql string) []string {
	statements := []string{}
	current := strings.Builder{}

	for _, char := range sql {
		current.WriteRune(char)
		if char == ';' {
			stmt := strings.TrimSpace(current.String())
			if stmt != "" && stmt != ";" {
				statements = append(statements, strings.TrimSuffix(stmt, ";"))
			}
			current.Reset()
		}
	}

	remaining := strings.TrimSpace(current.String())
	if remaining != "" && remaining != ";" {
		statements = append(statements, remaining)
	}

	return statements
}

func (ce *codeExecutor) isModifyingQuery(query string) bool {
	upper := strings.ToUpper(strings.TrimSpace(query))
	modifyingKeywords := []string{"INSERT", "UPDATE", "DELETE", "CREATE", "DROP", "ALTER", "TRUNCATE"}

	for _, keyword := range modifyingKeywords {
		if strings.HasPrefix(upper, keyword) {
			return true
		}
	}

	return false
}

func (ce *codeExecutor) compareOutputs(student, expected []map[string]interface{}) bool {
	if len(student) != len(expected) {
		return false
	}

	if len(student) == 0 && len(expected) == 0 {
		return true
	}

	for i, studentRow := range student {
		if i >= len(expected) {
			return false
		}
		expectedRow := expected[i]

		if len(studentRow) != len(expectedRow) {
			return false
		}

		for key, studentVal := range studentRow {
			expectedVal, exists := expectedRow[key]
			if !exists {
				return false
			}

			if !ce.compareValues(studentVal, expectedVal) {
				return false
			}
		}
	}

	return true
}

func (ce *codeExecutor) compareValues(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch aVal := a.(type) {
	case int32:
		switch bVal := b.(type) {
		case int32:
			return aVal == bVal
		case int64:
			return int64(aVal) == bVal
		case float64:
			return float64(aVal) == bVal
		default:
			return false
		}
	case int64:
		switch bVal := b.(type) {
		case int32:
			return aVal == int64(bVal)
		case int64:
			return aVal == bVal
		case float64:
			return float64(aVal) == bVal
		default:
			return false
		}
	case float64:
		switch bVal := b.(type) {
		case float64:
			return aVal == bVal
		case int32:
			return aVal == float64(bVal)
		case int64:
			return aVal == float64(bVal)
		default:
			return false
		}
	case string:
		if bStr, ok := b.(string); ok {
			return aVal == bStr
		}
		return false
	case bool:
		if bBool, ok := b.(bool); ok {
			return aVal == bBool
		}
		return false
	default:
		return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
	}
}
