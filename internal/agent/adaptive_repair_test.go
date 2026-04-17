package agent

import (
	"context"
	"testing"

	"ai-for-oj/internal/llm"
)

type fakeExecutorLLMClient struct {
	request llm.GenerateRequest
	resp    llm.GenerateResponse
	err     error
}

func (c *fakeExecutorLLMClient) Generate(_ context.Context, req llm.GenerateRequest) (llm.GenerateResponse, error) {
	c.request = req
	return c.resp, c.err
}

func TestExecuteLLMOnceReturnsExecutionMetadata(t *testing.T) {
	client := &fakeExecutorLLMClient{
		resp: llm.GenerateResponse{
			Model:        "",
			Content:      "```cpp\nint main() { return 0; }\n```",
			InputTokens:  17,
			OutputTokens: 9,
		},
	}

	got, err := executeLLMOnce(context.Background(), client, "default-model", "solve the problem")
	if err != nil {
		t.Fatalf("executeLLMOnce returned error: %v", err)
	}

	if client.request.Model != "default-model" {
		t.Fatalf("Generate model = %q, want %q", client.request.Model, "default-model")
	}
	if client.request.Prompt != "solve the problem" {
		t.Fatalf("Generate prompt = %q, want %q", client.request.Prompt, "solve the problem")
	}
	if got.PromptPreview != "solve the problem" {
		t.Fatalf("PromptPreview = %q, want %q", got.PromptPreview, "solve the problem")
	}
	if got.RawResponse != client.resp.Content {
		t.Fatalf("RawResponse = %q, want %q", got.RawResponse, client.resp.Content)
	}
	if got.TokenInput != client.resp.InputTokens {
		t.Fatalf("TokenInput = %d, want %d", got.TokenInput, client.resp.InputTokens)
	}
	if got.TokenOutput != client.resp.OutputTokens {
		t.Fatalf("TokenOutput = %d, want %d", got.TokenOutput, client.resp.OutputTokens)
	}
	if got.Model != "default-model" {
		t.Fatalf("Model = %q, want %q", got.Model, "default-model")
	}
	if got.LLMLatencyMS < 0 {
		t.Fatalf("LLMLatencyMS = %d, want >= 0", got.LLMLatencyMS)
	}
}
