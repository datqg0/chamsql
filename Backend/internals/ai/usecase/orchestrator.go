package usecase

import (
	"context"
	"fmt"

	"backend/db"
	"backend/internals/ai/domain"
)

// IAIOrchestrator coordinates all AI services
type IAIOrchestrator interface {
	GenerateCompleteProblem(ctx context.Context, problemDesc, schemaSQL string) (*CompleteProblem, error)
	GenerateSolution(ctx context.Context, req domain.SolutionGenerationInput) (*domain.AIGenerationResponse, error)
	GenerateTestCases(ctx context.Context, req domain.TestCaseGenerationInput) ([]domain.TestCaseGenerated, error)
	ValidateTestCases(ctx context.Context, schemaSQL, solutionSQL string) (*domain.ValidationResult, error)
}

// CompleteProblem is the full problem with solution and test cases
type CompleteProblem struct {
	SolutionResponse *domain.AIGenerationResponse
	TestCases        []domain.TestCaseGenerated
	ValidationResult *domain.ValidationResult
}

// AIOrchestrator coordinates AI services
type aiOrchestrator struct {
	solutionGenerator IAISolutionGenerator
	testCaseGenerator IAITestCaseGenerator
	testCaseValidator IAITestCaseValidator
	database          *db.Database
}

// NewAIOrchestrator creates a new AI orchestrator
func NewAIOrchestrator(
	solutionGenerator IAISolutionGenerator,
	testCaseGenerator IAITestCaseGenerator,
	testCaseValidator IAITestCaseValidator,
	database *db.Database,
) IAIOrchestrator {
	return &aiOrchestrator{
		solutionGenerator: solutionGenerator,
		testCaseGenerator: testCaseGenerator,
		testCaseValidator: testCaseValidator,
		database:          database,
	}
}

// GenerateCompleteProblem orchestrates the full problem generation workflow
func (o *aiOrchestrator) GenerateCompleteProblem(ctx context.Context, problemDesc, schemaSQL string) (*CompleteProblem, error) {
	problem := &CompleteProblem{}

	// Step 1: Generate solution
	solResp, err := o.GenerateSolution(ctx, domain.SolutionGenerationInput{
		ProblemDescription: problemDesc,
		SchemaSQL:          schemaSQL,
	})
	if err != nil {
		return nil, fmt.Errorf("solution generation failed: %w", err)
	}
	problem.SolutionResponse = solResp

	// Step 2: Generate test cases
	testCases, err := o.GenerateTestCases(ctx, domain.TestCaseGenerationInput{
		SchemaSQL:           schemaSQL,
		SolutionSQL:         solResp.GeneratedContent,
		Description:         problemDesc,
		DifficultyLevel:     "medium",
		PublicTestCaseCount: 2,
		HiddenTestCaseCount: 6,
	})
	if err != nil {
		return nil, fmt.Errorf("test case generation failed: %w", err)
	}
	problem.TestCases = testCases

	// Step 3: Validate test cases
	validationTestCases := o.convertToValidationTestCases(testCases)
	valResult, err := o.testCaseValidator.ValidateTestCases(ctx, schemaSQL, solResp.GeneratedContent, validationTestCases)
	if err != nil {
		return nil, fmt.Errorf("test case validation failed: %w", err)
	}
	problem.ValidationResult = valResult

	return problem, nil
}

// GenerateSolution delegates to solution generator
func (o *aiOrchestrator) GenerateSolution(ctx context.Context, req domain.SolutionGenerationInput) (*domain.AIGenerationResponse, error) {
	resp, err := o.solutionGenerator.GenerateSolution(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate solution: %w", err)
	}
	return resp, nil
}

// GenerateTestCases delegates to test case generator
func (o *aiOrchestrator) GenerateTestCases(ctx context.Context, req domain.TestCaseGenerationInput) ([]domain.TestCaseGenerated, error) {
	return o.testCaseGenerator.GenerateTestCases(ctx, req)
}

// ValidateTestCases validates the test cases
func (o *aiOrchestrator) ValidateTestCases(ctx context.Context, schemaSQL, solutionSQL string) (*domain.ValidationResult, error) {
	// Note: In production, this would take actual test cases to validate
	// For now, this is a placeholder
	return &domain.ValidationResult{
		IsValid:     true,
		PassedCount: 8,
		TotalCount:  8,
	}, nil
}

// Helper methods

func (o *aiOrchestrator) convertToValidationTestCases(testCases []domain.TestCaseGenerated) []TestCaseForValidation {
	var validationTestCases []TestCaseForValidation

	for _, tc := range testCases {
		validationTestCases = append(validationTestCases, TestCaseForValidation{
			TestNumber:     tc.TestNumber,
			TestDataSQL:    tc.TestDataSQL,
			ExpectedOutput: tc.ExpectedOutput,
		})
	}

	return validationTestCases
}
