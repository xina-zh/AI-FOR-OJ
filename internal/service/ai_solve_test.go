package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"ai-for-oj/internal/agent"
	"ai-for-oj/internal/llm"
	"ai-for-oj/internal/model"
	"ai-for-oj/internal/prompt"
	"ai-for-oj/internal/repository"
)

type fakeAISolveRunRepository struct {
	created           []*model.AISolveRun
	updated           []*model.AISolveRun
	getRun            *model.AISolveRun
	err               error
	nextID            uint
	rejectCanceledCtx bool
}

func (r *fakeAISolveRunRepository) Create(_ context.Context, run *model.AISolveRun) error {
	if r.err != nil {
		return r.err
	}
	if r.nextID == 0 {
		r.nextID = 1
	}
	run.ID = r.nextID
	copied := *run
	r.created = append(r.created, &copied)
	return nil
}

func (r *fakeAISolveRunRepository) Update(ctx context.Context, run *model.AISolveRun) error {
	if r.err != nil {
		return r.err
	}
	if r.rejectCanceledCtx && ctx != nil && ctx.Err() != nil {
		return ctx.Err()
	}
	copied := *run
	r.updated = append(r.updated, &copied)
	return nil
}

func (r *fakeAISolveRunRepository) GetByID(_ context.Context, runID uint) (*model.AISolveRun, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.getRun == nil || r.getRun.ID != runID {
		return nil, repository.ErrAISolveRunNotFound
	}
	return r.getRun, nil
}

type fakeLLMClient struct {
	response  llm.GenerateResponse
	err       error
	request   llm.GenerateRequest
	requests  []llm.GenerateRequest
	delay     time.Duration
	responses []llm.GenerateResponse
	errors    []error
	delays    []time.Duration
}

func (c *fakeLLMClient) Generate(_ context.Context, req llm.GenerateRequest) (llm.GenerateResponse, error) {
	c.request = req
	c.requests = append(c.requests, req)

	index := len(c.requests) - 1
	if index < len(c.delays) && c.delays[index] > 0 {
		time.Sleep(c.delays[index])
	} else if c.delay > 0 {
		time.Sleep(c.delay)
	}
	if index < len(c.errors) && c.errors[index] != nil {
		return llm.GenerateResponse{}, c.errors[index]
	}
	if index < len(c.responses) {
		return c.responses[index], nil
	}
	return c.response, c.err
}

type fakeJudgeSubmitter struct {
	input   JudgeSubmissionInput
	inputs  []JudgeSubmissionInput
	output  *JudgeSubmissionOutput
	outputs []*JudgeSubmissionOutput
	err     error
	errs    []error
}

func (s *fakeJudgeSubmitter) Submit(_ context.Context, input JudgeSubmissionInput) (*JudgeSubmissionOutput, error) {
	s.input = input
	s.inputs = append(s.inputs, input)
	index := len(s.inputs) - 1
	if index < len(s.errs) && s.errs[index] != nil {
		return nil, s.errs[index]
	}
	if index < len(s.outputs) && s.outputs[index] != nil {
		return s.outputs[index], nil
	}
	return s.output, s.err
}

func TestAISolveServiceSolve(t *testing.T) {
	llmClient := &fakeLLMClient{
		response: llm.GenerateResponse{
			Model:        "mock-cpp17",
			Content:      "```cpp\n#include <bits/stdc++.h>\nusing namespace std;\nint main(){cout<<1;return 0;}\n```",
			InputTokens:  123,
			OutputTokens: 45,
		},
		delay: 2 * time.Millisecond,
	}
	judgeSubmitter := &fakeJudgeSubmitter{
		output: &JudgeSubmissionOutput{
			SubmissionID: 10,
			ProblemID:    1,
			SourceType:   model.SourceTypeAI,
			Verdict:      "AC",
		},
	}
	runRepo := &fakeAISolveRunRepository{}
	service := NewAISolveService(
		fakeProblemRepository{
			problem: &model.Problem{
				BaseModel:     model.BaseModel{ID: 1},
				Title:         "A + B",
				Description:   "sum two integers",
				InputSpec:     "two ints",
				OutputSpec:    "one int",
				Samples:       `[{"input":"1 2","output":"3"}]`,
				TimeLimitMS:   1000,
				MemoryLimitMB: 256,
			},
		},
		runRepo,
		llmClient,
		judgeSubmitter,
		"default-model",
	)

	output, err := service.Solve(context.Background(), AISolveInput{ProblemID: 1})
	if err != nil {
		t.Fatalf("solve returned error: %v", err)
	}

	if !strings.Contains(llmClient.request.Prompt, "A + B") {
		t.Fatalf("expected prompt to include problem title, got %q", llmClient.request.Prompt)
	}
	if !strings.Contains(llmClient.request.Prompt, "PROMPT_TEMPLATE: default") {
		t.Fatalf("expected default prompt marker, got %q", llmClient.request.Prompt)
	}
	if !strings.Contains(llmClient.request.Prompt, "Do not include explanation outside the code unless necessary.") {
		t.Fatalf("expected default prompt template to be used, got %q", llmClient.request.Prompt)
	}
	if llmClient.request.Model != "default-model" {
		t.Fatalf("expected default model to be passed to llm client, got %q", llmClient.request.Model)
	}
	if output.PromptName != prompt.DefaultSolvePromptName {
		t.Fatalf("expected default prompt name in output, got %q", output.PromptName)
	}
	if output.AgentName != agent.DirectCodegenAgentName {
		t.Fatalf("expected default agent name in output, got %q", output.AgentName)
	}

	if judgeSubmitter.input.SourceType != model.SourceTypeAI {
		t.Fatalf("expected source type %s, got %s", model.SourceTypeAI, judgeSubmitter.input.SourceType)
	}

	if judgeSubmitter.input.Language != model.LanguageCPP17 {
		t.Fatalf("expected language %s, got %s", model.LanguageCPP17, judgeSubmitter.input.Language)
	}

	if !strings.Contains(output.ExtractedCode, "int main()") {
		t.Fatalf("expected extracted code in output, got %q", output.ExtractedCode)
	}

	if output.SubmissionID != 10 || output.Verdict != "AC" {
		t.Fatalf("expected submission result in output, got %+v", output)
	}
	if output.TokenInput != 123 || output.TokenOutput != 45 {
		t.Fatalf("expected token usage in output, got %+v", output)
	}
	if output.LLMLatencyMS < 0 || output.TotalLatencyMS < output.LLMLatencyMS {
		t.Fatalf("expected latency fields in output, got %+v", output)
	}

	if output.AISolveRunID == 0 {
		t.Fatal("expected ai solve run id to be returned")
	}

	if len(runRepo.updated) != 1 || runRepo.updated[0].Status != model.AISolveRunStatusSuccess {
		t.Fatalf("expected successful run to be persisted, got %+v", runRepo.updated)
	}
	if runRepo.updated[0].PromptName != prompt.DefaultSolvePromptName {
		t.Fatalf("expected default prompt name to be persisted, got %+v", runRepo.updated[0])
	}
	if runRepo.updated[0].AgentName != agent.DirectCodegenAgentName {
		t.Fatalf("expected default agent name to be persisted, got %+v", runRepo.updated[0])
	}
	if runRepo.updated[0].TokenInput != 123 || runRepo.updated[0].TokenOutput != 45 {
		t.Fatalf("expected token usage persisted, got %+v", runRepo.updated[0])
	}
	if runRepo.updated[0].TotalLatencyMS < runRepo.updated[0].LLMLatencyMS {
		t.Fatalf("expected latency persisted, got %+v", runRepo.updated[0])
	}
}

func TestAISolveServiceSolvePrefersRequestModel(t *testing.T) {
	llmClient := &fakeLLMClient{
		response: llm.GenerateResponse{
			Model:   "request-model",
			Content: "```cpp\nint main(){return 0;}\n```",
		},
	}
	judgeSubmitter := &fakeJudgeSubmitter{
		output: &JudgeSubmissionOutput{SubmissionID: 10, ProblemID: 1, SourceType: model.SourceTypeAI, Verdict: "AC"},
	}
	service := NewAISolveService(
		fakeProblemRepository{
			problem: &model.Problem{
				BaseModel:   model.BaseModel{ID: 1},
				Title:       "Echo",
				Description: "echo input",
				InputSpec:   "line",
				OutputSpec:  "line",
				Samples:     "[]",
			},
		},
		&fakeAISolveRunRepository{},
		llmClient,
		judgeSubmitter,
		"default-model",
	)

	output, err := service.Solve(context.Background(), AISolveInput{
		ProblemID:  1,
		Model:      "request-model",
		PromptName: prompt.StrictCPP17SolvePromptName,
		AgentName:  agent.DirectCodegenAgentName,
	})
	if err != nil {
		t.Fatalf("solve returned error: %v", err)
	}
	if llmClient.request.Model != "request-model" {
		t.Fatalf("expected request model to be passed to llm client, got %q", llmClient.request.Model)
	}
	if output.Model != "request-model" {
		t.Fatalf("expected output model to reflect request model, got %q", output.Model)
	}
	if output.PromptName != prompt.StrictCPP17SolvePromptName {
		t.Fatalf("expected output prompt name to reflect request prompt, got %q", output.PromptName)
	}
	if output.AgentName != agent.DirectCodegenAgentName {
		t.Fatalf("expected output agent name to reflect request agent, got %q", output.AgentName)
	}
	if !strings.Contains(output.PromptPreview, "PROMPT_TEMPLATE: strict_cpp17") {
		t.Fatalf("expected output prompt preview to reflect strict prompt, got %q", output.PromptPreview)
	}
	if !strings.Contains(llmClient.request.Prompt, "Return exactly one markdown code block with language tag cpp.") {
		t.Fatalf("expected strict prompt template to be used, got %q", llmClient.request.Prompt)
	}
	if !strings.Contains(llmClient.request.Prompt, "PROMPT_TEMPLATE: strict_cpp17") {
		t.Fatalf("expected strict prompt marker, got %q", llmClient.request.Prompt)
	}
}

func TestAISolveServiceSolveReturnsProblemNotFound(t *testing.T) {
	runRepo := &fakeAISolveRunRepository{}
	service := NewAISolveService(
		fakeProblemRepository{err: repository.ErrProblemNotFound},
		runRepo,
		&fakeLLMClient{},
		&fakeJudgeSubmitter{},
		"default-model",
	)

	output, err := service.Solve(context.Background(), AISolveInput{ProblemID: 999})
	if err != repository.ErrProblemNotFound {
		t.Fatalf("expected err %v, got %v", repository.ErrProblemNotFound, err)
	}
	if output == nil || output.AISolveRunID == 0 {
		t.Fatalf("expected failed solve to return run id, got %+v", output)
	}
	if len(runRepo.updated) != 1 || runRepo.updated[0].Status != model.AISolveRunStatusFailed {
		t.Fatalf("expected failed run to be persisted, got %+v", runRepo.updated)
	}
}

func TestAISolveServiceSolveReturnsLLMFailure(t *testing.T) {
	runRepo := &fakeAISolveRunRepository{}
	service := NewAISolveService(
		fakeProblemRepository{
			problem: &model.Problem{
				BaseModel:   model.BaseModel{ID: 1},
				Title:       "Echo",
				Description: "echo input",
				InputSpec:   "line",
				OutputSpec:  "same line",
				Samples:     "[]",
			},
		},
		runRepo,
		&fakeLLMClient{err: errors.New("upstream timeout")},
		&fakeJudgeSubmitter{},
		"default-model",
	)

	output, err := service.Solve(context.Background(), AISolveInput{ProblemID: 1})
	if !errors.Is(err, ErrAISolveLLMFailed) {
		t.Fatalf("expected err %v, got %v", ErrAISolveLLMFailed, err)
	}
	if output == nil || output.AISolveRunID == 0 {
		t.Fatalf("expected failed solve to return run id, got %+v", output)
	}
	if len(runRepo.updated) != 1 || runRepo.updated[0].ErrorMessage == "" {
		t.Fatalf("expected llm failure to be persisted, got %+v", runRepo.updated)
	}
	if runRepo.updated[0].TotalLatencyMS < 0 {
		t.Fatalf("expected total latency on failure, got %+v", runRepo.updated[0])
	}
	if output.PromptPreview == "" || runRepo.updated[0].PromptPreview == "" {
		t.Fatalf("expected llm failure to retain prompt preview, got output=%+v run=%+v", output, runRepo.updated[0])
	}
}

func TestAISolveServiceSolveRejectsUnknownPromptName(t *testing.T) {
	service := NewAISolveService(
		fakeProblemRepository{},
		&fakeAISolveRunRepository{},
		&fakeLLMClient{},
		&fakeJudgeSubmitter{},
		"default-model",
	)

	output, err := service.Solve(context.Background(), AISolveInput{
		ProblemID:  1,
		PromptName: "not-exists",
	})
	if !errors.Is(err, prompt.ErrUnknownSolvePrompt) {
		t.Fatalf("expected err %v, got %v", prompt.ErrUnknownSolvePrompt, err)
	}
	if output != nil {
		t.Fatalf("expected nil output on invalid prompt name, got %+v", output)
	}
}

func TestAISolveServiceSolveRejectsUnknownAgentName(t *testing.T) {
	service := NewAISolveService(
		fakeProblemRepository{},
		&fakeAISolveRunRepository{},
		&fakeLLMClient{},
		&fakeJudgeSubmitter{},
		"default-model",
	)

	output, err := service.Solve(context.Background(), AISolveInput{
		ProblemID: 1,
		AgentName: "not-exists",
	})
	if !errors.Is(err, agent.ErrUnknownSolveAgent) {
		t.Fatalf("expected err %v, got %v", agent.ErrUnknownSolveAgent, err)
	}
	if output != nil {
		t.Fatalf("expected nil output on invalid agent name, got %+v", output)
	}
}

func TestAISolveServiceSolveReturnsCodeExtractionFailure(t *testing.T) {
	runRepo := &fakeAISolveRunRepository{}
	service := NewAISolveService(
		fakeProblemRepository{
			problem: &model.Problem{
				BaseModel:   model.BaseModel{ID: 1},
				Title:       "Echo",
				Description: "echo input",
				InputSpec:   "line",
				OutputSpec:  "same line",
				Samples:     "[]",
			},
		},
		runRepo,
		&fakeLLMClient{response: llm.GenerateResponse{Model: "mock", Content: "   "}},
		&fakeJudgeSubmitter{},
		"default-model",
	)

	output, err := service.Solve(context.Background(), AISolveInput{ProblemID: 1})
	if !errors.Is(err, ErrAISolveCodeNotExtracted) {
		t.Fatalf("expected err %v, got %v", ErrAISolveCodeNotExtracted, err)
	}
	if output == nil || output.AISolveRunID == 0 {
		t.Fatalf("expected failed solve to return run id, got %+v", output)
	}
	if len(runRepo.updated) != 1 || runRepo.updated[0].Status != model.AISolveRunStatusFailed {
		t.Fatalf("expected extraction failure to be persisted, got %+v", runRepo.updated)
	}
}

func TestAISolveServiceSolveAdaptiveRepairRejectsMissingCPPBeforeJudge(t *testing.T) {
	runRepo := &fakeAISolveRunRepository{}
	judgeSubmitter := &fakeJudgeSubmitter{}
	service := NewAISolveService(
		fakeProblemRepository{problem: adaptiveServiceTestProblem()},
		runRepo,
		&fakeLLMClient{response: llm.GenerateResponse{Model: "mock-cpp17", Content: "I cannot solve this problem."}},
		judgeSubmitter,
		"default-model",
	)

	output, err := service.Solve(context.Background(), AISolveInput{
		ProblemID:  1,
		AgentName:  agent.DirectCodegenRepairAgentName,
		PromptName: prompt.DefaultSolvePromptName,
	})
	if !errors.Is(err, ErrAISolveCodeNotExtracted) {
		t.Fatalf("expected err %v, got %v", ErrAISolveCodeNotExtracted, err)
	}
	if output == nil || output.AISolveRunID == 0 {
		t.Fatalf("expected failed solve to return run id, got %+v", output)
	}
	if len(judgeSubmitter.inputs) != 0 {
		t.Fatalf("expected no judge submission for missing cpp, got %d", len(judgeSubmitter.inputs))
	}
	if len(runRepo.created) != 1 {
		t.Fatalf("expected one run creation, got %d", len(runRepo.created))
	}
	if len(runRepo.updated) != 1 || runRepo.updated[0].Status != model.AISolveRunStatusFailed {
		t.Fatalf("expected failed run to be persisted, got %+v", runRepo.updated)
	}
}

func adaptiveServiceTestProblem() *model.Problem {
	return &model.Problem{
		BaseModel:   model.BaseModel{ID: 1},
		Title:       "Echo",
		Description: "echo input",
		InputSpec:   "line",
		OutputSpec:  "line",
		Samples:     "[]",
	}
}

func TestAISolveServiceGetRun(t *testing.T) {
	runRepo := &fakeAISolveRunRepository{
		getRun: &model.AISolveRun{
			BaseModel:      model.BaseModel{ID: 5},
			ProblemID:      1,
			Model:          "mock-cpp17",
			PromptName:     prompt.DefaultSolvePromptName,
			AgentName:      agent.DirectCodegenAgentName,
			Status:         model.AISolveRunStatusSuccess,
			ErrorMessage:   "",
			TokenInput:     12,
			TokenOutput:    34,
			LLMLatencyMS:   5,
			TotalLatencyMS: 9,
		},
	}
	service := NewAISolveService(
		fakeProblemRepository{},
		runRepo,
		&fakeLLMClient{},
		&fakeJudgeSubmitter{},
		"default-model",
	)

	run, err := service.GetRun(context.Background(), 5)
	if err != nil {
		t.Fatalf("get run returned error: %v", err)
	}
	if run.ID != 5 || run.Status != model.AISolveRunStatusSuccess {
		t.Fatalf("unexpected run: %+v", run)
	}
	if run.TokenInput != 12 || run.TokenOutput != 34 || run.LLMLatencyMS != 5 || run.TotalLatencyMS != 9 {
		t.Fatalf("unexpected run cost fields: %+v", run)
	}
}

func TestAISolveServiceSolveAnalyzeThenCodegenAggregatesCost(t *testing.T) {
	llmClient := &fakeLLMClient{
		responses: []llm.GenerateResponse{
			{
				Model:        "gpt-5.4",
				Content:      "Problem understanding: ...\nAlgorithm idea: ...\nBoundary cases: ...",
				InputTokens:  40,
				OutputTokens: 20,
			},
			{
				Model:        "gpt-5.4",
				Content:      "```cpp\nint main(){return 0;}\n```",
				InputTokens:  60,
				OutputTokens: 30,
			},
		},
		delays: []time.Duration{2 * time.Millisecond, 3 * time.Millisecond},
	}
	judgeSubmitter := &fakeJudgeSubmitter{
		output: &JudgeSubmissionOutput{SubmissionID: 10, ProblemID: 1, SourceType: model.SourceTypeAI, Verdict: "AC"},
	}
	runRepo := &fakeAISolveRunRepository{}
	service := NewAISolveService(
		fakeProblemRepository{
			problem: &model.Problem{
				BaseModel:   model.BaseModel{ID: 1},
				Title:       "Echo",
				Description: "echo input",
				InputSpec:   "line",
				OutputSpec:  "line",
				Samples:     "[]",
			},
		},
		runRepo,
		llmClient,
		judgeSubmitter,
		"default-model",
	)

	output, err := service.Solve(context.Background(), AISolveInput{
		ProblemID:  1,
		Model:      "gpt-5.4",
		PromptName: prompt.StrictCPP17SolvePromptName,
		AgentName:  agent.AnalyzeThenCodegenAgentName,
	})
	if err != nil {
		t.Fatalf("solve returned error: %v", err)
	}
	if len(llmClient.requests) != 2 {
		t.Fatalf("expected two llm calls, got %d", len(llmClient.requests))
	}
	if !strings.Contains(llmClient.requests[0].Prompt, "First produce a concise analysis instead of code.") {
		t.Fatalf("expected first call to be analysis prompt, got %q", llmClient.requests[0].Prompt)
	}
	if !strings.Contains(llmClient.requests[1].Prompt, "Prior Analysis:") {
		t.Fatalf("expected second call to include prior analysis, got %q", llmClient.requests[1].Prompt)
	}
	if output.AgentName != agent.AnalyzeThenCodegenAgentName {
		t.Fatalf("expected analyze_then_codegen agent in output, got %q", output.AgentName)
	}
	if output.TokenInput != 100 || output.TokenOutput != 50 {
		t.Fatalf("expected aggregated token usage, got %+v", output)
	}
	if output.LLMLatencyMS < 5 {
		t.Fatalf("expected aggregated llm latency, got %+v", output)
	}
	if len(runRepo.updated) != 1 || runRepo.updated[0].AgentName != agent.AnalyzeThenCodegenAgentName {
		t.Fatalf("expected analyze_then_codegen to be persisted, got %+v", runRepo.updated)
	}
	if runRepo.updated[0].TokenInput != 100 || runRepo.updated[0].TokenOutput != 50 {
		t.Fatalf("expected aggregated token usage persisted, got %+v", runRepo.updated[0])
	}
}

func TestAISolveServiceSolveRepairsAfterFirstFailure(t *testing.T) {
	llmClient := &fakeLLMClient{
		responses: []llm.GenerateResponse{
			{
				Model:        "gpt-5.4",
				Content:      "```cpp\nint main(){return 1;}\n```",
				InputTokens:  10,
				OutputTokens: 20,
			},
			{
				Model:        "gpt-5.4",
				Content:      "```cpp\nint main(){return 0;}\n```",
				InputTokens:  30,
				OutputTokens: 40,
			},
		},
	}
	judgeSubmitter := &fakeJudgeSubmitter{
		outputs: []*JudgeSubmissionOutput{
			{
				SubmissionID: 11,
				ProblemID:    1,
				SourceType:   model.SourceTypeAI,
				Verdict:      "WA",
				ErrorMessage: "wrong answer on sample",
				PassedCount:  0,
				TotalCount:   2,
				TestCaseResults: []JudgeSubmissionCaseFeedback{
					{CaseIndex: 1, Verdict: "WA", Stdout: "1", Stderr: "", ExitCode: 0},
				},
			},
			{
				SubmissionID: 12,
				ProblemID:    1,
				SourceType:   model.SourceTypeAI,
				Verdict:      "AC",
				PassedCount:  2,
				TotalCount:   2,
			},
		},
	}
	runRepo := &fakeAISolveRunRepository{}
	service := NewAISolveService(
		fakeProblemRepository{
			problem: &model.Problem{
				BaseModel:   model.BaseModel{ID: 1},
				Title:       "Echo",
				Description: "echo input",
				InputSpec:   "line",
				OutputSpec:  "line",
				Samples:     "[]",
			},
		},
		runRepo,
		llmClient,
		judgeSubmitter,
		"default-model",
	)

	output, err := service.Solve(context.Background(), AISolveInput{
		ProblemID: 1,
		AgentName: agent.DirectCodegenRepairAgentName,
	})
	if err != nil {
		t.Fatalf("solve returned error: %v", err)
	}
	if output.Verdict != "AC" || output.SubmissionID != 12 {
		t.Fatalf("expected repaired solve to end in AC, got %+v", output)
	}
	if len(llmClient.requests) != 2 {
		t.Fatalf("expected initial generation plus one repair, got %d llm calls", len(llmClient.requests))
	}
	if len(judgeSubmitter.inputs) != 2 {
		t.Fatalf("expected two judge submissions, got %d", len(judgeSubmitter.inputs))
	}
	if !strings.Contains(llmClient.requests[1].Prompt, "Your previous submission failed.") {
		t.Fatalf("expected repair prompt marker, got %q", llmClient.requests[1].Prompt)
	}
	if !strings.Contains(llmClient.requests[1].Prompt, "wrong answer on sample") {
		t.Fatalf("expected repair prompt to include judge feedback, got %q", llmClient.requests[1].Prompt)
	}
	if output.TokenInput != 40 || output.TokenOutput != 60 {
		t.Fatalf("expected aggregated llm tokens across repair attempts, got %+v", output)
	}
	if len(runRepo.updated) != 1 || runRepo.updated[0].Verdict != "AC" {
		t.Fatalf("expected final run update to be AC, got %+v", runRepo.updated)
	}
}

func TestAISolveServiceSolveStopsAfterMaxRepairAttempts(t *testing.T) {
	llmClient := &fakeLLMClient{
		responses: []llm.GenerateResponse{
			{Model: "gpt-5.4", Content: "```cpp\nint main(){return 1;}\n```", InputTokens: 10, OutputTokens: 10},
			{Model: "gpt-5.4", Content: "```cpp\nint main(){return 2;}\n```", InputTokens: 20, OutputTokens: 20},
			{Model: "gpt-5.4", Content: "```cpp\nint main(){return 3;}\n```", InputTokens: 30, OutputTokens: 30},
		},
	}
	judgeSubmitter := &fakeJudgeSubmitter{
		outputs: []*JudgeSubmissionOutput{
			{SubmissionID: 11, ProblemID: 1, SourceType: model.SourceTypeAI, Verdict: "WA", PassedCount: 0, TotalCount: 3, ErrorMessage: "still wrong"},
			{SubmissionID: 12, ProblemID: 1, SourceType: model.SourceTypeAI, Verdict: "RE", PassedCount: 1, TotalCount: 3, ErrorMessage: "runtime error", RunStderr: "segmentation fault"},
			{SubmissionID: 13, ProblemID: 1, SourceType: model.SourceTypeAI, Verdict: "TLE", PassedCount: 2, TotalCount: 3, ErrorMessage: "time limit exceeded", TimedOut: true},
		},
	}
	runRepo := &fakeAISolveRunRepository{}
	service := NewAISolveService(
		fakeProblemRepository{
			problem: &model.Problem{
				BaseModel:   model.BaseModel{ID: 1},
				Title:       "Echo",
				Description: "echo input",
				InputSpec:   "line",
				OutputSpec:  "line",
				Samples:     "[]",
			},
		},
		runRepo,
		llmClient,
		judgeSubmitter,
		"default-model",
	)

	output, err := service.Solve(context.Background(), AISolveInput{
		ProblemID: 1,
		AgentName: agent.DirectCodegenRepairAgentName,
	})
	if err != nil {
		t.Fatalf("solve returned error: %v", err)
	}
	if len(llmClient.requests) != 3 || len(judgeSubmitter.inputs) != 3 {
		t.Fatalf("expected at most three total attempts, got llm=%d judge=%d", len(llmClient.requests), len(judgeSubmitter.inputs))
	}
	if output.Verdict != "TLE" || output.SubmissionID != 13 {
		t.Fatalf("expected final result from third attempt, got %+v", output)
	}
	if output.TokenInput != 60 || output.TokenOutput != 60 {
		t.Fatalf("expected aggregated token usage across three attempts, got %+v", output)
	}
	if len(runRepo.updated) != 1 || runRepo.updated[0].Verdict != "TLE" {
		t.Fatalf("expected final persisted verdict to be last attempt result, got %+v", runRepo.updated)
	}
}

func TestAISolveServiceSolveRepairPromptIncludesCompileFailureFeedback(t *testing.T) {
	llmClient := &fakeLLMClient{
		responses: []llm.GenerateResponse{
			{Model: "gpt-5.4", Content: "```cpp\nint main( { return 0; }\n```"},
			{Model: "gpt-5.4", Content: "```cpp\nint main(){return 0;}\n```"},
		},
	}
	judgeSubmitter := &fakeJudgeSubmitter{
		outputs: []*JudgeSubmissionOutput{
			{
				SubmissionID:  21,
				ProblemID:     1,
				SourceType:    model.SourceTypeAI,
				Verdict:       "CE",
				ErrorMessage:  "compile failed",
				CompileStderr: "error: expected ')' before '{' token",
				ExecStage:     "compile",
			},
			{
				SubmissionID: 22,
				ProblemID:    1,
				SourceType:   model.SourceTypeAI,
				Verdict:      "AC",
			},
		},
	}
	service := NewAISolveService(
		fakeProblemRepository{
			problem: &model.Problem{
				BaseModel:   model.BaseModel{ID: 1},
				Title:       "Compile Fix",
				Description: "fix syntax",
				InputSpec:   "input",
				OutputSpec:  "output",
				Samples:     "[]",
			},
		},
		&fakeAISolveRunRepository{},
		llmClient,
		judgeSubmitter,
		"default-model",
	)

	output, err := service.Solve(context.Background(), AISolveInput{
		ProblemID: 1,
		AgentName: agent.DirectCodegenRepairAgentName,
	})
	if err != nil {
		t.Fatalf("solve returned error: %v", err)
	}
	if output.Verdict != "AC" {
		t.Fatalf("expected compile error repair to end in AC, got %+v", output)
	}
	if len(llmClient.requests) != 2 {
		t.Fatalf("expected one repair after compile error, got %d calls", len(llmClient.requests))
	}
	if !strings.Contains(llmClient.requests[1].Prompt, "Compile Stderr:") ||
		!strings.Contains(llmClient.requests[1].Prompt, "expected ')' before '{' token") {
		t.Fatalf("expected compile stderr in repair prompt, got %q", llmClient.requests[1].Prompt)
	}
}

func TestAISolveServiceSolveDirectCodegenDoesNotRepair(t *testing.T) {
	llmClient := &fakeLLMClient{
		responses: []llm.GenerateResponse{
			{
				Model:        "gpt-5.4",
				Content:      "```cpp\nint main(){return 1;}\n```",
				InputTokens:  10,
				OutputTokens: 20,
			},
			{
				Model:        "gpt-5.4",
				Content:      "```cpp\nint main(){return 0;}\n```",
				InputTokens:  30,
				OutputTokens: 40,
			},
		},
	}
	judgeSubmitter := &fakeJudgeSubmitter{
		outputs: []*JudgeSubmissionOutput{
			{
				SubmissionID: 11,
				ProblemID:    1,
				SourceType:   model.SourceTypeAI,
				Verdict:      "WA",
				ErrorMessage: "wrong answer on sample",
				PassedCount:  0,
				TotalCount:   2,
			},
		},
	}
	runRepo := &fakeAISolveRunRepository{}
	service := NewAISolveService(
		fakeProblemRepository{
			problem: &model.Problem{
				BaseModel:   model.BaseModel{ID: 1},
				Title:       "Echo",
				Description: "echo input",
				InputSpec:   "line",
				OutputSpec:  "line",
				Samples:     "[]",
			},
		},
		runRepo,
		llmClient,
		judgeSubmitter,
		"default-model",
	)

	output, err := service.Solve(context.Background(), AISolveInput{
		ProblemID: 1,
		AgentName: agent.DirectCodegenAgentName,
	})
	if err != nil {
		t.Fatalf("solve returned error: %v", err)
	}
	if output.Verdict != "WA" || output.SubmissionID != 11 {
		t.Fatalf("expected direct_codegen to stop after first failed attempt, got %+v", output)
	}
	if len(llmClient.requests) != 1 {
		t.Fatalf("expected direct_codegen to avoid repair retries, got %d llm calls", len(llmClient.requests))
	}
	if len(judgeSubmitter.inputs) != 1 {
		t.Fatalf("expected only one judge attempt, got %d", len(judgeSubmitter.inputs))
	}
}

func TestTruncateForPreviewKeepsUTF8Valid(t *testing.T) {
	value := "中文题目预览abc"
	got := truncateForPreview(value, 3)
	if !utf8.ValidString(got) {
		t.Fatalf("expected utf8-valid preview, got %q", got)
	}
	if got != "中文题...(truncated)" {
		t.Fatalf("unexpected truncated preview: %q", got)
	}
}

func TestAISolveServiceSolveFinalizesRunWithCanceledRequestContext(t *testing.T) {
	llmClient := &fakeLLMClient{
		err: errors.New("upstream eof"),
	}
	runRepo := &fakeAISolveRunRepository{rejectCanceledCtx: true}
	service := NewAISolveService(
		fakeProblemRepository{
			problem: &model.Problem{
				BaseModel:   model.BaseModel{ID: 1},
				Title:       "中文题目",
				Description: "给定一个数组",
				InputSpec:   "输入 n",
				OutputSpec:  "输出答案",
				Samples:     "[]",
			},
		},
		runRepo,
		llmClient,
		&fakeJudgeSubmitter{},
		"default-model",
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	output, err := service.Solve(ctx, AISolveInput{ProblemID: 1})
	if !errors.Is(err, ErrAISolveLLMFailed) {
		t.Fatalf("expected err %v, got %v", ErrAISolveLLMFailed, err)
	}
	if output == nil || output.AISolveRunID == 0 {
		t.Fatalf("expected run id in output, got %+v", output)
	}
	if len(runRepo.updated) != 1 {
		t.Fatalf("expected terminal update to succeed with background context, got %+v", runRepo.updated)
	}
	if runRepo.updated[0].Status != model.AISolveRunStatusFailed {
		t.Fatalf("expected failed terminal status, got %+v", runRepo.updated[0])
	}
}
