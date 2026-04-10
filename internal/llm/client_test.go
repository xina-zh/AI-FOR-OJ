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

func containsAll(value string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(value, part) {
			return false
		}
	}
	return true
}
