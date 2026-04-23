package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"ai-for-oj/internal/agent"
	"ai-for-oj/internal/model"
	"ai-for-oj/internal/prompt"
	"ai-for-oj/internal/repository"
)

type fakeExperimentRepository struct {
	experiment *model.Experiment
	list       []model.Experiment
	runTrace   *model.ExperimentRun
	runs       []*model.ExperimentRun
	updates    []*model.Experiment
	nextID     uint
	getByID    *model.Experiment
	err        error
}

func (r *fakeExperimentRepository) Create(_ context.Context, experiment *model.Experiment) error {
	if r.err != nil {
		return r.err
	}
	if r.nextID == 0 {
		r.nextID = 1
	}
	experiment.ID = r.nextID
	copied := *experiment
	r.experiment = &copied
	return nil
}

func (r *fakeExperimentRepository) Update(_ context.Context, experiment *model.Experiment) error {
	if r.err != nil {
		return r.err
	}
	copied := *experiment
	r.experiment = &copied
	r.updates = append(r.updates, &copied)
	if r.getByID == nil || r.getByID.ID == experiment.ID {
		r.getByID = &copied
		r.getByID.Runs = make([]model.ExperimentRun, 0, len(r.runs))
		for _, run := range r.runs {
			r.getByID.Runs = append(r.getByID.Runs, *run)
		}
	}
	return nil
}

func (r *fakeExperimentRepository) CreateRun(_ context.Context, run *model.ExperimentRun) error {
	if r.err != nil {
		return r.err
	}
	run.ID = uint(len(r.runs) + 1)
	run.CreatedAt = time.Now().UTC()
	copied := *run
	r.runs = append(r.runs, &copied)
	return nil
}

func (r *fakeExperimentRepository) List(_ context.Context, query repository.ExperimentListQuery) ([]model.Experiment, int64, error) {
	if r.err != nil {
		return nil, 0, r.err
	}
	return r.list, int64(len(r.list)), nil
}

func (r *fakeExperimentRepository) GetByIDWithRuns(_ context.Context, experimentID uint) (*model.Experiment, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.getByID == nil || r.getByID.ID != experimentID {
		return nil, repository.ErrExperimentNotFound
	}
	return r.getByID, nil
}

func (r *fakeExperimentRepository) GetRunTrace(_ context.Context, runID uint) (*model.ExperimentRun, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.runTrace == nil || r.runTrace.ID != runID {
		return nil, repository.ErrExperimentRunNotFound
	}
	return r.runTrace, nil
}

type fakeBatchAISolver struct {
	outputs map[uint]*AISolveOutput
	errors  map[uint]error
	inputs  []AISolveInput
}

func (s *fakeBatchAISolver) Solve(_ context.Context, input AISolveInput) (*AISolveOutput, error) {
	s.inputs = append(s.inputs, input)
	return s.outputs[input.ProblemID], s.errors[input.ProblemID]
}

func TestExperimentServiceRun(t *testing.T) {
	repo := &fakeExperimentRepository{}
	aiSolver := &fakeBatchAISolver{
		outputs: map[uint]*AISolveOutput{
			1: {
				AISolveRunID:   11,
				ProblemID:      1,
				SubmissionID:   101,
				Verdict:        "AC",
				TokenInput:     100,
				TokenOutput:    40,
				LLMLatencyMS:   200,
				TotalLatencyMS: 350,
			},
			2: {
				AISolveRunID:   12,
				ProblemID:      2,
				SubmissionID:   102,
				Verdict:        "WA",
				ErrorMessage:   "wrong answer",
				TokenInput:     80,
				TokenOutput:    20,
				LLMLatencyMS:   150,
				TotalLatencyMS: 260,
			},
		},
		errors: map[uint]error{},
	}
	service := NewExperimentService(repo, aiSolver, "mock-cpp17")

	output, err := service.Run(context.Background(), RunExperimentInput{
		Name:       "batch-1",
		ProblemIDs: []uint{1, 2},
		Model:      "mock-cpp17",
		PromptName: prompt.StrictCPP17SolvePromptName,
		AgentName:  agent.AnalyzeThenCodegenAgentName,
	})
	if err != nil {
		t.Fatalf("run experiment returned error: %v", err)
	}

	if output.TotalCount != 2 || output.SuccessCount != 2 || output.ACCount != 1 || output.FailedCount != 0 {
		t.Fatalf("unexpected summary: %+v", output)
	}
	if output.VerdictDistribution.ACCount != 1 || output.VerdictDistribution.WACount != 1 {
		t.Fatalf("unexpected verdict distribution: %+v", output.VerdictDistribution)
	}
	if output.CostSummary.RunCount != 2 ||
		output.CostSummary.TotalTokenInput != 180 ||
		output.CostSummary.TotalTokenOutput != 60 ||
		output.CostSummary.TotalTokens != 240 ||
		output.CostSummary.TotalLLMLatencyMS != 350 ||
		output.CostSummary.TotalLatencyMS != 610 {
		t.Fatalf("unexpected cost summary totals: %+v", output.CostSummary)
	}
	if output.CostSummary.AverageTokenInput != 90 ||
		output.CostSummary.AverageTokenOutput != 30 ||
		output.CostSummary.AverageTotalTokens != 120 ||
		output.CostSummary.AverageLLMLatencyMS != 175 ||
		output.CostSummary.AverageTotalLatencyMS != 305 {
		t.Fatalf("unexpected cost summary averages: %+v", output.CostSummary)
	}

	if len(output.Runs) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(output.Runs))
	}
	if output.PromptName != prompt.StrictCPP17SolvePromptName {
		t.Fatalf("expected experiment prompt name in output, got %q", output.PromptName)
	}
	if output.AgentName != agent.AnalyzeThenCodegenAgentName {
		t.Fatalf("expected experiment agent name in output, got %q", output.AgentName)
	}
	if len(aiSolver.inputs) != 2 || aiSolver.inputs[0].Model != "mock-cpp17" || aiSolver.inputs[1].Model != "mock-cpp17" {
		t.Fatalf("expected experiment model to be passed to every solve, got %+v", aiSolver.inputs)
	}
	if aiSolver.inputs[0].PromptName != prompt.StrictCPP17SolvePromptName || aiSolver.inputs[1].PromptName != prompt.StrictCPP17SolvePromptName {
		t.Fatalf("expected experiment prompt name to be passed to every solve, got %+v", aiSolver.inputs)
	}
	if aiSolver.inputs[0].AgentName != agent.AnalyzeThenCodegenAgentName || aiSolver.inputs[1].AgentName != agent.AnalyzeThenCodegenAgentName {
		t.Fatalf("expected experiment agent name to be passed to every solve, got %+v", aiSolver.inputs)
	}

	if output.Runs[0].AttemptNo != 1 || output.Runs[1].AttemptNo != 2 {
		t.Fatalf("expected sequential attempt numbers, got %+v", output.Runs)
	}
	if len(repo.updates) < 3 {
		t.Fatalf("expected per-run progress updates plus final update, got %d", len(repo.updates))
	}
	if repo.updates[0].Status != ExperimentStatusRunning || repo.updates[0].SuccessCount != 1 || repo.updates[0].ACCount != 1 {
		t.Fatalf("expected first progress update after first run, got %+v", repo.updates[0])
	}
	if repo.updates[1].Status != ExperimentStatusRunning || repo.updates[1].SuccessCount != 2 || repo.updates[1].ACCount != 1 {
		t.Fatalf("expected second progress update after second run, got %+v", repo.updates[1])
	}
	if repo.updates[len(repo.updates)-1].Status != ExperimentStatusCompleted {
		t.Fatalf("expected final update to complete experiment, got %+v", repo.updates[len(repo.updates)-1])
	}
}

func TestExperimentServiceRunPassesToolingConfig(t *testing.T) {
	repo := &fakeExperimentRepository{}
	aiSolver := &fakeBatchAISolver{
		outputs: map[uint]*AISolveOutput{
			1: {
				AISolveRunID:  11,
				ProblemID:     1,
				SubmissionID:  101,
				Verdict:       "AC",
				ToolingConfig: `{"enabled":["sample_judge"],"max_calls":1,"per_tool_max_calls":{}}`,
			},
		},
		errors: map[uint]error{},
	}
	service := NewExperimentService(repo, aiSolver, "mock-cpp17")

	output, err := service.Run(context.Background(), RunExperimentInput{
		Name:          "tooling-exp",
		ProblemIDs:    []uint{1},
		Model:         "mock-cpp17",
		ToolingConfig: `{"enabled":["sample_judge"],"max_calls":1}`,
	})
	if err != nil {
		t.Fatalf("run experiment returned error: %v", err)
	}

	if len(aiSolver.inputs) != 1 || aiSolver.inputs[0].ToolingConfig != `{"enabled":["sample_judge"],"max_calls":1,"per_tool_max_calls":{}}` {
		t.Fatalf("expected tooling config to be passed to solve, got %+v", aiSolver.inputs)
	}
	if output.ToolingConfig != `{"enabled":["sample_judge"],"max_calls":1,"per_tool_max_calls":{}}` {
		t.Fatalf("expected canonical tooling config in output, got %s", output.ToolingConfig)
	}
	if repo.experiment == nil || repo.experiment.ToolingConfig != output.ToolingConfig {
		t.Fatalf("expected tooling config persisted, got %+v", repo.experiment)
	}
}

func TestExperimentServiceRunContinuesAfterFailure(t *testing.T) {
	repo := &fakeExperimentRepository{}
	aiSolver := &fakeBatchAISolver{
		outputs: map[uint]*AISolveOutput{
			1: {AISolveRunID: 21, ProblemID: 1, SubmissionID: 201, Verdict: "AC", TokenInput: 50, TokenOutput: 10, LLMLatencyMS: 90, TotalLatencyMS: 140},
			2: {AISolveRunID: 22, ProblemID: 2},
			3: {AISolveRunID: 23, ProblemID: 3, SubmissionID: 203, Verdict: "UNJUDGEABLE", TokenInput: 70, TokenOutput: 30, LLMLatencyMS: 110, TotalLatencyMS: 170},
		},
		errors: map[uint]error{
			2: errors.New("llm solve failed: upstream timeout"),
		},
	}
	service := NewExperimentService(repo, aiSolver, "mock-cpp17")

	output, err := service.Run(context.Background(), RunExperimentInput{
		Name:       "batch-2",
		ProblemIDs: []uint{1, 2, 3},
		Model:      "mock-cpp17",
	})
	if err != nil {
		t.Fatalf("run experiment returned error: %v", err)
	}

	if len(aiSolver.inputs) != 3 {
		t.Fatalf("expected all problems to be processed, got %d", len(aiSolver.inputs))
	}

	if output.TotalCount != 3 || output.SuccessCount != 2 || output.FailedCount != 1 || output.ACCount != 1 {
		t.Fatalf("unexpected summary after partial failure: %+v", output)
	}
	if output.VerdictDistribution.ACCount != 1 || output.VerdictDistribution.UnjudgeableCount != 1 || output.VerdictDistribution.UnknownCount != 1 {
		t.Fatalf("unexpected verdict distribution after partial failure: %+v", output.VerdictDistribution)
	}
	if output.CostSummary.RunCount != 3 ||
		output.CostSummary.TotalTokenInput != 120 ||
		output.CostSummary.TotalTokenOutput != 40 ||
		output.CostSummary.TotalTokens != 160 ||
		output.CostSummary.TotalLLMLatencyMS != 200 ||
		output.CostSummary.TotalLatencyMS != 310 {
		t.Fatalf("unexpected cost summary after partial failure: %+v", output.CostSummary)
	}
	if output.CostSummary.AverageTokenInput != 40 ||
		output.CostSummary.AverageTokenOutput != float64(40)/3 ||
		output.CostSummary.AverageTotalTokens != float64(160)/3 ||
		output.CostSummary.AverageLLMLatencyMS != float64(200)/3 ||
		output.CostSummary.AverageTotalLatencyMS != float64(310)/3 {
		t.Fatalf("unexpected cost summary averages after partial failure: %+v", output.CostSummary)
	}

	if output.Runs[1].Status != ExperimentRunStatusFailed || output.Runs[1].ErrorMessage == "" {
		t.Fatalf("expected failed run to be recorded, got %+v", output.Runs[1])
	}
	if aiSolver.inputs[0].Model != "mock-cpp17" || aiSolver.inputs[1].Model != "mock-cpp17" || aiSolver.inputs[2].Model != "mock-cpp17" {
		t.Fatalf("expected resolved experiment model to be preserved across all solves, got %+v", aiSolver.inputs)
	}
}

func TestExperimentServiceRunFallsBackToDefaultModelForEverySolve(t *testing.T) {
	repo := &fakeExperimentRepository{}
	aiSolver := &fakeBatchAISolver{
		outputs: map[uint]*AISolveOutput{
			1: {AISolveRunID: 31, ProblemID: 1, SubmissionID: 301, Verdict: "AC"},
		},
		errors: map[uint]error{},
	}
	service := NewExperimentService(repo, aiSolver, "default-model")

	output, err := service.Run(context.Background(), RunExperimentInput{
		Name:       "batch-default",
		ProblemIDs: []uint{1},
	})
	if err != nil {
		t.Fatalf("run experiment returned error: %v", err)
	}
	if output.Model != "default-model" {
		t.Fatalf("expected experiment output model to use default, got %q", output.Model)
	}
	if output.PromptName != prompt.DefaultSolvePromptName {
		t.Fatalf("expected experiment output prompt name to use default, got %q", output.PromptName)
	}
	if output.AgentName != agent.DirectCodegenAgentName {
		t.Fatalf("expected experiment output agent name to use default, got %q", output.AgentName)
	}
	if len(aiSolver.inputs) != 1 || aiSolver.inputs[0].Model != "default-model" {
		t.Fatalf("expected default model to be passed to solve, got %+v", aiSolver.inputs)
	}
	if len(aiSolver.inputs) != 1 || aiSolver.inputs[0].PromptName != prompt.DefaultSolvePromptName {
		t.Fatalf("expected default prompt to be passed to solve, got %+v", aiSolver.inputs)
	}
	if len(aiSolver.inputs) != 1 || aiSolver.inputs[0].AgentName != agent.DirectCodegenAgentName {
		t.Fatalf("expected default agent to be passed to solve, got %+v", aiSolver.inputs)
	}
}

func TestExperimentServiceGetReturnsZeroCostSummaryWithoutValidAISolveRuns(t *testing.T) {
	repo := &fakeExperimentRepository{
		getByID: &model.Experiment{
			BaseModel:    model.BaseModel{ID: 9, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
			Name:         "batch-3",
			ModelName:    "mock-cpp17",
			Status:       ExperimentStatusCompleted,
			TotalCount:   2,
			SuccessCount: 1,
			FailedCount:  1,
			Runs: []model.ExperimentRun{
				{CreatedModel: model.CreatedModel{ID: 1, CreatedAt: time.Now().UTC()}, ProblemID: 1, AttemptNo: 1, Status: ExperimentRunStatusSuccess, FinalVerdict: "AC"},
				{CreatedModel: model.CreatedModel{ID: 2, CreatedAt: time.Now().UTC()}, ProblemID: 2, AttemptNo: 2, Status: ExperimentRunStatusFailed, FinalVerdict: "", AISolveRunID: uintPtr(99)},
			},
		},
	}
	service := NewExperimentService(repo, &fakeBatchAISolver{}, "mock-cpp17")

	output, err := service.Get(context.Background(), 9)
	if err != nil {
		t.Fatalf("get experiment returned error: %v", err)
	}

	if output.CostSummary.RunCount != 0 ||
		output.CostSummary.TotalTokenInput != 0 ||
		output.CostSummary.TotalTokenOutput != 0 ||
		output.CostSummary.TotalTokens != 0 ||
		output.CostSummary.TotalLLMLatencyMS != 0 ||
		output.CostSummary.TotalLatencyMS != 0 ||
		output.CostSummary.AverageTokenInput != 0 ||
		output.CostSummary.AverageTokenOutput != 0 ||
		output.CostSummary.AverageTotalTokens != 0 ||
		output.CostSummary.AverageLLMLatencyMS != 0 ||
		output.CostSummary.AverageTotalLatencyMS != 0 {
		t.Fatalf("expected zero-value cost summary, got %+v", output.CostSummary)
	}
}

func TestExperimentServiceGetExposesAISolveRunSummaryFields(t *testing.T) {
	repo := &fakeExperimentRepository{
		getByID: &model.Experiment{
			BaseModel:    model.BaseModel{ID: 10, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
			Name:         "batch-4",
			ModelName:    "mock-cpp17",
			Status:       ExperimentStatusCompleted,
			TotalCount:   1,
			SuccessCount: 1,
			Runs: []model.ExperimentRun{
				{
					CreatedModel: model.CreatedModel{ID: 3, CreatedAt: time.Now().UTC()},
					ProblemID:    3,
					AttemptNo:    1,
					Status:       ExperimentRunStatusSuccess,
					AISolveRunID: uintPtr(77),
					AISolveRun: &model.AISolveRun{
						BaseModel:    model.BaseModel{ID: 77},
						AttemptCount: 4,
						FailureType:  "wrong_answer",
						StrategyPath: "initial,repair,final",
					},
				},
			},
		},
	}
	service := NewExperimentService(repo, &fakeBatchAISolver{}, "mock-cpp17")

	output, err := service.Get(context.Background(), 10)
	if err != nil {
		t.Fatalf("get experiment returned error: %v", err)
	}

	if len(output.Runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(output.Runs))
	}
	if output.Runs[0].AttemptCount != 4 || output.Runs[0].FailureType != "wrong_answer" || output.Runs[0].StrategyPath != "initial,repair,final" {
		t.Fatalf("expected AISolveRun summary fields to be exposed, got %+v", output.Runs[0])
	}
}
