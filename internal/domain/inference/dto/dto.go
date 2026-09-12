package dto

type ChatMessageDTO struct {
	Role    string `json:"role" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type ChatCompletionRequest struct {
	Model       string           `json:"model" binding:"required"`
	Messages    []ChatMessageDTO `json:"messages" binding:"required,min=1"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	Temperature float32          `json:"temperature,omitempty"`
	Stream      bool             `json:"stream,omitempty"`
}

type ChoiceMessageDTO struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChoiceDTO struct {
	Index        int              `json:"index"`
	Message      ChoiceMessageDTO `json:"message"`
	FinishReason string           `json:"finish_reason"`
}

type UsageDTO struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatCompletionResponse struct {
	ID      string      `json:"id"`
	Object  string      `json:"object"`
	Created int64       `json:"created"`
	Model   string      `json:"model"`
	Choices []ChoiceDTO `json:"choices"`
	Usage   UsageDTO    `json:"usage"`
}

type DeltaDTO struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

type StreamChoiceDTO struct {
	Index        int      `json:"index"`
	Delta        DeltaDTO `json:"delta"`
	FinishReason *string  `json:"finish_reason"`
}

type ChatCompletionChunk struct {
	ID      string            `json:"id"`
	Object  string            `json:"object"`
	Created int64             `json:"created"`
	Model   string            `json:"model"`
	Choices []StreamChoiceDTO `json:"choices"`
	Usage   *UsageDTO         `json:"usage,omitempty"`
}
