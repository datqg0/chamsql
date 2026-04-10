package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HuggingFaceConfig holds HuggingFace API configuration
type HuggingFaceConfig struct {
	APIKey  string
	Timeout time.Duration
}

// HuggingFaceClient wraps HuggingFace API calls
type HuggingFaceClient struct {
	apiKey  string
	timeout time.Duration
	client  *http.Client
}

// HFRequest represents a request to HuggingFace API
type HFRequest struct {
	Inputs string `json:"inputs"`
}

// HFResponse represents response from HuggingFace API
type HFResponse struct {
	GeneratedText string `json:"generated_text,omitempty"`
	Error         string `json:"error,omitempty"`
}

// NewHuggingFaceClient creates a new HuggingFace client
func NewHuggingFaceClient(config HuggingFaceConfig) *HuggingFaceClient {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &HuggingFaceClient{
		apiKey:  config.APIKey,
		timeout: config.Timeout,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// GenerateSolution calls HuggingFace to generate SQL solution
func (c *HuggingFaceClient) GenerateSolution(ctx context.Context, description string, schemaSQL string) (string, error) {
	prompt := fmt.Sprintf(`Given a SQL schema and a problem description, generate the SQL query to solve the problem.

Schema:
%s

Problem:
%s

SQL Query:`, schemaSQL, description)

	return c.callHuggingFace(ctx, prompt)
}

// GenerateTestCase generates test case descriptions
func (c *HuggingFaceClient) GenerateTestCase(ctx context.Context, description string, schemaSQL string, solutionSQL string) (string, error) {
	prompt := fmt.Sprintf(`Given a SQL problem schema and solution, generate test case data for it.

Schema:
%s

Solution:
%s

Generate INSERT statements and describe what this test case tests:`, schemaSQL, solutionSQL)

	return c.callHuggingFace(ctx, prompt)
}

// CallHuggingFace makes actual API call to HuggingFace
func (c *HuggingFaceClient) callHuggingFace(ctx context.Context, prompt string) (string, error) {
	url := "https://api-inference.huggingface.co/models/defog/sqlcoder"

	reqBody := HFRequest{
		Inputs: prompt,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call HuggingFace API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HuggingFace API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var hfResp []HFResponse
	if err := json.Unmarshal(body, &hfResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(hfResp) == 0 {
		return "", fmt.Errorf("empty response from HuggingFace")
	}

	if hfResp[0].Error != "" {
		return "", fmt.Errorf("HuggingFace error: %s", hfResp[0].Error)
	}

	return hfResp[0].GeneratedText, nil
}
