package agent

import (
	"context"
	"testing"

	"ai-for-oj/internal/llm"
)

func TestClassifyFailure(t *testing.T) {
	tests := []struct {
		name string
		in   JudgeFailureObservation
		want FailureType
	}{
		{
			name: "wrong answer verdict",
			in: JudgeFailureObservation{
				Verdict:     "WA",
				PassedCount: 1,
				TotalCount:  3,
				ExecStage:   "run",
			},
			want: FailureTypeWrongAnswer,
		},
		{
			name: "runtime error verdict",
			in: JudgeFailureObservation{
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
			in: JudgeFailureObservation{
				Verdict:     "TLE",
				PassedCount: 2,
				TotalCount:  3,
				ExecStage:   "run",
			},
			want: FailureTypeTimeLimit,
		},
		{
			name: "time limit timeout flag",
			in: JudgeFailureObservation{
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
			in: JudgeFailureObservation{
				PassedCount: 1,
				TotalCount:  3,
				ExecStage:   "run",
			},
			want: FailureTypeUnknown,
		},
		{
			name: "CE is unknown",
			in: JudgeFailureObservation{
				Verdict:       "CE",
				CompileStderr: "compiler error",
				PassedCount:   0,
				TotalCount:    3,
				ExecStage:     "compile",
			},
			want: FailureTypeUnknown,
		},
		{
			name: "compile stage stderr is unknown",
			in: JudgeFailureObservation{
				Verdict:       "",
				CompileStderr: "compiler error",
				PassedCount:   0,
				TotalCount:    3,
				ExecStage:     "compile",
			},
			want: FailureTypeUnknown,
		},
		{
			name: "other verdict stays unknown",
			in: JudgeFailureObservation{
				Verdict:     "AC",
				PassedCount: 3,
				TotalCount:  3,
				ExecStage:   "run",
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

func TestRepairPlanner(t *testing.T) {
	planner := NewRepairPlanner(3)

	tests := []struct {
		name          string
		in            RepairPlanInput
		wantStage     string
		allowFallback bool
		wantStop      bool
	}{
		{
			name: "wrong answer after initial attempt",
			in: RepairPlanInput{
				AttemptCount:   1,
				LastFailure:    FailureTypeWrongAnswer,
				PreviousStages: nil,
				MaxBudget:      3,
			},
			wantStage: "wa_analysis_repair",
		},
		{
			name: "runtime error routes to safety repair",
			in: RepairPlanInput{
				AttemptCount:   1,
				LastFailure:    FailureTypeRuntimeError,
				PreviousStages: nil,
				MaxBudget:      3,
			},
			wantStage: "re_safety_repair",
		},
		{
			name: "time limit routes to complexity rewrite",
			in: RepairPlanInput{
				AttemptCount:   1,
				LastFailure:    FailureTypeTimeLimit,
				PreviousStages: nil,
				MaxBudget:      3,
			},
			wantStage: "tle_complexity_rewrite",
		},
		{
			name: "repeated same failure beyond budget falls back or stops",
			in: RepairPlanInput{
				AttemptCount:   3,
				LastFailure:    FailureTypeWrongAnswer,
				PreviousStages: []string{"wa_analysis_repair", "wa_analysis_repair"},
				MaxBudget:      3,
			},
			wantStage:     "fallback_rewrite",
			allowFallback: true,
			wantStop:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := planner.Next(tt.in)
			if tt.wantStop {
				if got.Stop {
					return
				}
				if tt.allowFallback && got.Stage == tt.wantStage {
					return
				}
				t.Fatalf("planner.Next(...) = %+v, want stop or stage %q", got, tt.wantStage)
			}
			if got.Stop {
				t.Fatalf("planner.Next(...) stopped early: %+v", got)
			}
			if got.Stage != tt.wantStage {
				t.Fatalf("planner.Next(...) = %+v, want stage %q", got, tt.wantStage)
			}
		})
	}
}
