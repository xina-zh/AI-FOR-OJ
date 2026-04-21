package llm

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"ai-for-oj/internal/config"
)

const (
	ProviderMock             = "mock"
	ProviderOpenAICompatible = "openai_compatible"
	defaultMockModel         = "mock-cpp17"
	defaultMockResponse      = "```cpp\n#include <bits/stdc++.h>\nusing namespace std;\nint main(){ios::sync_with_stdio(false);cin.tie(nullptr);string line;bool first=true;while(getline(cin,line)){if(!first) cout << \"\\n\";cout << line;first=false;}return 0;}\n```"
	defaultChatMaxTokens     = 4096
)

var errEmptyResponseBody = errors.New("llm response body is empty")

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

type openAICompatibleEndpoint struct {
	baseURL string
	apiKey  string
}

type modelEndpointRoute struct {
	modelPrefix string
	endpoint    openAICompatibleEndpoint
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
		defaultEndpoint, err := newOpenAICompatibleEndpoint(defaultString(cfg.BaseURL, "https://api.openai.com/v1"), cfg.APIKey)
		if err != nil {
			return nil, err
		}
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = 60 * time.Second
		}

		var routes []modelEndpointRoute
		if strings.TrimSpace(cfg.GLMBaseURL) != "" || strings.TrimSpace(cfg.GLMAPIKey) != "" || strings.TrimSpace(cfg.GLMModelPrefix) != "" {
			if strings.TrimSpace(cfg.GLMAPIKey) == "" {
				return nil, fmt.Errorf("llm glm api key is required when glm route is configured")
			}
			glmEndpoint, err := newOpenAICompatibleEndpoint(cfg.GLMBaseURL, cfg.GLMAPIKey)
			if err != nil {
				return nil, fmt.Errorf("invalid llm glm route: %w", err)
			}
			routes = append(routes, modelEndpointRoute{
				modelPrefix: defaultString(cfg.GLMModelPrefix, "glm-"),
				endpoint:    glmEndpoint,
			})
		}

		return &OpenAICompatibleClient{
			endpoint: defaultEndpoint,
			routes:   routes,
			model:    cfg.Model,
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
	endpoint openAICompatibleEndpoint
	routes   []modelEndpointRoute
	model    string
	client   *http.Client
	logger   *slog.Logger
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
	resolvedModel := effectiveModel(req.Model, c.model)
	endpoint := c.endpointForModel(resolvedModel)
	requestURL := endpoint.baseURL + "/chat/completions"
	requestBody := openAICompatibleRequest{
		Model: resolvedModel,
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

	respBody, parsed, err := c.doChatCompletion(ctx, endpoint, requestURL, body)
	if err != nil {
		if c.logger != nil {
			c.logger.Warn("llm generation failed",
				"provider", ProviderOpenAICompatible,
				"model", resolvedModel,
				"url", requestURL,
				"error", err,
			)
		}
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

func (c *OpenAICompatibleClient) endpointForModel(model string) openAICompatibleEndpoint {
	model = strings.TrimSpace(model)
	for _, route := range c.routes {
		if strings.HasPrefix(model, strings.TrimSpace(route.modelPrefix)) {
			return route.endpoint
		}
	}
	return c.endpoint
}

type parsedOpenAICompatibleResponse struct {
	openAICompatibleResponse
	StatusCode int
}

func (c *OpenAICompatibleClient) doChatCompletion(
	ctx context.Context,
	endpoint openAICompatibleEndpoint,
	requestURL string,
	body []byte,
) ([]byte, parsedOpenAICompatibleResponse, error) {
	const maxAttempts = 5

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		respBody, parsed, err := c.doChatCompletionOnce(ctx, endpoint, requestURL, body)
		retryableStatus := err == nil && isRetryableStatusCode(parsed.StatusCode)
		if err == nil && (!retryableStatus || attempt == maxAttempts) {
			return respBody, parsed, nil
		}
		retryable := isRetryableRequestError(ctx, err) || retryableStatus
		if c.logger != nil {
			c.logger.Warn("llm request attempt failed",
				"provider", ProviderOpenAICompatible,
				"url", requestURL,
				"attempt", attempt,
				"max_attempts", maxAttempts,
				"status", parsed.StatusCode,
				"retryable", retryable,
				"error", err,
			)
		}
		if err == nil {
			err = fmt.Errorf("llm request failed url=%s status=%d", requestURL, parsed.StatusCode)
		}
		if attempt == maxAttempts || !retryable {
			return respBody, parsed, err
		}
		if transport, ok := c.client.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
		time.Sleep(retryDelay(attempt))
	}

	return nil, parsedOpenAICompatibleResponse{}, fmt.Errorf("llm request failed url=%s: exhausted retries", requestURL)
}

func (c *OpenAICompatibleClient) doChatCompletionOnce(
	ctx context.Context,
	endpoint openAICompatibleEndpoint,
	requestURL string,
	body []byte,
) ([]byte, parsedOpenAICompatibleResponse, error) {
	requestBodyPreview := truncateErrorBody(strings.TrimSpace(string(body)), 512)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(body))
	if err != nil {
		return nil, parsedOpenAICompatibleResponse{}, fmt.Errorf("build llm request url=%s: %w", requestURL, err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+endpoint.apiKey)
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
	if resp.StatusCode >= http.StatusBadRequest {
		_ = json.Unmarshal(respBody, &parsed)
		return respBody, parsedOpenAICompatibleResponse{
			openAICompatibleResponse: parsed,
			StatusCode:               resp.StatusCode,
		}, nil
	}
	if strings.TrimSpace(string(respBody)) == "" {
		return nil, parsedOpenAICompatibleResponse{StatusCode: resp.StatusCode}, fmt.Errorf(
			"decode llm response url=%s status=%d request_body=%s: %w",
			requestURL,
			resp.StatusCode,
			requestBodyPreview,
			errEmptyResponseBody,
		)
	}
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

func newOpenAICompatibleEndpoint(baseURL, apiKey string) (openAICompatibleEndpoint, error) {
	normalizedBaseURL := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if _, err := url.ParseRequestURI(normalizedBaseURL); err != nil {
		return openAICompatibleEndpoint{}, fmt.Errorf("invalid llm base url %q: %w", normalizedBaseURL, err)
	}
	return openAICompatibleEndpoint{
		baseURL: normalizedBaseURL,
		apiKey:  strings.TrimSpace(apiKey),
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

func isRetryableRequestError(ctx context.Context, err error) bool {
	if err == nil {
		return false
	}
	if ctx != nil && ctx.Err() != nil {
		return false
	}
	return isRetryableTransportError(err) || isRetryableTimeoutError(err) || errors.Is(err, errEmptyResponseBody)
}

func isRetryableTransportError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unexpected eof") ||
		strings.Contains(message, "eof") ||
		strings.Contains(message, "connection reset by peer")
}

func isRetryableTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	return errors.Is(err, context.DeadlineExceeded)
}

func isRetryableStatusCode(statusCode int) bool {
	switch statusCode {
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func retryDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	base := 300 * time.Millisecond
	delay := base << (attempt - 1)
	if delay > 3*time.Second {
		delay = 3 * time.Second
	}
	jitter := time.Duration(rand.Intn(151)) * time.Millisecond
	return delay + jitter
}

func truncateErrorBody(body string, limit int) string {
	body = strings.TrimSpace(body)
	if limit <= 0 {
		return body
	}
	if utf8.RuneCountInString(body) <= limit {
		return body
	}
	return string([]rune(body)[:limit]) + "...(truncated)"
}
