package pdf

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// PDFParser handles PDF file parsing
type PDFParser struct{}

// NewPDFParser creates a new PDF parser
func NewPDFParser() *PDFParser {
	return &PDFParser{}
}

// ExtractText extracts text content from PDF
func (p *PDFParser) ExtractText(pdfReader io.Reader) (string, error) {
	// Create temporary file since pdfcpu requires file path
	// For now, return simple error - in production, handle with temp file
	return "", fmt.Errorf("PDF extraction requires file handling - implement in infrastructure layer")
}

// ParseProblems extracts problems from PDF content
func (p *PDFParser) ParseProblems(content string) ([]ParsedProblem, error) {
	var problems []ParsedProblem

	// Split by problem markers (e.g., "Problem 1:", "Problem #2", "## Problem")
	problemSections := p.splitByProblem(content)

	for idx, section := range problemSections {
		problem := ParsedProblem{
			ProblemNumber: idx + 1,
		}

		// Extract title
		problem.Title = p.extractTitle(section)

		// Extract description
		problem.Description = p.extractDescription(section)

		// Extract difficulty
		problem.Difficulty = p.extractDifficulty(section)

		// Extract schema
		problem.SchemaSQL = p.extractSQL(section, "schema|create table|init")

		// Extract solution (if exists)
		solution := p.extractSQL(section, "solution|answer|sql query")
		if solution != "" {
			problem.SolutionSQL = &solution
		}

		// Extract test cases
		problem.TestCases = p.extractTestCases(section)

		problems = append(problems, problem)
	}

	return problems, nil
}

// ParsedProblem represents a parsed problem from PDF
type ParsedProblem struct {
	ProblemNumber int
	Title         string
	Description   string
	Difficulty    string
	SchemaSQL     string
	SolutionSQL   *string
	TestCases     []ParsedTestCase
}

// ParsedTestCase represents a parsed test case
type ParsedTestCase struct {
	TestNumber     int
	Description    string
	TestDataSQL    string
	ExpectedOutput json.RawMessage
	IsPublic       bool
}

// Helper functions

func (p *PDFParser) splitByProblem(content string) []string {
	// Match patterns like "Problem 1:", "Problem #2", "## Problem"
	re := regexp.MustCompile(`(?i)(problem\s+[#\d]+|##\s*problem)`)
	parts := re.Split(content, -1)

	// Filter out empty parts
	var sections []string
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			sections = append(sections, part)
		}
	}

	return sections
}

func (p *PDFParser) extractTitle(section string) string {
	lines := strings.Split(section, "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return "Untitled Problem"
}

func (p *PDFParser) extractDescription(section string) string {
	// Find content between title and schema/solution
	re := regexp.MustCompile(`(?i)schema|solution|create table|test cases`)
	parts := re.Split(section, 2)
	if len(parts) > 0 {
		desc := strings.TrimSpace(parts[0])
		// Remove problem number if present
		desc = regexp.MustCompile(`^\d+\.\s*`).ReplaceAllString(desc, "")
		return desc
	}
	return ""
}

func (p *PDFParser) extractDifficulty(section string) string {
	section = strings.ToLower(section)

	if strings.Contains(section, "easy") {
		return "easy"
	}
	if strings.Contains(section, "hard") {
		return "hard"
	}
	if strings.Contains(section, "medium") {
		return "medium"
	}

	// Default
	return "medium"
}

func (p *PDFParser) extractSQL(section string, pattern string) string {
	// Find section with pattern (schema, solution, etc.)
	re := regexp.MustCompile(fmt.Sprintf(`(?i)%s:?\s*([\s\S]*?)(?:test|input|expected|$)`, pattern))
	matches := re.FindStringSubmatch(section)

	if len(matches) > 1 {
		sql := strings.TrimSpace(matches[1])
		// Clean up SQL - remove markdown code blocks
		sql = strings.ReplaceAll(sql, "```sql", "")
		sql = strings.ReplaceAll(sql, "```", "")
		sql = strings.TrimSpace(sql)
		return sql
	}

	return ""
}

func (p *PDFParser) extractTestCases(section string) []ParsedTestCase {
	var testCases []ParsedTestCase

	// Simple pattern matching for test cases
	// Format: "Test 1: ..." or "Input: ... Expected: ..."
	lines := strings.Split(section, "\n")

	var currentTest ParsedTestCase
	testNumber := 1

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		// Detect test case start
		if regexp.MustCompile(`(?i)test\s+\d+|test case`).MatchString(line) {
			if currentTest.TestDataSQL != "" {
				testCases = append(testCases, currentTest)
			}
			currentTest = ParsedTestCase{
				TestNumber: testNumber,
				IsPublic:   true,
			}
			testNumber++
		}

		// Extract input/data
		if regexp.MustCompile(`(?i)input:|test data:|insert`).MatchString(line) {
			currentTest.TestDataSQL = line
		}

		// Extract expected output
		if regexp.MustCompile(`(?i)expected:|output:`).MatchString(line) {
			currentTest.Description = line
		}
	}

	// Add last test case if exists
	if currentTest.TestDataSQL != "" {
		testCases = append(testCases, currentTest)
	}

	return testCases
}
