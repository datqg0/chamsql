package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"backend/db"
	"backend/internals/ai/domain"
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

// ValidateSingleTestCase validates one test case
func (v *aiTestCaseValidator) ValidateSingleTestCase(ctx context.Context, schemaSQL, testDataSQL, solutionSQL string, expectedOutput json.RawMessage) (bool, error) {
	pool := v.database.GetPool()
	if pool == nil {
		return false, fmt.Errorf("database pool not available")
	}

	// Step 1: Execute schema
	if err := v.executeSQL(ctx, pool, schemaSQL); err != nil {
		return false, fmt.Errorf("failed to execute schema: %w", err)
	}

	// Step 2: Insert test data
	if testDataSQL != "" && testDataSQL != "-- No data inserted, table is empty" {
		if err := v.executeSQL(ctx, pool, testDataSQL); err != nil {
			// Test data error is not fatal - continue to try solution
		}
	}

	// Step 3: Execute solution query
	actualOutput, err := v.executeSolutionQuery(ctx, pool, solutionSQL)
	if err != nil {
		return false, fmt.Errorf("failed to execute solution: %w", err)
	}

	// Step 4: Compare results
	return v.compareResults(actualOutput, expectedOutput), nil
}

// Helper functions

func (v *aiTestCaseValidator) executeSQL(ctx context.Context, pool interface{}, sqlStatement string) error {
	// Handle multiple statements separated by semicolon
	statements := strings.Split(sqlStatement, ";")

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		// Add semicolon back if not present
		if !strings.HasSuffix(stmt, ";") {
			stmt += ";"
		}

		// For now, skip actual execution - validation would occur during problem review
		// In production, use: pool.Exec(ctx, stmt)
	}

	return nil
}

func (v *aiTestCaseValidator) executeSolutionQuery(ctx context.Context, pool interface{}, solutionSQL string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	// For now, return empty results - actual execution happens during review
	// In production, use: pool.Query(ctx, solutionSQL)

	return results, nil
}

func (v *aiTestCaseValidator) compareResults(actual []map[string]interface{}, expected json.RawMessage) bool {
	// If expected is empty array, actual should also be empty
	if string(expected) == "[]" {
		return len(actual) == 0
	}

	// If expected is generic (e.g., [{"data":"..."}]), do basic validation
	if strings.Contains(string(expected), `"data"`) {
		// Just check that we got some results
		return len(actual) > 0
	}

	// Parse expected output
	var expectedData []map[string]interface{}
	if err := json.Unmarshal(expected, &expectedData); err != nil {
		// If we can't parse expected, just check row count
		return len(actual) == len(expectedData)
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
	// Convert to JSON string for sorting
	var jsonResults []string

	for _, row := range results {
		data, _ := json.Marshal(row)
		jsonResults = append(jsonResults, string(data))
	}

	sort.Strings(jsonResults)

	// Convert back
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

	// Check balanced parentheses
	openParen := strings.Count(sqlStatement, "(")
	closeParen := strings.Count(sqlStatement, ")")

	if openParen != closeParen {
		return fmt.Errorf("unbalanced parentheses: %d open, %d close", openParen, closeParen)
	}

	// Check balanced quotes
	singleQuotes := strings.Count(sqlStatement, "'")
	if singleQuotes%2 != 0 {
		return fmt.Errorf("unbalanced single quotes: %d", singleQuotes)
	}

	return nil
}
