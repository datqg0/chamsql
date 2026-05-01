package ai

import "context"

// IChatLLMClient là interface cho chatbot AI client
type IChatLLMClient interface {
	Chat(ctx context.Context, messages []ChatMessage, tools []ChatTool) (*ChatCompletionResponse, error)
}

// Đảm bảo OpenAIChatClient implement interface này
var _ IChatLLMClient = (*OpenAIChatClient)(nil)
