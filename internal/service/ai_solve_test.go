package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"ai-for-oj/internal/llm"
	"ai-for-oj/internal/model"
	"ai-for-oj/internal/repository"
)

type fakeAISolveRunRepository struct {
	created *model.AISolveRun
	updated []*model.AISolveRun
	getRun  *model.AISolveRun
	err     error
	nextID  uint
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
	r.created = &copied
	return nil
}

func (r *fakeAISolveRunRepository) Update(_ context.Context, run *model.AISolveRun) error {
	if r.err != nil {
		return r.err
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
	response llm.GenerateResponse
	err      error
	request  llm.GenerateRequest
	delay    time.Duration
}

func (c *fakeLLMClient) Generate(_ context.Context, req llm.GenerateRequest) (llm.GenerateResponse, error) {
	c.request = req
	if c.delay > 0 {
		time.Sleep(c.delay)
	}
	return c.response, c.err
}

type fakeJudgeSubmitter struct {
	input  JudgeSubmissionInput
	output *JudgeSubmissionOutput
	err    error
}

func (s *fakeJudgeSubmitter) Submit(_ context.Context, input JudgeSubmissionInput) (*JudgeSubmissionOutput, error) {
	s.input = input
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
	if llmClient.request.Model != "default-model" {
		t.Fatalf("expected default model to be passed to llm client, got %q", llmClient.request.Model)
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

	output, err := service.Solve(context.Background(), AISolveInput{ProblemID: 1, Model: "request-model"})
	if err != nil {
		t.Fatalf("solve returned error: %v", err)
	}
	if llmClient.request.Model != "request-model" {
		t.Fatalf("expected request model to be passed to llm client, got %q", llmClient.request.Model)
	}
	if output.Model != "request-model" {
		t.Fatalf("expected output model to reflect request model, got %q", output.Model)
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

func TestAISolveServiceGetRun(t *testing.T) {
	runRepo := &fakeAISolveRunRepository{
		getRun: &model.AISolveRun{
			BaseModel:      model.BaseModel{ID: 5},
			ProblemID:      1,
			Model:          "mock-cpp17",
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
