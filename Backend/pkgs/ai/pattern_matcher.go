package infrastructure

import (
	"fmt"
	"regexp"
	"strings"
)

// PatternMatcher provides pattern-based SQL generation
type PatternMatcher struct{}

// NewPatternMatcher creates a new pattern matcher
func NewPatternMatcher() *PatternMatcher {
	return &PatternMatcher{}
}

// MatchPattern tries to match description against patterns
// Returns: (generatedSQL, confidence score 0-1)
func (pm *PatternMatcher) MatchPattern(description string, schemaSQL string) (string, float64) {
	desc := strings.ToLower(description)

	// Pattern 1: Find Top N records
	if re := regexp.MustCompile(`find\s+(?:top\s+)?(\d+)\s+(?:records?|rows?|items?|.*?)`); re.MatchString(desc) {
		matches := re.FindStringSubmatch(desc)
		if len(matches) > 0 {
			n := matches[1]
			// Detect what to order by
			if strings.Contains(desc, "highest") || strings.Contains(desc, "max") || strings.Contains(desc, "largest") {
				return fmt.Sprintf("SELECT * FROM ? ORDER BY ? DESC LIMIT %s;", n), 0.8
			}
			if strings.Contains(desc, "lowest") || strings.Contains(desc, "min") || strings.Contains(desc, "smallest") {
				return fmt.Sprintf("SELECT * FROM ? ORDER BY ? ASC LIMIT %s;", n), 0.8
			}
		}
	}

	// Pattern 2: Count with WHERE clause
	if strings.Contains(desc, "count") {
		if strings.Contains(desc, "where") || strings.Contains(desc, "condition") {
			return "SELECT COUNT(*) as count FROM ? WHERE ?;", 0.7
		}
		return "SELECT COUNT(*) as count FROM ?;", 0.75
	}

	// Pattern 3: Group By and Aggregate
	if strings.Contains(desc, "group by") || strings.Contains(desc, "grouped by") {
		if strings.Contains(desc, "sum") {
			return "SELECT ?, SUM(?) as total FROM ? GROUP BY ?;", 0.7
		}
		if strings.Contains(desc, "average") || strings.Contains(desc, "avg") {
			return "SELECT ?, AVG(?) as average FROM ? GROUP BY ?;", 0.7
		}
		if strings.Contains(desc, "count") {
			return "SELECT ?, COUNT(*) as count FROM ? GROUP BY ?;", 0.7
		}
		return "SELECT ?, ? FROM ? GROUP BY ?;", 0.6
	}

	// Pattern 4: JOIN tables
	if strings.Contains(desc, "join") {
		if strings.Contains(desc, "inner") {
			return "SELECT * FROM ? INNER JOIN ? ON ? = ?;", 0.65
		}
		if strings.Contains(desc, "left") {
			return "SELECT * FROM ? LEFT JOIN ? ON ? = ?;", 0.65
		}
		if strings.Contains(desc, "right") {
			return "SELECT * FROM ? RIGHT JOIN ? ON ? = ?;", 0.65
		}
		return "SELECT * FROM ? JOIN ? ON ? = ?;", 0.6
	}

	// Pattern 5: DISTINCT records
	if strings.Contains(desc, "distinct") || strings.Contains(desc, "unique") {
		return "SELECT DISTINCT ? FROM ?;", 0.75
	}

	// Pattern 6: ORDER BY
	if strings.Contains(desc, "sort") || strings.Contains(desc, "order") {
		if strings.Contains(desc, "descending") || strings.Contains(desc, "highest") || strings.Contains(desc, "largest") {
			return "SELECT * FROM ? ORDER BY ? DESC;", 0.7
		}
		return "SELECT * FROM ? ORDER BY ? ASC;", 0.7
	}

	// Pattern 7: Simple SELECT WHERE
	if strings.Contains(desc, "where") || strings.Contains(desc, "condition") || strings.Contains(desc, "filter") {
		return "SELECT * FROM ? WHERE ?;", 0.6
	}

	// Pattern 8: Simple SELECT
	if strings.Contains(desc, "select") || strings.Contains(desc, "find") || strings.Contains(desc, "get") {
		return "SELECT * FROM ?;", 0.5
	}

	// No pattern matched
	return "", 0.0
}

// GenerateTestCaseData generates basic test case data
func (pm *PatternMatcher) GenerateTestCaseData(schemaSQL string, count int) []string {
	testCases := []string{}

	// Pattern: If schema mentions "employees", generate employee test data
	if strings.Contains(strings.ToLower(schemaSQL), "employee") {
		testCases = append(testCases, `
INSERT INTO employees (id, name, salary, department) VALUES
(1, 'Alice Johnson', 75000, 'Engineering'),
(2, 'Bob Smith', 65000, 'Sales'),
(3, 'Charlie Brown', 55000, 'Engineering'),
(4, 'Diana Prince', 85000, 'Management');
`)
	}

	// Pattern: If schema mentions "customers", generate customer test data
	if strings.Contains(strings.ToLower(schemaSQL), "customer") {
		testCases = append(testCases, `
INSERT INTO customers (id, name, email, created_at) VALUES
(1, 'Customer A', 'a@example.com', '2024-01-01'),
(2, 'Customer B', 'b@example.com', '2024-02-01'),
(3, 'Customer C', 'c@example.com', '2024-03-01');
`)
	}

	// Pattern: If schema mentions "orders", generate order test data
	if strings.Contains(strings.ToLower(schemaSQL), "orders") {
		testCases = append(testCases, `
INSERT INTO orders (id, customer_id, amount, created_at) VALUES
(1, 1, 100.00, '2024-01-15'),
(2, 1, 150.00, '2024-01-20'),
(3, 2, 200.00, '2024-02-10');
`)
	}

	return testCases
}

// ValidateSQLSyntax checks if SQL has basic syntax correctness
func (pm *PatternMatcher) ValidateSQLSyntax(sql string) bool {
	sql = strings.TrimSpace(sql)

	// Must start with SELECT/INSERT/UPDATE/DELETE
	validStarts := []string{"SELECT", "INSERT", "UPDATE", "DELETE"}
	found := false
	for _, start := range validStarts {
		if strings.HasPrefix(strings.ToUpper(sql), start) {
			found = true
			break
		}
	}
	if !found {
		return false
	}

	// Must end with semicolon (optional but preferred)
	// Must have balanced parentheses
	openParen := strings.Count(sql, "(")
	closeParen := strings.Count(sql, ")")
	if openParen != closeParen {
		return false
	}

	return true
}
