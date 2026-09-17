package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Faithful001/aegis/internal/domain/inference"
	"github.com/Faithful001/aegis/internal/domain/inference/dto"
)

type AnthropicAdapter struct {
	httpClient *http.Client
}

func NewAnthropicAdapter(httpClient ...*http.Client) *AnthropicAdapter {
	client := &http.Client{Timeout: 60 * time.Second}
	if len(httpClient) > 0 && httpClient[0] != nil {
		client = httpClient[0]
	}
	return &AnthropicAdapter{httpClient: client}
}

func (a *AnthropicAdapter) getEndpoint(baseURL string) string {
	if baseURL != "" {
		return strings.TrimRight(baseURL, "/") + "/messages"
	}
	return "https://api.anthropic.com/v1/messages"
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	Messages  []anthropicMessage `json:"messages"`
	System    string             `json:"system,omitempty"`
	MaxTokens int                `json:"max_tokens"`
	Stream    bool               `json:"stream,omitempty"`
}

type anthropicContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type anthropicResponse struct {
	ID         string                  `json:"id"`
	Type       string                  `json:"type"`
	Role       string                  `json:"role"`
	Content    []anthropicContentBlock `json:"content"`
	Model      string                  `json:"model"`
	StopReason string                  `json:"stop_reason"`
	Usage      anthropicUsage          `json:"usage"`
}

func (a *AnthropicAdapter) formatRequest(req dto.ChatCompletionRequest) (anthropicRequest, error) {
	var systemPrompt string
	var messages []anthropicMessage

	for _, m := range req.Messages {
		if m.Role == "system" {
			if systemPrompt != "" {
				systemPrompt += "\n" + m.Content
			} else {
				systemPrompt = m.Content
			}
		} else {
			role := m.Role
			if role != "user" && role != "assistant" {
				role = "user"
			}
			messages = append(messages, anthropicMessage{
				Role:    role,
				Content: m.Content,
			})
		}
	}

	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 1024
	}

	return anthropicRequest{
		Model:     req.Model,
		Messages:  messages,
		System:    systemPrompt,
		MaxTokens: maxTokens,
		Stream:    req.Stream,
	}, nil
}

func (a *AnthropicAdapter) Generate(
	ctx context.Context,
	apiKey, baseURL string,
	req dto.ChatCompletionRequest,
) (*dto.ChatCompletionResponse, error) {
	antReq, err := a.formatRequest(req)
	if err != nil {
		return nil, err
	}
	antReq.Stream = false

	bodyBytes, err := json.Marshal(antReq)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.getEndpoint(baseURL), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProviderAPIError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: HTTP %d - %s", ErrProviderAPIError, resp.StatusCode, string(respBody))
	}

	var antResp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&antResp); err != nil {
		return nil, fmt.Errorf("failed to decode Anthropic response: %w", err)
	}

	var text string
	for _, c := range antResp.Content {
		if c.Type == "text" {
			text += c.Text
		}
	}

	finishReason := antResp.StopReason
	if finishReason == "end_turn" {
		finishReason = "stop"
	}

	return &dto.ChatCompletionResponse{
		ID:      antResp.ID,
		Object:  "chat.completion",
		Created: time.Now().UTC().Unix(),
		Model:   antResp.Model,
		Choices: []dto.ChoiceDTO{
			{
				Index: 0,
				Message: dto.ChoiceMessageDTO{
					Role:    "assistant",
					Content: text,
				},
				FinishReason: finishReason,
			},
		},
		Usage: dto.UsageDTO{
			PromptTokens:     antResp.Usage.InputTokens,
			CompletionTokens: antResp.Usage.OutputTokens,
			TotalTokens:      antResp.Usage.InputTokens + antResp.Usage.OutputTokens,
		},
	}, nil
}

func (a *AnthropicAdapter) StreamGenerate(
	ctx context.Context,
	apiKey, baseURL string,
	req dto.ChatCompletionRequest,
) (<-chan inference.StreamChunk, error) {
	antReq, err := a.formatRequest(req)
	if err != nil {
		return nil, err
	}
	antReq.Stream = true

	bodyBytes, err := json.Marshal(antReq)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.getEndpoint(baseURL), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProviderAPIError, err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("%w: HTTP %d - %s", ErrProviderAPIError, resp.StatusCode, string(respBody))
	}

	outChan := make(chan inference.StreamChunk)

	go func() {
		defer resp.Body.Close()
		defer close(outChan)

		scanner := bufio.NewScanner(resp.Body)
		var promptTokens, completionTokens int

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, ":") {
				continue
			}

			if strings.HasPrefix(line, "data: ") {
				dataStr := strings.TrimPrefix(line, "data: ")
				var event map[string]interface{}
				if err := json.Unmarshal([]byte(dataStr), &event); err != nil {
					continue
				}

				eventType, _ := event["type"].(string)
				switch eventType {
				case "message_start":
					if msg, ok := event["message"].(map[string]interface{}); ok {
						if usage, ok := msg["usage"].(map[string]interface{}); ok {
							if pt, ok := usage["input_tokens"].(float64); ok {
								promptTokens = int(pt)
							}
						}
					}
				case "content_block_delta":
					if delta, ok := event["delta"].(map[string]interface{}); ok {
						if text, ok := delta["text"].(string); ok {
							completionTokens++
							outChan <- inference.StreamChunk{
								Content:      text,
								Done:         false,
								PromptTokens: promptTokens,
								OutputTokens: completionTokens,
								TotalTokens:  promptTokens + completionTokens,
							}
						}
					}
				case "message_delta":
					if usage, ok := event["usage"].(map[string]interface{}); ok {
						if ot, ok := usage["output_tokens"].(float64); ok {
							completionTokens = int(ot)
						}
					}
				case "message_stop":
					outChan <- inference.StreamChunk{
						Done:         true,
						PromptTokens: promptTokens,
						OutputTokens: completionTokens,
						TotalTokens:  promptTokens + completionTokens,
					}
					return
				}
			}
		}
	}()

	return outChan, nil
}
