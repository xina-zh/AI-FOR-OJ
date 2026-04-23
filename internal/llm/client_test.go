package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf8"

	"ai-for-oj/internal/config"
)

func TestNewClientMockGenerate(t *testing.T) {
	client, err := NewClient(config.LLMConfig{
		Provider:     ProviderMock,
		Model:        "mock-cpp17",
		MockResponse: "```cpp\nint main(){return 0;}\n```",
	}, nil)
	if err != nil {
		t.Fatalf("new mock client returned error: %v", err)
	}

	resp, err := client.Generate(context.Background(), GenerateRequest{
		Prompt: "solve echo",
		Model:  "",
	})
	if err != nil {
		t.Fatalf("mock generate returned error: %v", err)
	}

	if resp.Model != "mock-cpp17" {
		t.Fatalf("expected mock model, got %q", resp.Model)
	}
	if resp.InputTokens == 0 || resp.OutputTokens == 0 {
		t.Fatalf("expected mock token counts, got %+v", resp)
	}
}

func TestNewClientOpenAICompatibleParsesResponse(t *testing.T) {
	var rawBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected auth header: %s", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("unexpected content type: %s", got)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Fatalf("unexpected accept header: %s", got)
		}
		if got := r.Header.Get("User-Agent"); got != "ai-for-oj/llm-client" {
			t.Fatalf("unexpected user agent: %s", got)
		}

		var req openAICompatibleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		bodyBytes, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("re-marshal request: %v", err)
		}
		if err := json.Unmarshal(bodyBytes, &rawBody); err != nil {
			t.Fatalf("decode remarshal request: %v", err)
		}
		if req.Model != "gpt-test" {
			t.Fatalf("unexpected request model: %s", req.Model)
		}
		if req.MaxTokens != defaultChatMaxTokens {
			t.Fatalf("unexpected max_tokens: %d", req.MaxTokens)
		}
		if _, ok := rawBody["temperature"]; ok {
			t.Fatalf("did not expect temperature in request body: %+v", rawBody)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"model": "gpt-test",
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"role":    "assistant",
						"content": "```cpp\nint main(){return 0;}\n```",
					},
				},
			},
			"usage": map[string]any{
				"prompt_tokens":     111,
				"completion_tokens": 22,
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(config.LLMConfig{
		Provider: ProviderOpenAICompatible,
		BaseURL:  server.URL,
		APIKey:   "test-key",
		Model:    "gpt-test",
		Timeout:  5 * time.Second,
	}, nil)
	if err != nil {
		t.Fatalf("new openai compatible client returned error: %v", err)
	}

	resp, err := client.Generate(context.Background(), GenerateRequest{
		Prompt: "solve echo",
		Model:  "",
	})
	if err != nil {
		t.Fatalf("openai compatible generate returned error: %v", err)
	}

	if resp.Model != "gpt-test" {
		t.Fatalf("expected parsed model, got %q", resp.Model)
	}
	if resp.InputTokens != 111 || resp.OutputTokens != 22 {
		t.Fatalf("expected parsed usage, got %+v", resp)
	}
	if !reflect.DeepEqual(rawBody["messages"], []any{
		map[string]any{
			"role":    "system",
			"content": "You are a competitive programming assistant. Return a correct cpp17 solution. Prefer returning only a markdown cpp code block.",
		},
		map[string]any{
			"role":    "user",
			"content": "solve echo",
		},
	}) {
		t.Fatalf("unexpected request messages: %+v", rawBody["messages"])
	}
}

func TestNewClientOpenAICompatibleRequiresAPIKey(t *testing.T) {
	_, err := NewClient(config.LLMConfig{
		Provider: ProviderOpenAICompatible,
		BaseURL:  "https://api.gptsapi.net/v1",
		Model:    "gpt-test",
	}, nil)
	if err == nil {
		t.Fatal("expected missing api key error")
	}
}

func TestNewClientOpenAICompatibleRejectsInvalidBaseURL(t *testing.T) {
	_, err := NewClient(config.LLMConfig{
		Provider: ProviderOpenAICompatible,
		BaseURL:  "://bad-url",
		APIKey:   "test-key",
		Model:    "gpt-test",
	}, nil)
	if err == nil {
		t.Fatal("expected invalid base url error")
	}
}

func TestOpenAICompatibleClientRoutesGLMModelsByRequestModel(t *testing.T) {
	var defaultCalls int32
	glmServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer glm-key" {
			t.Fatalf("unexpected glm auth header: %s", got)
		}

		var req openAICompatibleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode glm request: %v", err)
		}
		if req.Model != "glm-4.5" {
			t.Fatalf("unexpected glm request model: %s", req.Model)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"model": "glm-4.5",
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"role":    "assistant",
						"content": "```cpp\nint main(){return 0;}\n```",
					},
				},
			},
		})
	}))
	defer glmServer.Close()

	defaultServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&defaultCalls, 1)
		t.Fatal("default endpoint should not be used for glm request model")
	}))
	defer defaultServer.Close()

	client, err := NewClient(config.LLMConfig{
		Provider:       ProviderOpenAICompatible,
		BaseURL:        defaultServer.URL,
		APIKey:         "default-key",
		Model:          "gpt-test",
		Timeout:        5 * time.Second,
		GLMBaseURL:     glmServer.URL,
		GLMAPIKey:      "glm-key",
		GLMModelPrefix: "glm-",
	}, nil)
	if err != nil {
		t.Fatalf("new openai compatible client returned error: %v", err)
	}

	resp, err := client.Generate(context.Background(), GenerateRequest{
		Prompt: "solve echo",
		Model:  "glm-4.5",
	})
	if err != nil {
		t.Fatalf("glm-routed generate returned error: %v", err)
	}
	if resp.Model != "glm-4.5" {
		t.Fatalf("expected glm model, got %q", resp.Model)
	}
	if atomic.LoadInt32(&defaultCalls) != 0 {
		t.Fatalf("expected default endpoint to remain unused, got %d calls", defaultCalls)
	}
}

func TestOpenAICompatibleClientRoutesGLMModelsByDefaultModel(t *testing.T) {
	var defaultCalls int32
	glmServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer glm-key" {
			t.Fatalf("unexpected glm auth header: %s", got)
		}

		var req openAICompatibleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode glm request: %v", err)
		}
		if req.Model != "glm-4.5" {
			t.Fatalf("unexpected glm request model: %s", req.Model)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"model": "glm-4.5",
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"role":    "assistant",
						"content": "```cpp\nint main(){return 0;}\n```",
					},
				},
			},
		})
	}))
	defer glmServer.Close()

	defaultServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&defaultCalls, 1)
		t.Fatal("default endpoint should not be used for glm default model")
	}))
	defer defaultServer.Close()

	client, err := NewClient(config.LLMConfig{
		Provider:       ProviderOpenAICompatible,
		BaseURL:        defaultServer.URL,
		APIKey:         "default-key",
		Model:          "glm-4.5",
		Timeout:        5 * time.Second,
		GLMBaseURL:     glmServer.URL,
		GLMAPIKey:      "glm-key",
		GLMModelPrefix: "glm-",
	}, nil)
	if err != nil {
		t.Fatalf("new openai compatible client returned error: %v", err)
	}

	resp, err := client.Generate(context.Background(), GenerateRequest{
		Prompt: "solve echo",
	})
	if err != nil {
		t.Fatalf("glm default-model generate returned error: %v", err)
	}
	if resp.Model != "glm-4.5" {
		t.Fatalf("expected glm model, got %q", resp.Model)
	}
	if atomic.LoadInt32(&defaultCalls) != 0 {
		t.Fatalf("expected default endpoint to remain unused, got %d calls", defaultCalls)
	}
}

func TestOpenAICompatibleClientRoutesDeepSeekModelsByRequestModel(t *testing.T) {
	var defaultCalls int32
	deepseekServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer deepseek-key" {
			t.Fatalf("unexpected deepseek auth header: %s", got)
		}
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected deepseek path: %s", r.URL.Path)
		}

		var req openAICompatibleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode deepseek request: %v", err)
		}
		if req.Model != "deepseek-chat" {
			t.Fatalf("unexpected deepseek request model: %s", req.Model)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"model": "deepseek-chat",
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"role":    "assistant",
						"content": "```cpp\nint main(){return 0;}\n```",
					},
				},
			},
			"usage": map[string]any{
				"prompt_tokens":     12,
				"completion_tokens": 6,
			},
		})
	}))
	defer deepseekServer.Close()

	defaultServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&defaultCalls, 1)
		t.Fatal("default endpoint should not be used for deepseek request model")
	}))
	defer defaultServer.Close()

	client, err := NewClient(config.LLMConfig{
		Provider:            ProviderOpenAICompatible,
		BaseURL:             defaultServer.URL,
		APIKey:              "default-key",
		Model:               "gpt-test",
		Timeout:             5 * time.Second,
		DeepSeekBaseURL:     deepseekServer.URL,
		DeepSeekAPIKey:      "deepseek-key",
		DeepSeekModelPrefix: "deepseek-",
	}, nil)
	if err != nil {
		t.Fatalf("new openai compatible client returned error: %v", err)
	}

	resp, err := client.Generate(context.Background(), GenerateRequest{
		Prompt: "solve echo",
		Model:  "deepseek-chat",
	})
	if err != nil {
		t.Fatalf("deepseek-routed generate returned error: %v", err)
	}
	if resp.Model != "deepseek-chat" {
		t.Fatalf("expected deepseek model, got %q", resp.Model)
	}
	if resp.InputTokens != 12 || resp.OutputTokens != 6 {
		t.Fatalf("expected deepseek usage, got %+v", resp)
	}
	if atomic.LoadInt32(&defaultCalls) != 0 {
		t.Fatalf("expected default endpoint to remain unused, got %d calls", defaultCalls)
	}
}

func TestOpenAICompatibleClientRoutesDeepSeekModelsByDefaultModel(t *testing.T) {
	var defaultCalls int32
	deepseekServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer deepseek-key" {
			t.Fatalf("unexpected deepseek auth header: %s", got)
		}

		var req openAICompatibleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode deepseek request: %v", err)
		}
		if req.Model != "deepseek-reasoner" {
			t.Fatalf("unexpected deepseek request model: %s", req.Model)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"model": "deepseek-reasoner",
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"role":    "assistant",
						"content": "```cpp\nint main(){return 0;}\n```",
					},
				},
			},
		})
	}))
	defer deepseekServer.Close()

	defaultServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&defaultCalls, 1)
		t.Fatal("default endpoint should not be used for deepseek default model")
	}))
	defer defaultServer.Close()

	client, err := NewClient(config.LLMConfig{
		Provider:            ProviderOpenAICompatible,
		BaseURL:             defaultServer.URL,
		APIKey:              "default-key",
		Model:               "deepseek-reasoner",
		Timeout:             5 * time.Second,
		DeepSeekBaseURL:     deepseekServer.URL,
		DeepSeekAPIKey:      "deepseek-key",
		DeepSeekModelPrefix: "deepseek-",
	}, nil)
	if err != nil {
		t.Fatalf("new openai compatible client returned error: %v", err)
	}

	resp, err := client.Generate(context.Background(), GenerateRequest{
		Prompt: "solve echo",
	})
	if err != nil {
		t.Fatalf("deepseek default-model generate returned error: %v", err)
	}
	if resp.Model != "deepseek-reasoner" {
		t.Fatalf("expected deepseek model, got %q", resp.Model)
	}
	if atomic.LoadInt32(&defaultCalls) != 0 {
		t.Fatalf("expected default endpoint to remain unused, got %d calls", defaultCalls)
	}
}

func TestNewClientOpenAICompatibleRequiresDeepSeekAPIKeyWhenRouteConfigured(t *testing.T) {
	_, err := NewClient(config.LLMConfig{
		Provider:            ProviderOpenAICompatible,
		APIKey:              "default-key",
		DeepSeekModelPrefix: "deepseek-",
	}, nil)
	if err == nil {
		t.Fatal("expected missing deepseek api key error")
	}
	if !strings.Contains(err.Error(), "deepseek api key") {
		t.Fatalf("expected deepseek api key error, got %v", err)
	}
}

func TestOpenAICompatibleClientRetriesUnexpectedEOF(t *testing.T) {
	var attempts int32
	var requestBodies []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		requestBodies = append(requestBodies, string(bodyBytes))

		current := atomic.AddInt32(&attempts, 1)
		if current == 1 {
			hj, ok := w.(http.Hijacker)
			if !ok {
				t.Fatal("response writer does not support hijacking")
			}
			conn, rw, err := hj.Hijack()
			if err != nil {
				t.Fatalf("hijack: %v", err)
			}
			defer conn.Close()
			_, _ = rw.WriteString("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: 120\r\n\r\n{\"model\":\"gpt-test\"")
			_ = rw.Flush()
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"model": "gpt-test",
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"role":    "assistant",
						"content": "```cpp\nint main(){return 0;}\n```",
					},
				},
			},
			"usage": map[string]any{
				"prompt_tokens":     10,
				"completion_tokens": 5,
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(config.LLMConfig{
		Provider: ProviderOpenAICompatible,
		BaseURL:  server.URL,
		APIKey:   "test-key",
		Model:    "gpt-test",
		Timeout:  5 * time.Second,
	}, nil)
	if err != nil {
		t.Fatalf("new openai compatible client returned error: %v", err)
	}

	resp, err := client.Generate(context.Background(), GenerateRequest{
		Prompt: "solve echo",
	})
	if err != nil {
		t.Fatalf("generate should succeed after retry, got error: %v", err)
	}
	if atomic.LoadInt32(&attempts) != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
	if len(requestBodies) != 2 {
		t.Fatalf("expected 2 request bodies, got %d", len(requestBodies))
	}
	if requestBodies[0] != requestBodies[1] {
		t.Fatalf("expected retry to resend identical body, got %q vs %q", requestBodies[0], requestBodies[1])
	}
	if resp.Model != "gpt-test" || resp.InputTokens != 10 || resp.OutputTokens != 5 {
		t.Fatalf("unexpected retried response: %+v", resp)
	}
}

func TestOpenAICompatibleClientReturnsClearReadError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("response writer does not support hijacking")
		}
		conn, rw, err := hj.Hijack()
		if err != nil {
			t.Fatalf("hijack: %v", err)
		}
		defer conn.Close()
		_, _ = rw.WriteString("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: 120\r\n\r\n{\"model\":\"gpt-test\"")
		_ = rw.Flush()
	}))
	defer server.Close()

	client, err := NewClient(config.LLMConfig{
		Provider: ProviderOpenAICompatible,
		BaseURL:  server.URL,
		APIKey:   "test-key",
		Model:    "gpt-test",
		Timeout:  5 * time.Second,
	}, nil)
	if err != nil {
		t.Fatalf("new openai compatible client returned error: %v", err)
	}

	_, err = client.Generate(context.Background(), GenerateRequest{
		Prompt: "solve echo",
	})
	if err == nil {
		t.Fatal("expected read error")
	}
	if got := err.Error(); !containsAll(got, "read llm response", server.URL, "status=200", "request_body=") {
		t.Fatalf("expected clearer error context, got %q", got)
	}
}

func TestOpenAICompatibleClientReturnsClearNonJSONStatusError(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		http.Error(w, "upstream gateway error", http.StatusBadGateway)
	}))
	defer server.Close()

	client, err := NewClient(config.LLMConfig{
		Provider: ProviderOpenAICompatible,
		BaseURL:  server.URL,
		APIKey:   "test-key",
		Model:    "gpt-test",
		Timeout:  5 * time.Second,
	}, nil)
	if err != nil {
		t.Fatalf("new openai compatible client returned error: %v", err)
	}

	_, err = client.Generate(context.Background(), GenerateRequest{
		Prompt: "solve echo",
	})
	if err == nil {
		t.Fatal("expected status error")
	}
	if atomic.LoadInt32(&attempts) != 5 {
		t.Fatalf("expected 5 attempts for retryable 502, got %d", attempts)
	}
	if got := err.Error(); !containsAll(got, "llm request failed", "status=502", "upstream gateway error") {
		t.Fatalf("expected clearer non-json status error, got %q", got)
	}
}

func TestOpenAICompatibleClientRetriesTimeoutOnce(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt32(&attempts, 1)
		if current == 1 {
			time.Sleep(120 * time.Millisecond)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"model": "gpt-test",
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"role":    "assistant",
						"content": "```cpp\nint main(){return 0;}\n```",
					},
				},
			},
			"usage": map[string]any{
				"prompt_tokens":     8,
				"completion_tokens": 4,
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(config.LLMConfig{
		Provider: ProviderOpenAICompatible,
		BaseURL:  server.URL,
		APIKey:   "test-key",
		Model:    "gpt-test",
		Timeout:  50 * time.Millisecond,
	}, nil)
	if err != nil {
		t.Fatalf("new openai compatible client returned error: %v", err)
	}

	resp, err := client.Generate(context.Background(), GenerateRequest{
		Prompt: "solve echo",
	})
	if err != nil {
		t.Fatalf("generate should succeed after timeout retry, got error: %v", err)
	}
	if atomic.LoadInt32(&attempts) != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
	if resp.Model != "gpt-test" || resp.InputTokens != 8 || resp.OutputTokens != 4 {
		t.Fatalf("unexpected retried timeout response: %+v", resp)
	}
}

func TestOpenAICompatibleClientRetriesEmptyBody(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt32(&attempts, 1)
		if current < 3 {
			w.WriteHeader(http.StatusOK)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"model": "gpt-test",
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"role":    "assistant",
						"content": "```cpp\nint main(){return 0;}\n```",
					},
				},
			},
			"usage": map[string]any{
				"prompt_tokens":     9,
				"completion_tokens": 6,
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(config.LLMConfig{
		Provider: ProviderOpenAICompatible,
		BaseURL:  server.URL,
		APIKey:   "test-key",
		Model:    "gpt-test",
		Timeout:  5 * time.Second,
	}, nil)
	if err != nil {
		t.Fatalf("new openai compatible client returned error: %v", err)
	}

	resp, err := client.Generate(context.Background(), GenerateRequest{
		Prompt: "solve echo",
	})
	if err != nil {
		t.Fatalf("generate should succeed after empty-body retries, got error: %v", err)
	}
	if atomic.LoadInt32(&attempts) != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
	if resp.Model != "gpt-test" || resp.InputTokens != 9 || resp.OutputTokens != 6 {
		t.Fatalf("unexpected retried empty-body response: %+v", resp)
	}
}

func TestTruncateErrorBodyKeepsUTF8Valid(t *testing.T) {
	value := "中文错误响应内容abc"
	got := truncateErrorBody(value, 4)
	if !utf8.ValidString(got) {
		t.Fatalf("expected utf8-valid truncated body, got %q", got)
	}
	if got != "中文错误...(truncated)" {
		t.Fatalf("unexpected truncated body: %q", got)
	}
}

func containsAll(value string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(value, part) {
			return false
		}
	}
	return true
}
