package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	aiUsecase "backend/internals/ai/usecase"
	pdfDomain "backend/internals/pdf/domain"
	"backend/internals/pdf/repository"
	"backend/pkgs/pdf"
)

// IUploadManager handles PDF upload workflow
type IUploadManager interface {
	HandleUpload(ctx context.Context, lecturerID int64, filePath, fileName, originalFilename string) (*pdfDomain.PDFUpload, error)
	ProcessExtraction(ctx context.Context, pdfUploadID int64) error
	GenerateAIContent(ctx context.Context, pdfUploadID int64) error
	GetUploadStatus(ctx context.Context, uploadID int64) (*pdfDomain.PDFUpload, error)
}

// uploadManager implements IUploadManager
type uploadManager struct {
	pdfRepo           repository.IPDFRepository
	pdfParser         *pdf.PDFParser
	solutionGenerator aiUsecase.IAISolutionGenerator
	testCaseGenerator aiUsecase.IAITestCaseGenerator
	testCaseValidator aiUsecase.IAITestCaseValidator
	aiOrchestrator    aiUsecase.IAIOrchestrator
}

// NewUploadManager creates a new upload manager
func NewUploadManager(
	pdfRepo repository.IPDFRepository,
	pdfParser *pdf.PDFParser,
	solutionGenerator aiUsecase.IAISolutionGenerator,
	testCaseGenerator aiUsecase.IAITestCaseGenerator,
	testCaseValidator aiUsecase.IAITestCaseValidator,
	aiOrchestrator aiUsecase.IAIOrchestrator,
) IUploadManager {
	return &uploadManager{
		pdfRepo:           pdfRepo,
		pdfParser:         pdfParser,
		solutionGenerator: solutionGenerator,
		testCaseGenerator: testCaseGenerator,
		testCaseValidator: testCaseValidator,
		aiOrchestrator:    aiOrchestrator,
	}
}

// HandleUpload creates a new PDF upload record
func (m *uploadManager) HandleUpload(ctx context.Context, lecturerID int64, filePath, fileName, originalFilename string) (*pdfDomain.PDFUpload, error) {
	upload, err := m.pdfRepo.CreatePDFUpload(ctx, lecturerID, filePath, fileName, originalFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF upload: %w", err)
	}

	return upload, nil
}

// ProcessExtraction parses PDF and creates problem review queue entries
func (m *uploadManager) ProcessExtraction(ctx context.Context, pdfUploadID int64) error {
	// Step 1: Update status to parsing
	_, err := m.pdfRepo.UpdatePDFUploadStatus(ctx, pdfUploadID, "parsing")
	if err != nil {
		return fmt.Errorf("failed to update status to parsing: %w", err)
	}

	// Step 2: Get PDF upload to find file path
	upload, err := m.pdfRepo.GetPDFUploadByID(ctx, pdfUploadID)
	if err != nil {
		return fmt.Errorf("failed to get PDF upload: %w", err)
	}

	// Step 3: Parse PDF - extract text first, then parse problems
	// NOTE: In production, this should download from MinIO first
	problems, err := m.pdfParser.ParseProblems(upload.FilePath)
	if err != nil {
		_, _ = m.pdfRepo.UpdatePDFUploadError(ctx, pdfUploadID, fmt.Sprintf("PDF parsing failed: %v", err))
		return fmt.Errorf("failed to parse PDF: %w", err)
	}

	// Step 4: Create extraction result structure
	extractionResult := map[string]interface{}{
		"total_problems": len(problems),
		"problems":       problems,
	}
	extractionBytes, err := json.Marshal(extractionResult)
	if err != nil {
		return fmt.Errorf("failed to marshal extraction result: %w", err)
	}

	// Step 5: Update PDF upload with extraction result
	_, err = m.pdfRepo.UpdatePDFUploadWithExtraction(ctx, pdfUploadID, "generating", extractionBytes)
	if err != nil {
		return fmt.Errorf("failed to update PDF upload with extraction: %w", err)
	}

	// Step 6: Create problem review queue entries for each problem
	for i, problem := range problems {
		// Convert ParsedTestCase to TestCaseData
		testCases := make([]pdfDomain.TestCaseData, len(problem.TestCases))
		for j, tc := range problem.TestCases {
			testCases[j] = pdfDomain.TestCaseData{
				TestNumber:     tc.TestNumber,
				Description:    tc.Description,
				TestDataSQL:    tc.TestDataSQL,
				ExpectedOutput: tc.ExpectedOutput,
				IsPublic:       tc.IsPublic,
			}
		}

		problemDraft := &pdfDomain.ProblemDraft{
			Title:       problem.Title,
			Description: problem.Description,
			Difficulty:  problem.Difficulty,
			InitScript:  problem.SchemaSQL,
			TestCases:   testCases,
		}

		draftBytes, err := json.Marshal(problemDraft)
		if err != nil {
			return fmt.Errorf("failed to marshal problem draft: %w", err)
		}

		_, err = m.pdfRepo.CreateProblemReviewQueue(ctx, pdfUploadID, i+1, draftBytes)
		if err != nil {
			return fmt.Errorf("failed to create problem review queue for problem %d: %w", i+1, err)
		}
	}

	return nil
}

// GenerateAIContent generates AI content for all extracted problems
func (m *uploadManager) GenerateAIContent(ctx context.Context, pdfUploadID int64) error {
	// Get all pending problems for this PDF upload
	problems, err := m.pdfRepo.GetProblemReviewQueueByPDF(ctx, pdfUploadID)
	if err != nil {
		return fmt.Errorf("failed to get problems for PDF: %w", err)
	}

	// For each problem, generate AI content
	for _, problem := range problems {
		if problem.Status != "pending" {
			continue // Skip non-pending problems
		}

		// Parse the problem draft
		var draft pdfDomain.ProblemDraft
		err := json.Unmarshal(problem.ProblemDraft, &draft)
		if err != nil {
			return fmt.Errorf("failed to parse problem draft: %w", err)
		}

		// Generate complete problem using orchestrator
		completeProblem, err := m.aiOrchestrator.GenerateCompleteProblem(ctx, draft.Description, draft.InitScript)
		if err != nil {
			return fmt.Errorf("failed to generate AI content for problem %d: %w", problem.ProblemNumber, err)
		}

		// Update problem draft with AI-generated solution
		if completeProblem.SolutionResponse != nil {
			draft.SolutionQuery = completeProblem.SolutionResponse.GeneratedContent
		}

		// Update test cases if generated
		if len(completeProblem.TestCases) > 0 {
			draft.TestCases = make([]pdfDomain.TestCaseData, len(completeProblem.TestCases))
			for i, tc := range completeProblem.TestCases {
				draft.TestCases[i] = pdfDomain.TestCaseData{
					TestNumber:     tc.TestNumber,
					Description:    tc.Description,
					TestDataSQL:    tc.TestDataSQL,
					ExpectedOutput: tc.ExpectedOutput,
					IsPublic:       tc.IsPublic,
				}
			}
		}

		updatedDraft, err := json.Marshal(draft)
		if err != nil {
			return fmt.Errorf("failed to marshal updated draft: %w", err)
		}

		_, err = m.pdfRepo.UpdateProblemReviewDraft(ctx, problem.ID, updatedDraft, nil)
		if err != nil {
			return fmt.Errorf("failed to update problem draft: %w", err)
		}
	}

	// Update PDF upload status to completed
	_, err = m.pdfRepo.UpdatePDFUploadStatus(ctx, pdfUploadID, "completed")
	if err != nil {
		return fmt.Errorf("failed to update PDF upload status: %w", err)
	}

	return nil
}

// GetUploadStatus retrieves the status of a PDF upload
func (m *uploadManager) GetUploadStatus(ctx context.Context, uploadID int64) (*pdfDomain.PDFUpload, error) {
	upload, err := m.pdfRepo.GetPDFUploadByID(ctx, uploadID)
	if err != nil {
		return nil, fmt.Errorf("failed to get upload status: %w", err)
	}

	return upload, nil
}
