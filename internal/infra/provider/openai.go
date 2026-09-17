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

type OpenAIAdapter struct {
	httpClient *http.Client
}

func NewOpenAIAdapter(httpClient ...*http.Client) *OpenAIAdapter {
	client := &http.Client{Timeout: 60 * time.Second}
	if len(httpClient) > 0 && httpClient[0] != nil {
		client = httpClient[0]
	}
	return &OpenAIAdapter{httpClient: client}
}

func (a *OpenAIAdapter) getEndpoint(baseURL string) string {
	if baseURL != "" {
		return strings.TrimRight(baseURL, "/") + "/chat/completions"
	}
	return "https://api.openai.com/v1/chat/completions"
}

func (a *OpenAIAdapter) Generate(
	ctx context.Context,
	apiKey, baseURL string,
	req dto.ChatCompletionRequest,
) (*dto.ChatCompletionResponse, error) {
	req.Stream = false
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.getEndpoint(baseURL), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProviderAPIError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: HTTP %d - %s", ErrProviderAPIError, resp.StatusCode, string(respBody))
	}

	var completion dto.ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&completion); err != nil {
		return nil, fmt.Errorf("failed to decode OpenAI response: %w", err)
	}

	return &completion, nil
}

func (a *OpenAIAdapter) StreamGenerate(
	ctx context.Context,
	apiKey, baseURL string,
	req dto.ChatCompletionRequest,
) (<-chan inference.StreamChunk, error) {
	req.Stream = true
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.getEndpoint(baseURL), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

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
		var totalPromptTokens, totalCompletionTokens int

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, ":") {
				continue
			}

			if strings.HasPrefix(line, "data: ") {
				dataStr := strings.TrimPrefix(line, "data: ")
				if dataStr == "[DONE]" {
					outChan <- inference.StreamChunk{
						Done:         true,
						PromptTokens: totalPromptTokens,
						OutputTokens: totalCompletionTokens,
						TotalTokens:  totalPromptTokens + totalCompletionTokens,
					}
					return
				}

				var raw map[string]interface{}
				if err := json.Unmarshal([]byte(dataStr), &raw); err != nil {
					continue
				}

				// Extract delta content
				var text string
				if choices, ok := raw["choices"].([]interface{}); ok && len(choices) > 0 {
					if choice, ok := choices[0].(map[string]interface{}); ok {
						if delta, ok := choice["delta"].(map[string]interface{}); ok {
							if content, ok := delta["content"].(string); ok {
								text = content
								totalCompletionTokens++
							}
						}
					}
				}

				// Check usage if returned by OpenAI
				if usage, ok := raw["usage"].(map[string]interface{}); ok {
					if pt, ok := usage["prompt_tokens"].(float64); ok {
						totalPromptTokens = int(pt)
					}
					if ct, ok := usage["completion_tokens"].(float64); ok {
						totalCompletionTokens = int(ct)
					}
				}

				outChan <- inference.StreamChunk{
					Content:      text,
					Done:         false,
					PromptTokens: totalPromptTokens,
					OutputTokens: totalCompletionTokens,
					TotalTokens:  totalPromptTokens + totalCompletionTokens,
				}
			}
		}

		if err := scanner.Err(); err != nil {
			outChan <- inference.StreamChunk{
				Error: err.Error(),
				Done:  true,
			}
		}
	}()

	return outChan, nil
}
