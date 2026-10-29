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
	providerDTO "github.com/Faithful001/aegis/internal/domain/provider/dto"
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
		u := strings.TrimRight(baseURL, "/")
		if strings.HasSuffix(u, "/chat/completions") {
			return u
		}
		if strings.HasSuffix(u, "/v1") {
			return u + "/chat/completions"
		}
		return u + "/v1/chat/completions"
	}
	return "https://api.openai.com/v1/chat/completions"
}

type openAIMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float32         `json:"temperature,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

func formatOpenAIRequest(req dto.ChatCompletionRequest, stream bool) openAIRequest {
	oaiReq := openAIRequest{
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      stream,
		Messages:    make([]openAIMessage, len(req.Messages)),
	}

	for i, m := range req.Messages {
		var hasImages bool
		for _, att := range m.Attachments {
			if att.FileType == "image" && (att.Content != "" || att.URL != "") {
				hasImages = true
				break
			}
		}

		if hasImages {
			var parts []map[string]interface{}
			if m.Content != "" {
				parts = append(parts, map[string]interface{}{
					"type": "text",
					"text": m.Content,
				})
			}
			for _, att := range m.Attachments {
				if att.FileType == "image" {
					imgURL := att.URL
					if att.Content != "" {
						imgURL = fmt.Sprintf("data:%s;base64,%s", att.MimeType, att.Content)
					}
					if imgURL != "" {
						parts = append(parts, map[string]interface{}{
							"type": "image_url",
							"image_url": map[string]string{
								"url": imgURL,
							},
						})
					}
				}
			}
			oaiReq.Messages[i] = openAIMessage{
				Role:    m.Role,
				Content: parts,
			}
		} else {
			oaiReq.Messages[i] = openAIMessage{
				Role:    m.Role,
				Content: m.Content,
			}
		}
	}
	return oaiReq
}

func (a *OpenAIAdapter) Generate(
	ctx context.Context,
	apiKey, baseURL string,
	req dto.ChatCompletionRequest,
) (*dto.ChatCompletionResponse, error) {
	oaiReq := formatOpenAIRequest(req, false)
	bodyBytes, err := json.Marshal(oaiReq)
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
	oaiReq := formatOpenAIRequest(req, true)
	bodyBytes, err := json.Marshal(oaiReq)
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

func (a *OpenAIAdapter) getModelsEndpoint(baseURL string) string {
	if baseURL != "" {
		u := strings.TrimRight(baseURL, "/")
		if strings.HasSuffix(u, "/models") {
			return u
		}
		if strings.HasSuffix(u, "/v1") {
			return u + "/models"
		}
		return u + "/v1/models"
	}
	return "https://api.openai.com/v1/models"
}

type openAIModelItem struct {
	ID      string `json:"id"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

type openAIModelsResponse struct {
	Data []openAIModelItem `json:"data"`
}

func (a *OpenAIAdapter) ListModels(ctx context.Context, apiKey, baseURL string) ([]providerDTO.ModelDescriptor, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", a.getModelsEndpoint(baseURL), nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to query openai models api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai api error: HTTP %d - %s", resp.StatusCode, string(respBody))
	}

	var mResp openAIModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&mResp); err != nil {
		return nil, fmt.Errorf("failed to decode openai models response: %w", err)
	}

	results := make([]providerDTO.ModelDescriptor, 0, len(mResp.Data))
	for _, m := range mResp.Data {
		lowerID := strings.ToLower(m.ID)
		// Filter for chat completion models
		if !strings.HasPrefix(lowerID, "gpt-") &&
			!strings.HasPrefix(lowerID, "o1") &&
			!strings.HasPrefix(lowerID, "o3") &&
			!strings.HasPrefix(lowerID, "chatgpt-") {
			continue
		}
		// Filter out non-chat specialized variations
		if strings.Contains(lowerID, "realtime") ||
			strings.Contains(lowerID, "audio") ||
			strings.Contains(lowerID, "tts") ||
			strings.Contains(lowerID, "transcription") {
			continue
		}

		isFast := strings.Contains(lowerID, "mini") || strings.Contains(lowerID, "turbo") || strings.Contains(lowerID, "nano")
		desc := "OpenAI frontier model"
		if strings.HasPrefix(lowerID, "o1") || strings.HasPrefix(lowerID, "o3") {
			desc = "Reasoning model designed for complex science, math & code"
		} else if strings.Contains(lowerID, "4o-mini") {
			desc = "Fast, lightweight omni model for cost-effective tasks"
		} else if strings.Contains(lowerID, "4o") {
			desc = "Flagship omni model for multimodal reasoning & intelligence"
		}

		results = append(results, providerDTO.ModelDescriptor{
			ID:          m.ID,
			Name:        m.ID,
			Provider:    "openai",
			Description: desc,
			Fast:        isFast,
		})
	}

	return results, nil
}
