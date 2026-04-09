package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"ai-for-oj/internal/config"
)

const (
	ProviderMock             = "mock"
	ProviderOpenAICompatible = "openai_compatible"
	defaultMockModel         = "mock-cpp17"
	defaultMockResponse      = "```cpp\n#include <bits/stdc++.h>\nusing namespace std;\nint main(){ios::sync_with_stdio(false);cin.tie(nullptr);string line;bool first=true;while(getline(cin,line)){if(!first) cout << \"\\n\";cout << line;first=false;}return 0;}\n```"
)

type Client interface {
	Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error)
}

type GenerateRequest struct {
	Prompt string
	Model  string
}

type GenerateResponse struct {
	Model        string
	Content      string
	InputTokens  int64
	OutputTokens int64
}

func NewClient(cfg config.LLMConfig, logger *slog.Logger) (Client, error) {
	switch cfg.Provider {
	case "", ProviderMock:
		return &MockClient{
			model:    effectiveModel(cfg.Model, defaultMockModel),
			response: defaultString(cfg.MockResponse, defaultMockResponse),
		}, nil
	case ProviderOpenAICompatible:
		if strings.TrimSpace(cfg.APIKey) == "" {
			return nil, fmt.Errorf("llm api key is required when provider=%s", ProviderOpenAICompatible)
		}
		baseURL := strings.TrimRight(defaultString(cfg.BaseURL, "https://api.openai.com/v1"), "/")
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
		return &OpenAICompatibleClient{
			baseURL: baseURL,
			apiKey:  cfg.APIKey,
			model:   cfg.Model,
			client:  &http.Client{Timeout: timeout},
			logger:  logger,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported llm provider: %s", cfg.Provider)
	}
}

type MockClient struct {
	model    string
	response string
}

func (c *MockClient) Generate(_ context.Context, req GenerateRequest) (GenerateResponse, error) {
	return GenerateResponse{
		Model:        effectiveModel(req.Model, c.model),
		Content:      c.response,
		InputTokens:  roughTokenCount(req.Prompt),
		OutputTokens: roughTokenCount(c.response),
	}, nil
}

type OpenAICompatibleClient struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
	logger  *slog.Logger
}

type openAICompatibleRequest struct {
	Model       string                    `json:"model"`
	Messages    []openAICompatibleMessage `json:"messages"`
	Temperature float64                   `json:"temperature"`
}

type openAICompatibleMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAICompatibleResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message openAICompatibleMessage `json:"message"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int64 `json:"prompt_tokens"`
		CompletionTokens int64 `json:"completion_tokens"`
	} `json:"usage,omitempty"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *OpenAICompatibleClient) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	body, err := json.Marshal(openAICompatibleRequest{
		Model: effectiveModel(req.Model, c.model),
		Messages: []openAICompatibleMessage{
			{
				Role:    "system",
				Content: "You are a competitive programming assistant. Return a correct cpp17 solution. Prefer returning only a markdown cpp code block.",
			},
			{
				Role:    "user",
				Content: req.Prompt,
			},
		},
		Temperature: 0.2,
	})
	if err != nil {
		return GenerateResponse{}, fmt.Errorf("marshal llm request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return GenerateResponse{}, fmt.Errorf("build llm request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return GenerateResponse{}, fmt.Errorf("send llm request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return GenerateResponse{}, fmt.Errorf("read llm response: %w", err)
	}

	var parsed openAICompatibleResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return GenerateResponse{}, fmt.Errorf("decode llm response: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		if parsed.Error != nil && strings.TrimSpace(parsed.Error.Message) != "" {
			return GenerateResponse{}, fmt.Errorf("llm request failed: %s", parsed.Error.Message)
		}
		return GenerateResponse{}, fmt.Errorf("llm request failed with status %d", resp.StatusCode)
	}

	if len(parsed.Choices) == 0 {
		return GenerateResponse{}, fmt.Errorf("llm response has no choices")
	}

	content := parsed.Choices[0].Message.Content
	if c.logger != nil {
		c.logger.Info("llm generation completed", "provider", ProviderOpenAICompatible, "model", effectiveModel(parsed.Model, req.Model, c.model))
	}

	return GenerateResponse{
		Model:        effectiveModel(parsed.Model, req.Model, c.model),
		Content:      content,
		InputTokens:  usagePromptTokens(parsed.Usage),
		OutputTokens: usageCompletionTokens(parsed.Usage),
	}, nil
}

func effectiveModel(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func roughTokenCount(value string) int64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	return int64(len(strings.Fields(value)))
}

func usagePromptTokens(usage *struct {
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
}) int64 {
	if usage == nil {
		return 0
	}
	return usage.PromptTokens
}

func usageCompletionTokens(usage *struct {
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
}) int64 {
	if usage == nil {
		return 0
	}
	return usage.CompletionTokens
}
