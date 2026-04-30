package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"backend/db"
	"backend/internals/ai/domain"
	"backend/pkgs/ai"
	"backend/sql/models"
)

// IAITestCaseGenerator defines test case generation service
type IAITestCaseGenerator interface {
	GenerateTestCases(ctx context.Context, req domain.TestCaseGenerationInput) ([]domain.TestCaseGenerated, error)
}

// AITestCaseGenerator implements test case generation
type aiTestCaseGenerator struct {
	database  *db.Database
	queries   *models.Queries
	llmClient ai.LLMClient
	provider  string
}

// NewAITestCaseGenerator creates a new test case generator
func NewAITestCaseGenerator(
	database *db.Database,
	llmClient ai.LLMClient,
	provider string,
) IAITestCaseGenerator {
	return &aiTestCaseGenerator{
		database:  database,
		queries:   models.New(database.GetPool()),
		llmClient: llmClient,
		provider:  provider,
	}
}

// GenerateTestCases generates test cases from schema and solution
func (g *aiTestCaseGenerator) GenerateTestCases(ctx context.Context, req domain.TestCaseGenerationInput) ([]domain.TestCaseGenerated, error) {
	var testCases []domain.TestCaseGenerated

	// Step 1: Parse schema to understand table structure
	tables := g.parseSchema(req.SchemaSQL)
	if len(tables) == 0 {
		return nil, fmt.Errorf("could not parse tables from schema")
	}

	// Step 2: Generate basic test data variations
	basicTestCases := g.generateBasicTestCases(req.SchemaSQL, tables)
	testCases = append(testCases, basicTestCases...)

	// Step 3: Generate boundary test cases
	boundaryTestCases := g.generateBoundaryTestCases(req.SchemaSQL, tables)
	testCases = append(testCases, boundaryTestCases...)

	// Step 4: Generate edge case test cases
	edgeCaseTestCases := g.generateEdgeCaseTestCases(req.SchemaSQL, tables)
	testCases = append(testCases, edgeCaseTestCases...)

	// Step 5: Try LLM generation for higher quality data (Tier 2 Upgrade)
	if g.llmClient != nil {
		llmSQL, err := g.llmClient.GenerateTestCase(ctx, req.Description, req.SchemaSQL, req.SolutionSQL)
		if err == nil && llmSQL != "" {
			testCases = append(testCases, domain.TestCaseGenerated{
				TestNumber:     len(testCases) + 1,
				Description:    "AI Generated specialized test data",
				TestDataSQL:    llmSQL,
				ExpectedOutput: json.RawMessage(`[]`), // To be validated by validator
				IsPublic:       false,
				Difficulty:     "hard",
			})
		}
	}

	// Step 6: Assign public/hidden based on count preference
	testCases = g.assignPublicHidden(testCases, req.PublicTestCaseCount)

	// Step 7: Add descriptions for each test case
	testCases = g.addTestCaseDescriptions(testCases, req.Description)

	return testCases, nil
}

// TableSchema represents parsed table structure
type TableSchema struct {
	Name    string
	Columns []ColumnSchema
}

type ColumnSchema struct {
	Name     string
	DataType string
	IsNull   bool
}

// parseSchema extracts table and column information from SQL
func (g *aiTestCaseGenerator) parseSchema(schemaSQL string) []TableSchema {
	var tables []TableSchema

	// Split by CREATE TABLE statements
	parts := strings.Split(strings.ToUpper(schemaSQL), "CREATE TABLE")

	for _, part := range parts[1:] { // Skip first empty part
		table := g.parseTableDefinition(part)
		if table.Name != "" {
			tables = append(tables, table)
		}
	}

	return tables
}

func (g *aiTestCaseGenerator) parseTableDefinition(def string) TableSchema {
	table := TableSchema{}

	// Extract table name
	lines := strings.Split(def, "\n")
	if len(lines) > 0 {
		tableLine := strings.TrimSpace(lines[0])
		tableLine = strings.TrimSuffix(tableLine, "(")
		parts := strings.Fields(tableLine)
		if len(parts) > 0 {
			table.Name = strings.Trim(parts[0], "(),;")
		}
	}

	// Extract columns
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ")") || strings.HasPrefix(line, "PRIMARY") || strings.HasPrefix(line, "FOREIGN") || strings.HasPrefix(line, "UNIQUE") {
			continue
		}

		col := g.parseColumn(line)
		if col.Name != "" {
			table.Columns = append(table.Columns, col)
		}
	}

	return table
}

func (g *aiTestCaseGenerator) parseColumn(colDef string) ColumnSchema {
	col := ColumnSchema{}

	// Remove trailing comma and comments
	colDef = strings.TrimSuffix(colDef, ",")
	colDef = strings.Split(colDef, "--")[0]
	colDef = strings.TrimSpace(colDef)

	parts := strings.Fields(colDef)
	if len(parts) < 2 {
		return col
	}

	col.Name = parts[0]
	col.DataType = parts[1]
	col.IsNull = !strings.Contains(strings.ToUpper(colDef), "NOT NULL")

	return col
}

// generateBasicTestCases generates common test data scenarios
func (g *aiTestCaseGenerator) generateBasicTestCases(schemaSQL string, tables []TableSchema) []domain.TestCaseGenerated {
	var testCases []domain.TestCaseGenerated

	// Test Case 1: Empty table
	testCases = append(testCases, domain.TestCaseGenerated{
		TestNumber:     1,
		Description:    "Empty table - no data",
		TestDataSQL:    "-- No data inserted, table is empty",
		ExpectedOutput: json.RawMessage(`[]`),
		IsPublic:       true,
		Difficulty:     "easy",
	})

	// Test Case 2: Single row
	testCases = append(testCases, domain.TestCaseGenerated{
		TestNumber:     2,
		Description:    "Single row - basic functionality",
		TestDataSQL:    g.generateSingleRowData(tables[0]),
		ExpectedOutput: json.RawMessage(`[{"data":"minimal"}]`),
		IsPublic:       true,
		Difficulty:     "easy",
	})

	// Test Case 3: Multiple rows
	testCases = append(testCases, domain.TestCaseGenerated{
		TestNumber:     3,
		Description:    "Multiple rows - normal case",
		TestDataSQL:    g.generateMultipleRowData(tables[0], 5),
		ExpectedOutput: json.RawMessage(`[{"data":"5rows"}]`),
		IsPublic:       true,
		Difficulty:     "medium",
	})

	return testCases
}

// generateBoundaryTestCases generates boundary value test cases
func (g *aiTestCaseGenerator) generateBoundaryTestCases(schemaSQL string, tables []TableSchema) []domain.TestCaseGenerated {
	var testCases []domain.TestCaseGenerated

	// Test Case 4: NULL values
	testCases = append(testCases, domain.TestCaseGenerated{
		TestNumber:     4,
		Description:    "Boundary: NULL values",
		TestDataSQL:    g.generateNullValueData(tables[0]),
		ExpectedOutput: json.RawMessage(`[{"data":"with_nulls"}]`),
		IsPublic:       false,
		Difficulty:     "hard",
	})

	// Test Case 5: Large values
	testCases = append(testCases, domain.TestCaseGenerated{
		TestNumber:     5,
		Description:    "Boundary: Large numeric values",
		TestDataSQL:    g.generateLargeValueData(tables[0]),
		ExpectedOutput: json.RawMessage(`[{"data":"large_values"}]`),
		IsPublic:       false,
		Difficulty:     "hard",
	})

	// Test Case 6: Empty strings
	testCases = append(testCases, domain.TestCaseGenerated{
		TestNumber:     6,
		Description:    "Boundary: Empty strings",
		TestDataSQL:    g.generateEmptyStringData(tables[0]),
		ExpectedOutput: json.RawMessage(`[{"data":"empty_strings"}]`),
		IsPublic:       false,
		Difficulty:     "medium",
	})

	return testCases
}

// generateEdgeCaseTestCases generates edge case test cases
func (g *aiTestCaseGenerator) generateEdgeCaseTestCases(schemaSQL string, tables []TableSchema) []domain.TestCaseGenerated {
	var testCases []domain.TestCaseGenerated

	// Test Case 7: Duplicate values
	testCases = append(testCases, domain.TestCaseGenerated{
		TestNumber:     7,
		Description:    "Edge case: Duplicate values",
		TestDataSQL:    g.generateDuplicateValueData(tables[0]),
		ExpectedOutput: json.RawMessage(`[{"data":"duplicates"}]`),
		IsPublic:       false,
		Difficulty:     "medium",
	})

	// Test Case 8: Special characters
	testCases = append(testCases, domain.TestCaseGenerated{
		TestNumber:     8,
		Description:    "Edge case: Special characters in strings",
		TestDataSQL:    g.generateSpecialCharData(tables[0]),
		ExpectedOutput: json.RawMessage(`[{"data":"special_chars"}]`),
		IsPublic:       false,
		Difficulty:     "hard",
	})

	return testCases
}

// Data generation helpers

func (g *aiTestCaseGenerator) generateSingleRowData(table TableSchema) string {
	insertSQL := fmt.Sprintf("INSERT INTO %s ", table.Name)

	if len(table.Columns) == 0 {
		return insertSQL + "DEFAULT VALUES;"
	}

	// Build column list
	var colNames []string
	var colValues []string

	for _, col := range table.Columns {
		colNames = append(colNames, col.Name)

		// Generate appropriate value based on data type
		value := g.getDefaultValue(col)
		colValues = append(colValues, value)
	}

	insertSQL += fmt.Sprintf("(%s) VALUES (%s);", strings.Join(colNames, ", "), strings.Join(colValues, ", "))
	return insertSQL
}

func (g *aiTestCaseGenerator) generateMultipleRowData(table TableSchema, rowCount int) string {
	if len(table.Columns) == 0 {
		return fmt.Sprintf("INSERT INTO %s DEFAULT VALUES;", table.Name)
	}

	insertSQL := fmt.Sprintf("INSERT INTO %s (", table.Name)

	// Column names
	var colNames []string
	for _, col := range table.Columns {
		colNames = append(colNames, col.Name)
	}
	insertSQL += strings.Join(colNames, ", ") + ") VALUES"

	// Generate multiple rows
	var rows []string
	for i := 1; i <= rowCount; i++ {
		var values []string
		for _, col := range table.Columns {
			value := g.getVariantValue(col, i)
			values = append(values, value)
		}
		rows = append(rows, fmt.Sprintf("(%s)", strings.Join(values, ", ")))
	}

	insertSQL += strings.Join(rows, ", ") + ";"
	return insertSQL
}

func (g *aiTestCaseGenerator) generateNullValueData(table TableSchema) string {
	if len(table.Columns) == 0 {
		return fmt.Sprintf("INSERT INTO %s DEFAULT VALUES;", table.Name)
	}

	insertSQL := fmt.Sprintf("INSERT INTO %s (", table.Name)

	var colNames []string
	for _, col := range table.Columns {
		colNames = append(colNames, col.Name)
	}
	insertSQL += strings.Join(colNames, ", ") + ") VALUES"

	var values []string
	for _, col := range table.Columns {
		if col.IsNull {
			values = append(values, "NULL")
		} else {
			values = append(values, g.getDefaultValue(col))
		}
	}

	insertSQL += fmt.Sprintf("(%s);", strings.Join(values, ", "))
	return insertSQL
}

func (g *aiTestCaseGenerator) generateLargeValueData(table TableSchema) string {
	if len(table.Columns) == 0 {
		return fmt.Sprintf("INSERT INTO %s DEFAULT VALUES;", table.Name)
	}

	insertSQL := fmt.Sprintf("INSERT INTO %s (", table.Name)

	var colNames []string
	for _, col := range table.Columns {
		colNames = append(colNames, col.Name)
	}
	insertSQL += strings.Join(colNames, ", ") + ") VALUES"

	var values []string
	for _, col := range table.Columns {
		if strings.Contains(strings.ToUpper(col.DataType), "INT") {
			values = append(values, "9223372036854775807") // Max int64
		} else if strings.Contains(strings.ToUpper(col.DataType), "VARCHAR") {
			values = append(values, "'"+strings.Repeat("x", 100)+"'")
		} else {
			values = append(values, g.getDefaultValue(col))
		}
	}

	insertSQL += fmt.Sprintf("(%s);", strings.Join(values, ", "))
	return insertSQL
}

func (g *aiTestCaseGenerator) generateEmptyStringData(table TableSchema) string {
	if len(table.Columns) == 0 {
		return fmt.Sprintf("INSERT INTO %s DEFAULT VALUES;", table.Name)
	}

	insertSQL := fmt.Sprintf("INSERT INTO %s (", table.Name)

	var colNames []string
	for _, col := range table.Columns {
		colNames = append(colNames, col.Name)
	}
	insertSQL += strings.Join(colNames, ", ") + ") VALUES"

	var values []string
	for _, col := range table.Columns {
		if strings.Contains(strings.ToUpper(col.DataType), "VARCHAR") || strings.Contains(strings.ToUpper(col.DataType), "TEXT") {
			values = append(values, "''")
		} else {
			values = append(values, g.getDefaultValue(col))
		}
	}

	insertSQL += fmt.Sprintf("(%s);", strings.Join(values, ", "))
	return insertSQL
}

func (g *aiTestCaseGenerator) generateDuplicateValueData(table TableSchema) string {
	if len(table.Columns) == 0 {
		return fmt.Sprintf("INSERT INTO %s DEFAULT VALUES;", table.Name)
	}

	insertSQL := fmt.Sprintf("INSERT INTO %s (", table.Name)

	var colNames []string
	for _, col := range table.Columns {
		colNames = append(colNames, col.Name)
	}
	insertSQL += strings.Join(colNames, ", ") + ") VALUES"

	var rows []string
	for i := 0; i < 3; i++ {
		var values []string
		for _, col := range table.Columns {
			values = append(values, g.getDefaultValue(col))
		}
		rows = append(rows, fmt.Sprintf("(%s)", strings.Join(values, ", ")))
	}

	insertSQL += strings.Join(rows, ", ") + ";"
	return insertSQL
}

func (g *aiTestCaseGenerator) generateSpecialCharData(table TableSchema) string {
	if len(table.Columns) == 0 {
		return fmt.Sprintf("INSERT INTO %s DEFAULT VALUES;", table.Name)
	}

	insertSQL := fmt.Sprintf("INSERT INTO %s (", table.Name)

	var colNames []string
	for _, col := range table.Columns {
		colNames = append(colNames, col.Name)
	}
	insertSQL += strings.Join(colNames, ", ") + ") VALUES"

	var values []string
	for _, col := range table.Columns {
		if strings.Contains(strings.ToUpper(col.DataType), "VARCHAR") || strings.Contains(strings.ToUpper(col.DataType), "TEXT") {
			values = append(values, `'test''s "quoted" &special'`)
		} else {
			values = append(values, g.getDefaultValue(col))
		}
	}

	insertSQL += fmt.Sprintf("(%s);", strings.Join(values, ", "))
	return insertSQL
}

// Helper functions

func (g *aiTestCaseGenerator) getDefaultValue(col ColumnSchema) string {
	dataType := strings.ToUpper(col.DataType)

	if strings.Contains(dataType, "INT") {
		return "1"
	} else if strings.Contains(dataType, "DECIMAL") || strings.Contains(dataType, "FLOAT") || strings.Contains(dataType, "DOUBLE") {
		return "100.50"
	} else if strings.Contains(dataType, "VARCHAR") || strings.Contains(dataType, "TEXT") {
		return "'Test Data'"
	} else if strings.Contains(dataType, "DATE") || strings.Contains(dataType, "TIMESTAMP") {
		return "'2024-01-01'"
	} else if strings.Contains(dataType, "BOOLEAN") {
		return "true"
	}

	if col.IsNull {
		return "NULL"
	}
	return "'default'"
}

func (g *aiTestCaseGenerator) getVariantValue(col ColumnSchema, variant int) string {
	dataType := strings.ToUpper(col.DataType)

	if strings.Contains(dataType, "INT") {
		return fmt.Sprintf("%d", variant*10)
	} else if strings.Contains(dataType, "DECIMAL") || strings.Contains(dataType, "FLOAT") {
		return fmt.Sprintf("%.2f", float64(variant)*10.5)
	} else if strings.Contains(dataType, "VARCHAR") || strings.Contains(dataType, "TEXT") {
		return fmt.Sprintf("'Value %d'", variant)
	} else if strings.Contains(dataType, "DATE") || strings.Contains(dataType, "TIMESTAMP") {
		return fmt.Sprintf("'2024-01-%02d'", variant)
	}

	return g.getDefaultValue(col)
}

// assignPublicHidden assigns whether test cases are public or hidden
func (g *aiTestCaseGenerator) assignPublicHidden(testCases []domain.TestCaseGenerated, publicCount int) []domain.TestCaseGenerated {
	for i := range testCases {
		if i < publicCount {
			testCases[i].IsPublic = true
		} else {
			testCases[i].IsPublic = false
		}
	}
	return testCases
}

// addTestCaseDescriptions adds descriptions to test cases
func (g *aiTestCaseGenerator) addTestCaseDescriptions(testCases []domain.TestCaseGenerated, problemDesc string) []domain.TestCaseGenerated {
	descriptions := []string{
		"Basic case with no data",
		"Single row basic case",
		"Multiple rows normal scenario",
		"Boundary: Testing with NULL values",
		"Boundary: Large numeric values",
		"Boundary: Empty strings",
		"Edge case: Duplicate values",
		"Edge case: Special characters",
	}

	for i := range testCases {
		if i < len(descriptions) {
			testCases[i].Description = descriptions[i]
		}
	}

	return testCases
}
