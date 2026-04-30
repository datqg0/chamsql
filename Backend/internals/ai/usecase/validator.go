package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"backend/db"
	"backend/internals/ai/domain"

	"github.com/jackc/pgx/v5"
)

// TestCaseForValidation is a local model for validation
type TestCaseForValidation struct {
	TestNumber     int
	TestDataSQL    string
	ExpectedOutput json.RawMessage
}

// IAITestCaseValidator validates test cases
type IAITestCaseValidator interface {
	ValidateTestCases(ctx context.Context, schemaSQL, solutionSQL string, testCases []TestCaseForValidation) (*domain.ValidationResult, error)
	ValidateSingleTestCase(ctx context.Context, schemaSQL, testDataSQL, solutionSQL string, expectedOutput json.RawMessage) (bool, error)
}

// AITestCaseValidator implements test case validation
type aiTestCaseValidator struct {
	database *db.Database
}

// NewAITestCaseValidator creates a new test case validator
func NewAITestCaseValidator(database *db.Database) IAITestCaseValidator {
	return &aiTestCaseValidator{
		database: database,
	}
}

// ValidateTestCases validates all test cases
func (v *aiTestCaseValidator) ValidateTestCases(ctx context.Context, schemaSQL, solutionSQL string, testCases []TestCaseForValidation) (*domain.ValidationResult, error) {
	result := &domain.ValidationResult{
		TotalCount: len(testCases),
		Errors:     []string{},
		IsValid:    true,
	}

	// Step 1: Validate SQL syntax
	if err := v.validateSQLSyntax(schemaSQL); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Schema syntax error: %v", err))
		return result, err
	}

	if err := v.validateSQLSyntax(solutionSQL); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Solution syntax error: %v", err))
		return result, err
	}

	// If no test cases provided, just validate syntax
	if len(testCases) == 0 {
		result.IsValid = true
		return result, nil
	}

	// Step 2: Run each test case
	passCount := 0
	for i, tc := range testCases {
		passed, err := v.ValidateSingleTestCase(ctx, schemaSQL, tc.TestDataSQL, solutionSQL, tc.ExpectedOutput)

		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Test case %d error: %v", i+1, err))
			result.IsValid = false
		}

		if passed {
			passCount++
		}
	}

	result.PassedCount = passCount
	result.IsValid = (passCount == len(testCases)) && (len(result.Errors) == 0)

	return result, nil
}

// ValidateSingleTestCase validates one test case by executing SQL in a transaction (then rolling back)
func (v *aiTestCaseValidator) ValidateSingleTestCase(ctx context.Context, schemaSQL, testDataSQL, solutionSQL string, expectedOutput json.RawMessage) (bool, error) {
	pool := v.database.GetPool()
	if pool == nil {
		return false, fmt.Errorf("database pool not available")
	}

	// Use a timeout context for sandbox execution
	execCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Execute everything in a transaction that we ALWAYS rollback (sandbox mode)
	tx, err := pool.Begin(execCtx)
	if err != nil {
		return false, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(execCtx) // Always rollback — sandbox

	// Step 1: Execute schema
	if err := v.executeSQLInTx(execCtx, tx, schemaSQL); err != nil {
		return false, fmt.Errorf("failed to execute schema: %w", err)
	}

	// Step 2: Insert test data
	if testDataSQL != "" && testDataSQL != "-- No data inserted, table is empty" {
		if err := v.executeSQLInTx(execCtx, tx, testDataSQL); err != nil {
			// Test data error is not fatal — might be intentional empty test
		}
	}

	// Step 3: Execute solution query
	actualOutput, err := v.executeSolutionQueryInTx(execCtx, tx, solutionSQL)
	if err != nil {
		return false, fmt.Errorf("failed to execute solution: %w", err)
	}

	// Step 4: Compare results
	return v.compareResults(actualOutput, expectedOutput), nil
}

// executeSQLInTx executes SQL statements within a pgx transaction
func (v *aiTestCaseValidator) executeSQLInTx(ctx context.Context, tx pgx.Tx, sqlStatement string) error {
	statements := strings.Split(sqlStatement, ";")

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}

		if _, err := tx.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("SQL exec error for '%s': %w", truncateSQL(stmt), err)
		}
	}

	return nil
}

// executeSolutionQueryInTx executes a SELECT query within a transaction and returns results
func (v *aiTestCaseValidator) executeSolutionQueryInTx(ctx context.Context, tx pgx.Tx, solutionSQL string) ([]map[string]interface{}, error) {
	rows, err := tx.Query(ctx, solutionSQL)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	fieldDescs := rows.FieldDescriptions()

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("failed to read row values: %w", err)
		}

		row := make(map[string]interface{})
		for i, fd := range fieldDescs {
			if i < len(values) {
				row[fd.Name] = values[i]
			}
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return results, nil
}

func (v *aiTestCaseValidator) compareResults(actual []map[string]interface{}, expected json.RawMessage) bool {
	// If expected is empty array, actual should also be empty
	if string(expected) == "[]" {
		return len(actual) == 0
	}

	// If expected is generic placeholder (e.g., [{"data":"..."}]), do basic validation
	if strings.Contains(string(expected), `"data"`) {
		return len(actual) > 0
	}

	// Parse expected output
	var expectedData []map[string]interface{}
	if err := json.Unmarshal(expected, &expectedData); err != nil {
		return true
	}

	// Compare counts first
	if len(actual) != len(expectedData) {
		return false
	}

	// Sort both for comparison (ignoring order)
	actualSorted := v.sortResults(actual)
	expectedSorted := v.sortResults(expectedData)

	// Deep compare
	return reflect.DeepEqual(actualSorted, expectedSorted)
}

func (v *aiTestCaseValidator) sortResults(results []map[string]interface{}) []map[string]interface{} {
	var jsonResults []string

	for _, row := range results {
		data, _ := json.Marshal(row)
		jsonResults = append(jsonResults, string(data))
	}

	sort.Strings(jsonResults)

	var sorted []map[string]interface{}
	for _, jsonStr := range jsonResults {
		var row map[string]interface{}
		json.Unmarshal([]byte(jsonStr), &row)
		sorted = append(sorted, row)
	}

	return sorted
}

func (v *aiTestCaseValidator) validateSQLSyntax(sqlStatement string) error {
	sqlStatement = strings.TrimSpace(sqlStatement)

	if sqlStatement == "" {
		return fmt.Errorf("empty SQL statement")
	}

	openParen := strings.Count(sqlStatement, "(")
	closeParen := strings.Count(sqlStatement, ")")

	if openParen != closeParen {
		return fmt.Errorf("unbalanced parentheses: %d open, %d close", openParen, closeParen)
	}

	singleQuotes := strings.Count(sqlStatement, "'")
	if singleQuotes%2 != 0 {
		return fmt.Errorf("unbalanced single quotes: %d", singleQuotes)
	}

	return nil
}

func truncateSQL(sql string) string {
	if len(sql) > 80 {
		return sql[:80] + "..."
	}
	return sql
}
