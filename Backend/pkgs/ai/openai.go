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

// OpenAIConfig holds OpenAI API configuration
type OpenAIConfig struct {
	APIKey  string
	Model   string
	Timeout time.Duration
}

// OpenAIClient wraps OpenAI API calls
type OpenAIClient struct {
	apiKey  string
	model   string
	timeout time.Duration
	client  *http.Client
}

// OAIRequest represents a request to OpenAI Chat Completion API
type OAIRequest struct {
	Model    string       `json:"model"`
	Messages []OAIMessage `json:"messages"`
}

// OAIMessage represents a message in OpenAI chat conversation
type OAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OAIResponse represents response from OpenAI API
type OAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(config OpenAIConfig) *OpenAIClient {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.Model == "" {
		config.Model = "gpt-4o"
	}

	return &OpenAIClient{
		apiKey:  config.APIKey,
		model:   config.Model,
		timeout: config.Timeout,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// GenerateSolution calls OpenAI to generate SQL solution
func (c *OpenAIClient) GenerateSolution(ctx context.Context, description string, schemaSQL string) (string, error) {
	prompt := fmt.Sprintf(`Given a SQL schema and a problem description, generate the SQL query to solve the problem.
Return ONLY the SQL query, no markdown blocks, no explanations.

Schema:
%s

Problem:
%s

SQL Query:`, schemaSQL, description)

	return c.callOpenAI(ctx, prompt)
}

// GenerateTestCase calls OpenAI to generate test case data
func (c *OpenAIClient) GenerateTestCase(ctx context.Context, description string, schemaSQL string, solutionSQL string) (string, error) {
	prompt := fmt.Sprintf(`Given a SQL problem schema and solution, generate test case data for it.
Return ONLY the SQL INSERT statements, no markdown blocks, no explanations.

Schema:
%s

Solution:
%s

Generate INSERT statements:`, schemaSQL, solutionSQL)

	return c.callOpenAI(ctx, prompt)
}

// callOpenAI makes actual API call to OpenAI
func (c *OpenAIClient) callOpenAI(ctx context.Context, prompt string) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"

	reqBody := OAIRequest{
		Model: c.model,
		Messages: []OAIMessage{
			{Role: "system", Content: "You are a SQL expert and database tutor."},
			{Role: "user", Content: prompt},
		},
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
		return "", fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var oaiResp OAIResponse
	if err := json.Unmarshal(body, &oaiResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if oaiResp.Error.Message != "" {
		return "", fmt.Errorf("OpenAI error: %s", oaiResp.Error.Message)
	}

	if len(oaiResp.Choices) == 0 {
		return "", fmt.Errorf("empty response from OpenAI")
	}

	return oaiResp.Choices[0].Message.Content, nil
}
