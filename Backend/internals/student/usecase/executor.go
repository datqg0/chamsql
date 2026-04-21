package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"backend/pkgs/runner"
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
	runner runner.Runner
}

func NewCodeExecutor(queryRunner runner.Runner) CodeExecutor {
	return &codeExecutor{
		runner: queryRunner,
	}
}

func (ce *codeExecutor) ExecuteCode(ctx context.Context, code, initScript, solutionQuery string, databaseType string, timeout time.Duration) (*ExecutionResult, error) {
	if strings.TrimSpace(code) == "" {
		return &ExecutionResult{
			Success:      false,
			ErrorMessage: "code cannot be empty",
		}, nil
	}

	startTime := time.Now()
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result := &ExecutionResult{Success: false}

	dbType, err := normalizeDBType(databaseType)
	if err != nil {
		result.ErrorMessage = err.Error()
		result.ExecutionTime = int32(time.Since(startTime).Milliseconds())
		return result, nil
	}

	actualResult, err := ce.runner.ExecuteWithSetup(ctxWithTimeout, dbType, initScript, strings.TrimSpace(code))
	if err != nil || actualResult.Error != "" {
		result.ErrorMessage = extractRunnerError(actualResult, err)
		result.ExecutionTime = resolveExecutionTime(actualResult, startTime)
		return result, nil
	}

	result.Output = rowsToMaps(actualResult.Columns, actualResult.Rows)
	result.Success = true
	result.ExecutionTime = int32(actualResult.ExecutionMs)
	result.IsCorrect = true
	result.Score = 100.0

	if strings.TrimSpace(solutionQuery) == "" {
		return result, nil
	}

	expectedResult, expectedErr := ce.runner.ExecuteWithSetup(ctxWithTimeout, dbType, initScript, strings.TrimSpace(solutionQuery))
	if expectedErr != nil || expectedResult.Error != "" {
		result.Success = false
		result.IsCorrect = false
		result.Score = 0.0
		result.ErrorMessage = fmt.Sprintf("expected query error: %s", extractRunnerError(expectedResult, expectedErr))
		result.ExecutionTime = resolveExecutionTime(expectedResult, startTime)
		return result, nil
	}

	result.ExpectedOutput = rowsToMaps(expectedResult.Columns, expectedResult.Rows)
	compareResult := ce.runner.Compare(expectedResult, actualResult, true)
	result.IsCorrect = compareResult.IsCorrect
	if !result.IsCorrect {
		result.Score = 0.0
	}

	return result, nil
}

func normalizeDBType(databaseType string) (runner.DBType, error) {
	switch strings.ToLower(strings.TrimSpace(databaseType)) {
	case "", string(runner.DBTypePostgreSQL):
		return runner.DBTypePostgreSQL, nil
	case string(runner.DBTypeMySQL):
		return runner.DBTypeMySQL, nil
	case string(runner.DBTypeSQLServer):
		return runner.DBTypeSQLServer, nil
	default:
		return "", fmt.Errorf("unsupported database type: %s", databaseType)
	}
}

func rowsToMaps(columns []string, rows [][]interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		entry := make(map[string]interface{}, len(row))
		for i, value := range row {
			key := fmt.Sprintf("column_%d", i+1)
			if i < len(columns) {
				key = columns[i]
			}
			entry[key] = value
		}
		result = append(result, entry)
	}
	return result
}

func extractRunnerError(result *runner.QueryResult, execErr error) string {
	if result != nil && result.Error != "" {
		return result.Error
	}
	if execErr != nil {
		return execErr.Error()
	}
	return "query execution failed"
}

func resolveExecutionTime(result *runner.QueryResult, startTime time.Time) int32 {
	if result != nil && result.ExecutionMs > 0 {
		return int32(result.ExecutionMs)
	}
	return int32(time.Since(startTime).Milliseconds())
}
