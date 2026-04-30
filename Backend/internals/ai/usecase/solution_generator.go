package usecase

import (
	"context"
	"fmt"

	"backend/internals/ai/domain"
	"backend/pkgs/ai"
)

// IAISolutionGenerator defines solution generation service
type IAISolutionGenerator interface {
	GenerateSolution(ctx context.Context, req domain.SolutionGenerationInput) (*domain.AIGenerationResponse, error)
}

// AISolutionGenerator implements solution generation
type aiSolutionGenerator struct {
	patternMatcher *ai.PatternMatcher
	llmClient      ai.LLMClient
	provider       string
	hybridMode     bool
}

// NewAISolutionGenerator creates a new solution generator
func NewAISolutionGenerator(
	patternMatcher *ai.PatternMatcher,
	llmClient ai.LLMClient,
	provider string,
) IAISolutionGenerator {
	return &aiSolutionGenerator{
		patternMatcher: patternMatcher,
		llmClient:      llmClient,
		provider:       provider,
		hybridMode:     true,
	}
}

// GenerateSolution generates SQL solution from description
func (g *aiSolutionGenerator) GenerateSolution(ctx context.Context, req domain.SolutionGenerationInput) (*domain.AIGenerationResponse, error) {
	resp := &domain.AIGenerationResponse{}

	// Step 1: Try pattern matching first (fast, deterministic)
	patternSQL, confidence := g.patternMatcher.MatchPattern(req.ProblemDescription, req.SchemaSQL)

	if confidence >= 0.75 {
		// Confidence high enough to use pattern match
		resp.GeneratedContent = patternSQL
		resp.ConfidenceScore = confidence
		resp.AIProvider = "pattern"
		return resp, nil
	}

	// Step 2: If pattern confidence is low, try LLM (OpenAI or HuggingFace)
	if g.llmClient != nil && req.SchemaSQL != "" {
		llmResult, err := g.llmClient.GenerateSolution(ctx, req.ProblemDescription, req.SchemaSQL)
		if err == nil && llmResult != "" {
			// Validate LLM result
			if g.patternMatcher.ValidateSQLSyntax(llmResult) {
				resp.GeneratedContent = llmResult
				resp.ConfidenceScore = 0.85
				resp.AIProvider = g.provider
				return resp, nil
			}
		}
	}

	// Step 3: Fallback to pattern match if available
	if patternSQL != "" && confidence > 0 {
		resp.GeneratedContent = patternSQL
		resp.ConfidenceScore = confidence
		resp.AIProvider = "pattern"
		return resp, nil
	}

	// Step 4: No solution generated
	resp.Error = "could not generate solution from description"
	return resp, fmt.Errorf("failed to generate solution")
}
