package pdf

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	ledpdf "github.com/ledongthuc/pdf"
)

// PDFParser handles PDF file parsing
type PDFParser struct{}

// NewPDFParser creates a new PDF parser
func NewPDFParser() *PDFParser {
	return &PDFParser{}
}

// ExtractTextFromFile reads a PDF from disk and returns all plain text.
// Supports Vietnamese unicode (UTF-8) as returned by ledongthuc/pdf.
func (p *PDFParser) ExtractTextFromFile(filePath string) (string, error) {
	f, r, err := ledpdf.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("cannot open PDF file: %w", err)
	}
	defer f.Close()

	var sb strings.Builder
	for i := 1; i <= r.NumPage(); i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			// Bỏ qua trang lỗi, không dừng toàn bộ quá trình
			continue
		}
		sb.WriteString(text)
		sb.WriteString("\n")
	}

	result := sb.String()
	if strings.TrimSpace(result) == "" {
		return "", fmt.Errorf("PDF contains no extractable text (may be image-only or encrypted)")
	}
	return result, nil
}

// ParseProblems extracts problems from already-extracted PDF text content.
// Input: raw text string from ExtractTextFromFile
func (p *PDFParser) ParseProblems(content string) ([]ParsedProblem, error) {
	var problems []ParsedProblem

	// Split by problem markers (e.g., "Problem 1:", "Problem #2", "## Problem", "Bài 1:", "Câu 1:")
	problemSections := p.splitByProblem(content)

	if len(problemSections) == 0 {
		// Treat entire content as one problem
		problemSections = []string{content}
	}

	for idx, section := range problemSections {
		section = strings.TrimSpace(section)
		if section == "" {
			continue
		}

		problem := ParsedProblem{
			ProblemNumber: idx + 1,
		}

		// Extract title
		problem.Title = p.extractTitle(section)

		// Extract description
		problem.Description = p.extractDescription(section)

		// Extract difficulty
		problem.Difficulty = p.extractDifficulty(section)

		// Extract schema SQL
		problem.SchemaSQL = p.extractSQL(section, "schema|create table|init|khởi tạo|tạo bảng")

		// Extract solution if explicitly present
		solution := p.extractSQL(section, "solution|answer|sql query|đáp án|lời giải")
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
	// Match Vietnamese and English problem markers (Problem, Bài, Câu, Question, Section, ...)
	// Hỗ trợ cả các biến thể có dấu, không dấu, số thứ tự, và ký hiệu ##
	re := regexp.MustCompile(`(?im)(^|\n)(problem\s*[#\d]*|##\s*problem|bài\s*\d+|câu\s*\d+|question\s*\d+|đề\s+bài\s*\d+)`)
	indices := re.FindAllStringIndex(content, -1)

	if len(indices) == 0 {
		return nil
	}

	var sections []string
	for i, idx := range indices {
		start := idx[0]
		var end int
		if i+1 < len(indices) {
			end = indices[i+1][0]
		} else {
			end = len(content)
		}
		section := strings.TrimSpace(content[start:end])
		if section != "" {
			sections = append(sections, section)
		}
	}
	return sections
}

func (p *PDFParser) extractTitle(section string) string {
	lines := strings.Split(section, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			// Remove problem number prefix if present (e.g., "Bài 1: Title")
			trimmed = regexp.MustCompile(`(?i)^(problem|bài|câu|question)\s+\d+[:\.\)]\s*`).ReplaceAllString(trimmed, "")
			if trimmed != "" {
				return trimmed
			}
		}
	}
	return fmt.Sprintf("Problem %d", 1)
}

func (p *PDFParser) extractDescription(section string) string {
	// Find content before schema/solution markers
	re := regexp.MustCompile(`(?i)(schema|solution|create table|test cases?|đáp án|tạo bảng|input|output)`)
	parts := re.Split(section, 2)
	if len(parts) > 0 {
		desc := strings.TrimSpace(parts[0])
		// Remove problem number header line
		lines := strings.Split(desc, "\n")
		if len(lines) > 1 {
			desc = strings.TrimSpace(strings.Join(lines[1:], "\n"))
		}
		return desc
	}
	return ""
}

func (p *PDFParser) extractDifficulty(section string) string {
	lower := strings.ToLower(section)

	if strings.Contains(lower, "easy") || strings.Contains(lower, "dễ") {
		return "easy"
	}
	if strings.Contains(lower, "hard") || strings.Contains(lower, "khó") {
		return "hard"
	}
	if strings.Contains(lower, "medium") || strings.Contains(lower, "trung bình") {
		return "medium"
	}
	return "medium"
}

func (p *PDFParser) extractSQL(section string, pattern string) string {
	// Try to find SQL in code blocks first
	codeBlockRe := regexp.MustCompile("(?is)```(?:sql)?\\s*(.+?)```")
	matches := codeBlockRe.FindStringSubmatch(section)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Find section with keyword pattern
	re := regexp.MustCompile(fmt.Sprintf(`(?i)(?:%s):?\s*([\s\S]*?)(?:test|input|expected|output|đáp án|$)`, pattern))
	kMatches := re.FindStringSubmatch(section)
	if len(kMatches) > 1 {
		sql := strings.TrimSpace(kMatches[1])
		sql = strings.ReplaceAll(sql, "```sql", "")
		sql = strings.ReplaceAll(sql, "```", "")
		return strings.TrimSpace(sql)
	}

	return ""
}

func (p *PDFParser) extractTestCases(section string) []ParsedTestCase {
	var testCases []ParsedTestCase

	lines := strings.Split(section, "\n")
	var currentTest ParsedTestCase
	testNumber := 1

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Detect test case start (Vietnamese + English)
		if regexp.MustCompile(`(?i)(test\s+\d+|test case\s*\d*|trường hợp\s+\d+|tc\s*\d+)`).MatchString(line) {
			if currentTest.TestDataSQL != "" {
				testCases = append(testCases, currentTest)
			}
			currentTest = ParsedTestCase{
				TestNumber: testNumber,
				IsPublic:   true,
			}
			testNumber++
		}

		// Extract SQL data
		if regexp.MustCompile(`(?i)(input:|test data:|insert\s+into|dữ liệu:)`).MatchString(line) {
			currentTest.TestDataSQL = line
		}

		// Extract expected output description
		if regexp.MustCompile(`(?i)(expected:|output:|kết quả:)`).MatchString(line) {
			currentTest.Description = line
		}
	}

	if currentTest.TestDataSQL != "" {
		testCases = append(testCases, currentTest)
	}

	return testCases
}
