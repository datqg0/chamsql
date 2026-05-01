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

type ChatMessage struct {
    Role       string     `json:"role"`
    Content    string     `json:"content"`
    ToolCallID string     `json:"tool_call_id,omitempty"`
    ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
    Name       string     `json:"name,omitempty"`
}

type ToolCall struct {
    ID       string       `json:"id"`
    Type     string       `json:"type"`
    Function FunctionCall `json:"function"`
}

type FunctionCall struct {
    Name      string `json:"name"`
    Arguments string `json:"arguments"`
}

type ChatTool struct {
    Type     string           `json:"type"`
    Function ChatToolFunction `json:"function"`
}

type ChatToolFunction struct {
    Name        string         `json:"name"`
    Description string         `json:"description"`
    Parameters  ToolParameters `json:"parameters"`
}

type ToolParameters struct {
    Type       string              `json:"type"`
    Properties map[string]Property `json:"properties"`
    Required   []string            `json:"required,omitempty"`
}

type Property struct {
    Type        string   `json:"type"`
    Description string   `json:"description"`
    Enum        []string `json:"enum,omitempty"`
}

type chatCompletionRequest struct {
    Model       string        `json:"model"`
    Messages    []ChatMessage `json:"messages"`
    Tools       []ChatTool    `json:"tools,omitempty"`
    ToolChoice  string        `json:"tool_choice,omitempty"`
    MaxTokens   int           `json:"max_tokens,omitempty"`
    Temperature float64       `json:"temperature,omitempty"`
}

type ChatCompletionResponse struct {
    Choices []struct {
        Message      ChatMessage `json:"message"`
        FinishReason string      `json:"finish_reason"`
    } `json:"choices"`
}

type OpenAIChatClient struct {
    apiKey     string
    baseURL    string
    model      string
    httpClient *http.Client
}

func NewOpenAIChatClient(apiKey, baseURL, model string) *OpenAIChatClient {
    if baseURL == "" {
        baseURL = "https://api.openai.com/v1"
    }
    if model == "" {
        model = "gpt-4o-mini"
    }
    return &OpenAIChatClient{
        apiKey:  apiKey,
        baseURL: baseURL,
        model:   model,
        httpClient: &http.Client{Timeout: 60 * time.Second},
    }
}

func (c *OpenAIChatClient) Chat(ctx context.Context, messages []ChatMessage, tools []ChatTool) (*ChatCompletionResponse, error) {
    reqBody := chatCompletionRequest{
        Model:       c.model,
        Messages:    messages,
        MaxTokens:   1024,
        Temperature: 0.7,
    }
    if len(tools) > 0 {
        reqBody.Tools = tools
        reqBody.ToolChoice = "auto"
    }

    body, err := json.Marshal(reqBody)
    if err != nil {
        return nil, err
    }

    httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
        c.baseURL+"/chat/completions", bytes.NewReader(body))
    if err != nil {
        return nil, err
    }
    httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        b, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("OpenAI API error %d: %s", resp.StatusCode, string(b))
    }

    var result ChatCompletionResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return &result, nil
}
