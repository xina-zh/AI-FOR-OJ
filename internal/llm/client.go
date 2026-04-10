package llm

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ai-for-oj/internal/config"
)

const (
	ProviderMock             = "mock"
	ProviderOpenAICompatible = "openai_compatible"
	defaultMockModel         = "mock-cpp17"
	defaultMockResponse      = "```cpp\n#include <bits/stdc++.h>\nusing namespace std;\nint main(){ios::sync_with_stdio(false);cin.tie(nullptr);string line;bool first=true;while(getline(cin,line)){if(!first) cout << \"\\n\";cout << line;first=false;}return 0;}\n```"
	defaultChatMaxTokens     = 4096
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
		if _, err := url.ParseRequestURI(baseURL); err != nil {
			return nil, fmt.Errorf("invalid llm base url %q: %w", baseURL, err)
		}
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
		return &OpenAICompatibleClient{
			baseURL: baseURL,
			apiKey:  cfg.APIKey,
			model:   cfg.Model,
			client: &http.Client{
				Timeout:   timeout,
				Transport: newOpenAICompatibleTransport(),
			},
			logger: logger,
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
	Model     string                    `json:"model"`
	Messages  []openAICompatibleMessage `json:"messages"`
	MaxTokens int                       `json:"max_tokens,omitempty"`
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
	requestURL := c.baseURL + "/chat/completions"
	requestBody := openAICompatibleRequest{
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
		MaxTokens: defaultChatMaxTokens,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return GenerateResponse{}, fmt.Errorf("marshal llm request: %w", err)
	}

	respBody, parsed, err := c.doChatCompletion(ctx, requestURL, body)
	if err != nil {
		return GenerateResponse{}, err
	}

	if parsed.StatusCode >= http.StatusBadRequest {
		if parsed.Error != nil && strings.TrimSpace(parsed.Error.Message) != "" {
			return GenerateResponse{}, fmt.Errorf("llm request failed url=%s status=%d: %s", requestURL, parsed.StatusCode, parsed.Error.Message)
		}
		if trimmed := strings.TrimSpace(string(respBody)); trimmed != "" {
			return GenerateResponse{}, fmt.Errorf("llm request failed url=%s status=%d body=%s", requestURL, parsed.StatusCode, truncateErrorBody(trimmed, 512))
		}
		return GenerateResponse{}, fmt.Errorf("llm request failed url=%s status=%d", requestURL, parsed.StatusCode)
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

type parsedOpenAICompatibleResponse struct {
	openAICompatibleResponse
	StatusCode int
}

func (c *OpenAICompatibleClient) doChatCompletion(ctx context.Context, requestURL string, body []byte) ([]byte, parsedOpenAICompatibleResponse, error) {
	const maxAttempts = 2

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		respBody, parsed, err := c.doChatCompletionOnce(ctx, requestURL, body)
		if err == nil {
			return respBody, parsed, nil
		}
		if attempt == maxAttempts || !isRetryableTransportError(err) {
			return nil, parsed, err
		}
		if transport, ok := c.client.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}

	return nil, parsedOpenAICompatibleResponse{}, fmt.Errorf("llm request failed url=%s: exhausted retries", requestURL)
}

func (c *OpenAICompatibleClient) doChatCompletionOnce(ctx context.Context, requestURL string, body []byte) ([]byte, parsedOpenAICompatibleResponse, error) {
	requestBodyPreview := truncateErrorBody(strings.TrimSpace(string(body)), 512)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(body))
	if err != nil {
		return nil, parsedOpenAICompatibleResponse{}, fmt.Errorf("build llm request url=%s: %w", requestURL, err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", "ai-for-oj/llm-client")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, parsedOpenAICompatibleResponse{}, fmt.Errorf(
			"send llm request url=%s request_body=%s: %w",
			requestURL,
			requestBodyPreview,
			err,
		)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, parsedOpenAICompatibleResponse{StatusCode: resp.StatusCode}, fmt.Errorf(
			"read llm response url=%s status=%d request_body=%s: %w",
			requestURL,
			resp.StatusCode,
			requestBodyPreview,
			err,
		)
	}

	var parsed openAICompatibleResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, parsedOpenAICompatibleResponse{StatusCode: resp.StatusCode}, fmt.Errorf(
			"decode llm response url=%s status=%d request_body=%s body=%s: %w",
			requestURL,
			resp.StatusCode,
			requestBodyPreview,
			truncateErrorBody(strings.TrimSpace(string(respBody)), 512),
			err,
		)
	}

	return respBody, parsedOpenAICompatibleResponse{
		openAICompatibleResponse: parsed,
		StatusCode:               resp.StatusCode,
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

func newOpenAICompatibleTransport() *http.Transport {
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2:     false,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSNextProto:          map[string]func(string, *tls.Conn) http.RoundTripper{},
	}
}

func isRetryableTransportError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unexpected eof") || strings.Contains(message, "eof")
}

func truncateErrorBody(body string, limit int) string {
	body = strings.TrimSpace(body)
	if limit <= 0 || len(body) <= limit {
		return body
	}
	return body[:limit] + "...(truncated)"
}
