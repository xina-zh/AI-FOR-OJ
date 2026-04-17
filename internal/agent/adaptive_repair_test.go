package agent

import (
	"context"
	"strings"
	"testing"

	"ai-for-oj/internal/llm"
	"ai-for-oj/internal/model"
	"ai-for-oj/internal/prompt"
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
		name      string
		in        RepairPlanInput
		wantStage string
		wantStop  bool
	}{
		{
			name: "wrong answer after initial attempt",
			in: RepairPlanInput{
				AttemptCount:   1,
				LastFailure:    FailureTypeWrongAnswer,
				PreviousStages: nil,
				MaxBudget:      3,
			},
			wantStage: RepairStageWAAnalysisRepair,
		},
		{
			name: "runtime error routes to safety repair",
			in: RepairPlanInput{
				AttemptCount:   1,
				LastFailure:    FailureTypeRuntimeError,
				PreviousStages: nil,
				MaxBudget:      3,
			},
			wantStage: RepairStageRESafetyRepair,
		},
		{
			name: "time limit routes to complexity rewrite",
			in: RepairPlanInput{
				AttemptCount:   1,
				LastFailure:    FailureTypeTimeLimit,
				PreviousStages: nil,
				MaxBudget:      3,
			},
			wantStage: RepairStageTLEComplexityRewrite,
		},
		{
			name: "targeted stage already used falls back",
			in: RepairPlanInput{
				AttemptCount:   2,
				LastFailure:    FailureTypeWrongAnswer,
				PreviousStages: []string{RepairStageWAAnalysisRepair},
				MaxBudget:      3,
			},
			wantStage: RepairStageFallbackRewrite,
		},
		{
			name: "fallback stage already used stops",
			in: RepairPlanInput{
				AttemptCount:   2,
				LastFailure:    FailureTypeWrongAnswer,
				PreviousStages: []string{RepairStageWAAnalysisRepair, RepairStageFallbackRewrite},
				MaxBudget:      3,
			},
			wantStop: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := planner.Next(tt.in)
			if got.Stop {
				if tt.wantStop {
					return
				}
				t.Fatalf("planner.Next(...) stopped early: %+v", got)
			}
			if tt.wantStop {
				t.Fatalf("planner.Next(...) = %+v, want stop", got)
			}
			if got.Stage != tt.wantStage {
				t.Fatalf("planner.Next(...) = %+v, want stage %q", got, tt.wantStage)
			}
		})
	}
}

func TestAdaptiveRepairCoordinatorStopsAfterAC(t *testing.T) {
	client := &fakeSolveLLMClient{
		responses: []llm.GenerateResponse{
			{
				Model:        "solver-model",
				Content:      "```cpp\nint main() { return 0; }\n```",
				InputTokens:  11,
				OutputTokens: 7,
			},
		},
	}
	submitter := &fakeAdaptiveRepairJudgeSubmitter{
		outputs: []*JudgeFeedback{
			{
				Verdict:     "AC",
				PassedCount: 3,
				TotalCount:  3,
			},
		},
	}
	coord := NewAdaptiveRepairCoordinator(3)

	got, err := coord.Execute(context.Background(), client, SolveInput{
		Problem:        adaptiveRepairTestProblem(),
		Model:          "default-model",
		PromptName:     "default",
		JudgeSubmitter: submitter,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if got.AttemptCount != 1 {
		t.Fatalf("AttemptCount = %d, want 1", got.AttemptCount)
	}
	if len(got.StrategyPath) != 0 {
		t.Fatalf("StrategyPath = %+v, want empty", got.StrategyPath)
	}
	if len(got.Attempts) != 1 {
		t.Fatalf("Attempts length = %d, want 1", len(got.Attempts))
	}
	if len(client.requests) != 1 {
		t.Fatalf("LLM request count = %d, want 1", len(client.requests))
	}
	if len(submitter.submissions) != 1 {
		t.Fatalf("judge submission count = %d, want 1", len(submitter.submissions))
	}
}

func TestAdaptiveRepairStrategyFallsBackWithoutJudgeSubmitter(t *testing.T) {
	client := &fakeSolveLLMClient{
		responses: []llm.GenerateResponse{
			{
				Model:        "solver-model",
				Content:      "```cpp\nint main() { return 0; }\n```",
				InputTokens:  11,
				OutputTokens: 7,
			},
		},
	}

	input := SolveInput{
		Problem:    adaptiveRepairTestProblem(),
		Model:      "default-model",
		PromptName: prompt.StrictCPP17SolvePromptName,
	}

	got, err := adaptiveRepairStrategy{}.Execute(context.Background(), client, input)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if got.AgentName != AdaptiveRepairV1AgentName {
		t.Fatalf("AgentName = %q, want %q", got.AgentName, AdaptiveRepairV1AgentName)
	}
	if len(client.requests) != 1 {
		t.Fatalf("LLM request count = %d, want 1", len(client.requests))
	}
	directClient := &fakeSolveLLMClient{
		responses: []llm.GenerateResponse{
			{
				Model:        "solver-model",
				Content:      "```cpp\nint main() { return 0; }\n```",
				InputTokens:  11,
				OutputTokens: 7,
			},
		},
	}
	directGot, directErr := directCodegenStrategy{}.Execute(context.Background(), directClient, input)
	if directErr != nil {
		t.Fatalf("directCodegenStrategy.Execute returned error: %v", directErr)
	}
	if got.Model != directGot.Model ||
		got.PromptPreview != directGot.PromptPreview ||
		got.RawResponse != directGot.RawResponse ||
		got.TokenInput != directGot.TokenInput ||
		got.TokenOutput != directGot.TokenOutput {
		t.Fatalf("adaptive repair fallback output = %+v, want direct solve output %+v", got, directGot)
	}
	if got.LLMLatencyMS < 0 || directGot.LLMLatencyMS < 0 {
		t.Fatalf("unexpected negative latency: adaptive=%d direct=%d", got.LLMLatencyMS, directGot.LLMLatencyMS)
	}
	if len(directClient.requests) != 1 {
		t.Fatalf("direct solve LLM request count = %d, want 1", len(directClient.requests))
	}
	if client.requests[0].Prompt != directClient.requests[0].Prompt {
		t.Fatalf("fallback prompt = %q, want direct prompt %q", client.requests[0].Prompt, directClient.requests[0].Prompt)
	}
	if !strings.Contains(client.requests[0].Prompt, "PROMPT_TEMPLATE: strict_cpp17") {
		t.Fatalf("prompt = %q, want strict_cpp17 solve prompt", client.requests[0].Prompt)
	}
}

func TestAdaptiveRepairCoordinatorRoutesVerdictsToRepairStages(t *testing.T) {
	tests := []struct {
		name      string
		verdict   string
		wantStage string
		wantMark  string
	}{
		{
			name:      "wa verdict uses wa analysis repair",
			verdict:   "WA",
			wantStage: RepairStageWAAnalysisRepair,
			wantMark:  "PROMPT_TEMPLATE: repair_wa",
		},
		{
			name:      "re verdict uses safety repair",
			verdict:   "RE",
			wantStage: RepairStageRESafetyRepair,
			wantMark:  "PROMPT_TEMPLATE: repair_re",
		},
		{
			name:      "tle verdict uses complexity rewrite",
			verdict:   "TLE",
			wantStage: RepairStageTLEComplexityRewrite,
			wantMark:  "PROMPT_TEMPLATE: repair_tle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakeSolveLLMClient{
				responses: []llm.GenerateResponse{
					{
						Model:        "solver-model",
						Content:      "```cpp\nint main() { return 1; }\n```",
						InputTokens:  11,
						OutputTokens: 7,
					},
					{
						Model:        "solver-model",
						Content:      "```cpp\nint main() { return 0; }\n```",
						InputTokens:  13,
						OutputTokens: 9,
					},
				},
			}
			submitter := &fakeAdaptiveRepairJudgeSubmitter{
				outputs: []*JudgeFeedback{
					{
						Verdict:     tt.verdict,
						PassedCount: 1,
						TotalCount:  3,
					},
					{
						Verdict:     "AC",
						PassedCount: 3,
						TotalCount:  3,
					},
				},
			}
			coord := NewAdaptiveRepairCoordinator(3)

			got, err := coord.Execute(context.Background(), client, SolveInput{
				Problem:        adaptiveRepairTestProblem(),
				Model:          "default-model",
				PromptName:     "default",
				JudgeSubmitter: submitter,
			})
			if err != nil {
				t.Fatalf("Execute returned error: %v", err)
			}
			if got.AttemptCount != 2 {
				t.Fatalf("AttemptCount = %d, want 2", got.AttemptCount)
			}
			if len(got.StrategyPath) != 1 || got.StrategyPath[0] != tt.wantStage {
				t.Fatalf("StrategyPath = %+v, want [%s]", got.StrategyPath, tt.wantStage)
			}
			if len(got.Attempts) != 2 {
				t.Fatalf("Attempts length = %d, want 2", len(got.Attempts))
			}
			if got.Attempts[1].Stage != tt.wantStage {
				t.Fatalf("second attempt stage = %q, want %q", got.Attempts[1].Stage, tt.wantStage)
			}
			if !strings.Contains(client.requests[1].Prompt, tt.wantMark) {
				t.Fatalf("repair prompt = %q, want marker %q", client.requests[1].Prompt, tt.wantMark)
			}
		})
	}
}

func TestAdaptiveRepairCoordinatorAccumulatesAttemptMetadata(t *testing.T) {
	client := &fakeSolveLLMClient{
		responses: []llm.GenerateResponse{
			{
				Model:        "solver-model",
				Content:      "```cpp\nint main() { return 1; }\n```",
				InputTokens:  10,
				OutputTokens: 5,
			},
			{
				Model:        "solver-model",
				Content:      "```cpp\nint main() { return 2; }\n```",
				InputTokens:  11,
				OutputTokens: 6,
			},
			{
				Model:        "solver-model",
				Content:      "```cpp\nint main() { return 3; }\n```",
				InputTokens:  12,
				OutputTokens: 7,
			},
			{
				Model:        "solver-model",
				Content:      "```cpp\nint main() { return 0; }\n```",
				InputTokens:  13,
				OutputTokens: 8,
			},
		},
	}
	submitter := &fakeAdaptiveRepairJudgeSubmitter{
		outputs: []*JudgeFeedback{
			{Verdict: "WA", PassedCount: 1, TotalCount: 3},
			{Verdict: "RE", PassedCount: 1, TotalCount: 3},
			{Verdict: "TLE", PassedCount: 1, TotalCount: 3},
			{Verdict: "AC", PassedCount: 3, TotalCount: 3},
		},
	}
	coord := NewAdaptiveRepairCoordinator(4)

	got, err := coord.Execute(context.Background(), client, SolveInput{
		Problem:        adaptiveRepairTestProblem(),
		Model:          "default-model",
		PromptName:     "default",
		JudgeSubmitter: submitter,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if got.AttemptCount != 4 {
		t.Fatalf("AttemptCount = %d, want 4", got.AttemptCount)
	}
	if len(got.StrategyPath) != 3 {
		t.Fatalf("StrategyPath length = %d, want 3", len(got.StrategyPath))
	}
	wantPath := []string{
		RepairStageWAAnalysisRepair,
		RepairStageRESafetyRepair,
		RepairStageTLEComplexityRewrite,
	}
	for i, want := range wantPath {
		if got.StrategyPath[i] != want {
			t.Fatalf("StrategyPath[%d] = %q, want %q", i, got.StrategyPath[i], want)
		}
	}
	if len(got.Attempts) != 4 {
		t.Fatalf("Attempts length = %d, want 4", len(got.Attempts))
	}
	for i, want := range append([]string{"initial_solve"}, wantPath...) {
		if got.Attempts[i].Stage != want {
			t.Fatalf("Attempts[%d].Stage = %q, want %q", i, got.Attempts[i].Stage, want)
		}
	}
	if got.Attempts[0].FailureType != FailureTypeWrongAnswer {
		t.Fatalf("Attempts[0].FailureType = %q, want %q", got.Attempts[0].FailureType, FailureTypeWrongAnswer)
	}
	if got.Attempts[1].FailureType != FailureTypeRuntimeError {
		t.Fatalf("Attempts[1].FailureType = %q, want %q", got.Attempts[1].FailureType, FailureTypeRuntimeError)
	}
	if got.Attempts[2].FailureType != FailureTypeTimeLimit {
		t.Fatalf("Attempts[2].FailureType = %q, want %q", got.Attempts[2].FailureType, FailureTypeTimeLimit)
	}
}

type fakeAdaptiveRepairJudgeSubmitter struct {
	submissions []string
	outputs     []*JudgeFeedback
	errs        []error
}

func (s *fakeAdaptiveRepairJudgeSubmitter) Submit(_ context.Context, sourceCode string) (*JudgeFeedback, error) {
	s.submissions = append(s.submissions, sourceCode)
	index := len(s.submissions) - 1
	if index < len(s.errs) && s.errs[index] != nil {
		return nil, s.errs[index]
	}
	if index < len(s.outputs) && s.outputs[index] != nil {
		return s.outputs[index], nil
	}
	return &JudgeFeedback{Verdict: "AC"}, nil
}

func adaptiveRepairTestProblem() *model.Problem {
	return &model.Problem{
		Title:       "Echo",
		Description: "echo input",
		InputSpec:   "one line",
		OutputSpec:  "same line",
		Samples:     "[]",
	}
}
