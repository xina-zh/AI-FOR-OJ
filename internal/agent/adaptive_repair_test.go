package agent

import (
	"context"
	"testing"

	"ai-for-oj/internal/llm"
)

func TestClassifyFailure(t *testing.T) {
	tests := []struct {
		name string
		in   FailureObservation
		want FailureType
	}{
		{
			name: "wrong answer verdict",
			in: FailureObservation{
				Verdict:     "WA",
				PassedCount: 1,
				TotalCount:  3,
				ExecStage:   "run",
			},
			want: FailureTypeWrongAnswer,
		},
		{
			name: "runtime error verdict",
			in: FailureObservation{
				Verdict:     "RE",
				RunStderr:   "segmentation fault",
				PassedCount: 0,
				TotalCount:  3,
				ExecStage:   "run",
			},
			want: FailureTypeRuntimeError,
		},
		{
			name: "time limit verdict",
			in: FailureObservation{
				Verdict:     "TLE",
				PassedCount: 2,
				TotalCount:  3,
				ExecStage:   "run",
			},
			want: FailureTypeTimeLimit,
		},
		{
			name: "time limit timeout flag",
			in: FailureObservation{
				Verdict:     "",
				TimedOut:    true,
				PassedCount: 0,
				TotalCount:  3,
				ExecStage:   "run",
			},
			want: FailureTypeTimeLimit,
		},
		{
			name: "unknown empty verdict",
			in: FailureObservation{
				PassedCount: 1,
				TotalCount:  3,
				ExecStage:   "run",
			},
			want: FailureTypeUnknown,
		},
		{
			name: "unknown other verdict",
			in: FailureObservation{
				Verdict:       "CE",
				CompileStderr: "compiler error",
				PassedCount:   0,
				TotalCount:    3,
				ExecStage:     "compile",
			},
			want: FailureTypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyFailure(tt.in)
			if got != tt.want {
				t.Fatalf("ClassifyFailure(...) = %q, want %q", got, tt.want)
			}
		})
	}
}

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
