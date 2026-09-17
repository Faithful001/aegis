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

type GeminiAdapter struct {
	httpClient *http.Client
}

func NewGeminiAdapter(httpClient ...*http.Client) *GeminiAdapter {
	client := &http.Client{Timeout: 60 * time.Second}
	if len(httpClient) > 0 && httpClient[0] != nil {
		client = httpClient[0]
	}
	return &GeminiAdapter{httpClient: client}
}

func (a *GeminiAdapter) getEndpoint(baseURL, model, action string) string {
	if baseURL != "" {
		return fmt.Sprintf("%s/models/%s:%s", strings.TrimRight(baseURL, "/"), model, action)
	}
	return fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:%s", model, action)
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiGenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

type geminiRequest struct {
	Contents          []geminiContent        `json:"contents"`
	SystemInstruction *geminiContent         `json:"systemInstruction,omitempty"`
	GenerationConfig  geminiGenerationConfig `json:"generationConfig"`
}

type geminiCandidate struct {
	Content      geminiContent `json:"content"`
	FinishReason string        `json:"finishReason"`
}

type geminiUsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

type geminiResponse struct {
	Candidates    []geminiCandidate   `json:"candidates"`
	UsageMetadata geminiUsageMetadata `json:"usageMetadata"`
}

func (a *GeminiAdapter) formatRequest(req dto.ChatCompletionRequest) (geminiRequest, error) {
	var contents []geminiContent
	var systemContent *geminiContent

	for _, m := range req.Messages {
		if m.Role == "system" {
			systemContent = &geminiContent{
				Parts: []geminiPart{{Text: m.Content}},
			}
		} else {
			role := m.Role
			if role == "assistant" {
				role = "model"
			} else {
				role = "user"
			}
			contents = append(contents, geminiContent{
				Role:  role,
				Parts: []geminiPart{{Text: m.Content}},
			})
		}
	}

	return geminiRequest{
		Contents:          contents,
		SystemInstruction: systemContent,
		GenerationConfig: geminiGenerationConfig{
			Temperature:     float64(req.Temperature),
			MaxOutputTokens: req.MaxTokens,
		},
	}, nil
}

func (a *GeminiAdapter) Generate(
	ctx context.Context,
	apiKey, baseURL string,
	req dto.ChatCompletionRequest,
) (*dto.ChatCompletionResponse, error) {
	gReq, err := a.formatRequest(req)
	if err != nil {
		return nil, err
	}

	bodyBytes, err := json.Marshal(gReq)
	if err != nil {
		return nil, err
	}

	endpoint := a.getEndpoint(baseURL, req.Model, "generateContent")
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-goog-api-key", apiKey)

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProviderAPIError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: HTTP %d - %s", ErrProviderAPIError, resp.StatusCode, string(respBody))
	}

	var gResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&gResp); err != nil {
		return nil, fmt.Errorf("failed to decode Gemini response: %w", err)
	}

	var text string
	var finishReason string
	if len(gResp.Candidates) > 0 {
		candidate := gResp.Candidates[0]
		for _, p := range candidate.Content.Parts {
			text += p.Text
		}
		finishReason = strings.ToLower(candidate.FinishReason)
		if finishReason == "stop" {
			finishReason = "stop"
		}
	}

	return &dto.ChatCompletionResponse{
		ID:      fmt.Sprintf("gemini-%d", time.Now().UnixNano()),
		Object:  "chat.completion",
		Created: time.Now().UTC().Unix(),
		Model:   req.Model,
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
			PromptTokens:     gResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: gResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      gResp.UsageMetadata.TotalTokenCount,
		},
	}, nil
}

func (a *GeminiAdapter) StreamGenerate(
	ctx context.Context,
	apiKey, baseURL string,
	req dto.ChatCompletionRequest,
) (<-chan inference.StreamChunk, error) {
	gReq, err := a.formatRequest(req)
	if err != nil {
		return nil, err
	}

	bodyBytes, err := json.Marshal(gReq)
	if err != nil {
		return nil, err
	}

	endpoint := a.getEndpoint(baseURL, req.Model, "streamGenerateContent?alt=sse")
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-goog-api-key", apiKey)

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
				var gResp geminiResponse
				if err := json.Unmarshal([]byte(dataStr), &gResp); err != nil {
					continue
				}

				var text string
				if len(gResp.Candidates) > 0 {
					for _, p := range gResp.Candidates[0].Content.Parts {
						text += p.Text
					}
				}

				if gResp.UsageMetadata.PromptTokenCount > 0 {
					promptTokens = gResp.UsageMetadata.PromptTokenCount
				}
				if gResp.UsageMetadata.CandidatesTokenCount > 0 {
					completionTokens = gResp.UsageMetadata.CandidatesTokenCount
				}

				outChan <- inference.StreamChunk{
					Content:      text,
					Done:         false,
					PromptTokens: promptTokens,
					OutputTokens: completionTokens,
					TotalTokens:  promptTokens + completionTokens,
				}
			}
		}

		outChan <- inference.StreamChunk{
			Done:         true,
			PromptTokens: promptTokens,
			OutputTokens: completionTokens,
			TotalTokens:  promptTokens + completionTokens,
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
