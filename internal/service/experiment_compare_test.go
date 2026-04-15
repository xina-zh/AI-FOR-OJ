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

type fakeExperimentCompareRepository struct {
	created *model.ExperimentCompare
	updated *model.ExperimentCompare
	getByID *model.ExperimentCompare
	nextID  uint
	err     error
}

func (r *fakeExperimentCompareRepository) Create(_ context.Context, compare *model.ExperimentCompare) error {
	if r.err != nil {
		return r.err
	}
	if r.nextID == 0 {
		r.nextID = 1
	}
	compare.ID = r.nextID
	compare.CreatedAt = time.Now().UTC()
	compare.UpdatedAt = compare.CreatedAt
	copied := *compare
	r.created = &copied
	return nil
}

func (r *fakeExperimentCompareRepository) Update(_ context.Context, compare *model.ExperimentCompare) error {
	if r.err != nil {
		return r.err
	}
	compare.UpdatedAt = time.Now().UTC()
	copied := *compare
	r.updated = &copied
	r.getByID = &copied
	return nil
}

func (r *fakeExperimentCompareRepository) GetByID(_ context.Context, compareID uint) (*model.ExperimentCompare, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.getByID == nil || r.getByID.ID != compareID {
		return nil, repository.ErrExperimentCompareNotFound
	}
	return r.getByID, nil
}

type fakeExperimentRunner struct {
	runOutputs []*ExperimentOutput
	runErrors  []error
	getMap     map[uint]*ExperimentOutput
	runInputs  []RunExperimentInput
}

func (r *fakeExperimentRunner) Run(_ context.Context, input RunExperimentInput) (*ExperimentOutput, error) {
	r.runInputs = append(r.runInputs, input)
	index := len(r.runInputs) - 1
	if index < len(r.runErrors) && r.runErrors[index] != nil {
		return nil, r.runErrors[index]
	}
	return r.runOutputs[index], nil
}

func (r *fakeExperimentRunner) Get(_ context.Context, experimentID uint) (*ExperimentOutput, error) {
	output, ok := r.getMap[experimentID]
	if !ok {
		return nil, repository.ErrExperimentNotFound
	}
	return output, nil
}

func TestExperimentCompareServiceCompare(t *testing.T) {
	repo := &fakeExperimentCompareRepository{}
	runner := &fakeExperimentRunner{
		runOutputs: []*ExperimentOutput{
			{
				ID:                  10,
				Name:                "baseline",
				Model:               "mock-a",
				PromptName:          prompt.DefaultSolvePromptName,
				AgentName:           agent.DirectCodegenAgentName,
				TotalCount:          3,
				SuccessCount:        3,
				ACCount:             1,
				FailedCount:         0,
				VerdictDistribution: VerdictDistribution{ACCount: 1, WACount: 1, CECount: 1},
				CostSummary: ExperimentCostSummary{
					TotalTokenInput:       300,
					TotalTokenOutput:      120,
					TotalTokens:           420,
					AverageTokenInput:     100,
					AverageTokenOutput:    40,
					AverageTotalTokens:    140,
					TotalLLMLatencyMS:     450,
					TotalLatencyMS:        780,
					AverageLLMLatencyMS:   150,
					AverageTotalLatencyMS: 260,
					RunCount:              3,
				},
				Runs: []ExperimentRunOutput{
					{ProblemID: 1, Verdict: "WA", Status: ExperimentRunStatusSuccess, SubmissionID: uintPtr(101)},
					{ProblemID: 2, Verdict: "AC", Status: ExperimentRunStatusSuccess, SubmissionID: uintPtr(102)},
					{ProblemID: 3, Verdict: "CE", Status: ExperimentRunStatusSuccess, SubmissionID: uintPtr(103)},
				},
			},
			{
				ID:                  20,
				Name:                "candidate",
				Model:               "mock-b",
				PromptName:          prompt.StrictCPP17SolvePromptName,
				AgentName:           agent.AnalyzeThenCodegenAgentName,
				TotalCount:          3,
				SuccessCount:        3,
				ACCount:             2,
				FailedCount:         0,
				VerdictDistribution: VerdictDistribution{ACCount: 2, RECount: 1},
				CostSummary: ExperimentCostSummary{
					TotalTokenInput:       360,
					TotalTokenOutput:      150,
					TotalTokens:           510,
					AverageTokenInput:     120,
					AverageTokenOutput:    50,
					AverageTotalTokens:    170,
					TotalLLMLatencyMS:     510,
					TotalLatencyMS:        870,
					AverageLLMLatencyMS:   170,
					AverageTotalLatencyMS: 290,
					RunCount:              3,
				},
				Runs: []ExperimentRunOutput{
					{ProblemID: 1, Verdict: "AC", Status: ExperimentRunStatusSuccess, SubmissionID: uintPtr(201)},
					{ProblemID: 2, Verdict: "AC", Status: ExperimentRunStatusSuccess, SubmissionID: uintPtr(202)},
					{ProblemID: 3, Verdict: "RE", Status: ExperimentRunStatusSuccess, SubmissionID: uintPtr(203)},
				},
			},
		},
		getMap: map[uint]*ExperimentOutput{
			10: {
				ID:                  10,
				Name:                "baseline",
				Model:               "mock-a",
				PromptName:          prompt.DefaultSolvePromptName,
				AgentName:           agent.DirectCodegenAgentName,
				TotalCount:          3,
				SuccessCount:        3,
				ACCount:             1,
				FailedCount:         0,
				VerdictDistribution: VerdictDistribution{ACCount: 1, WACount: 1, CECount: 1},
				CostSummary: ExperimentCostSummary{
					TotalTokenInput:       300,
					TotalTokenOutput:      120,
					TotalTokens:           420,
					AverageTokenInput:     100,
					AverageTokenOutput:    40,
					AverageTotalTokens:    140,
					TotalLLMLatencyMS:     450,
					TotalLatencyMS:        780,
					AverageLLMLatencyMS:   150,
					AverageTotalLatencyMS: 260,
					RunCount:              3,
				},
				Runs: []ExperimentRunOutput{
					{ProblemID: 1, Verdict: "WA", Status: ExperimentRunStatusSuccess, SubmissionID: uintPtr(101)},
					{ProblemID: 2, Verdict: "AC", Status: ExperimentRunStatusSuccess, SubmissionID: uintPtr(102)},
					{ProblemID: 3, Verdict: "CE", Status: ExperimentRunStatusSuccess, SubmissionID: uintPtr(103)},
				},
			},
			20: {
				ID:                  20,
				Name:                "candidate",
				Model:               "mock-b",
				PromptName:          prompt.StrictCPP17SolvePromptName,
				AgentName:           agent.AnalyzeThenCodegenAgentName,
				TotalCount:          3,
				SuccessCount:        3,
				ACCount:             2,
				FailedCount:         0,
				VerdictDistribution: VerdictDistribution{ACCount: 2, RECount: 1},
				CostSummary: ExperimentCostSummary{
					TotalTokenInput:       360,
					TotalTokenOutput:      150,
					TotalTokens:           510,
					AverageTokenInput:     120,
					AverageTokenOutput:    50,
					AverageTotalTokens:    170,
					TotalLLMLatencyMS:     510,
					TotalLatencyMS:        870,
					AverageLLMLatencyMS:   170,
					AverageTotalLatencyMS: 290,
					RunCount:              3,
				},
				Runs: []ExperimentRunOutput{
					{ProblemID: 1, Verdict: "AC", Status: ExperimentRunStatusSuccess, SubmissionID: uintPtr(201)},
					{ProblemID: 2, Verdict: "AC", Status: ExperimentRunStatusSuccess, SubmissionID: uintPtr(202)},
					{ProblemID: 3, Verdict: "RE", Status: ExperimentRunStatusSuccess, SubmissionID: uintPtr(203)},
				},
			},
		},
	}
	service := NewExperimentCompareService(repo, runner, "mock-default")

	output, err := service.Compare(context.Background(), CompareExperimentInput{
		Name:                "compare-1",
		ProblemIDs:          []uint{1, 2, 3},
		BaselineModel:       "mock-a",
		CandidateModel:      "mock-b",
		BaselinePromptName:  prompt.DefaultSolvePromptName,
		CandidatePromptName: prompt.StrictCPP17SolvePromptName,
		BaselineAgentName:   agent.DirectCodegenAgentName,
		CandidateAgentName:  agent.AnalyzeThenCodegenAgentName,
	})
	if err != nil {
		t.Fatalf("compare returned error: %v", err)
	}

	if len(runner.runInputs) != 2 {
		t.Fatalf("expected baseline and candidate to run, got %d calls", len(runner.runInputs))
	}
	if runner.runInputs[0].Model != "mock-a" || runner.runInputs[1].Model != "mock-b" {
		t.Fatalf("expected baseline/candidate models to stay independent, got %+v", runner.runInputs)
	}
	if runner.runInputs[0].PromptName != prompt.DefaultSolvePromptName || runner.runInputs[1].PromptName != prompt.StrictCPP17SolvePromptName {
		t.Fatalf("expected baseline/candidate prompt names to stay independent, got %+v", runner.runInputs)
	}
	if runner.runInputs[0].AgentName != agent.DirectCodegenAgentName || runner.runInputs[1].AgentName != agent.AnalyzeThenCodegenAgentName {
		t.Fatalf("expected baseline/candidate agent names to stay independent, got %+v", runner.runInputs)
	}

	if output.CompareDimension != ExperimentCompareDimensionModel {
		t.Fatalf("expected compare dimension model, got %s", output.CompareDimension)
	}
	if output.BaselinePromptName != prompt.DefaultSolvePromptName || output.CandidatePromptName != prompt.StrictCPP17SolvePromptName {
		t.Fatalf("unexpected top-level prompt names: %+v", output)
	}
	if output.BaselineAgentName != agent.DirectCodegenAgentName || output.CandidateAgentName != agent.AnalyzeThenCodegenAgentName {
		t.Fatalf("unexpected top-level agent names: %+v", output)
	}

	if output.DeltaACCount != 1 || output.DeltaFailedCount != 0 {
		t.Fatalf("unexpected delta summary: %+v", output)
	}
	if output.BaselineDistribution.WACount != 1 || output.CandidateDistribution.ACCount != 2 || output.DeltaDistribution.ACCount != 1 || output.DeltaDistribution.WACount != -1 {
		t.Fatalf("unexpected verdict distribution delta: %+v", output)
	}
	if output.CostComparison.BaselineTotalTokens != 420 ||
		output.CostComparison.CandidateTotalTokens != 510 ||
		output.CostComparison.DeltaTotalTokens != 90 ||
		output.CostComparison.DeltaAverageTotalTokens != 30 ||
		output.CostComparison.DeltaTotalLatencyMS != 90 ||
		output.CostComparison.DeltaAverageTotalLatencyMS != 30 {
		t.Fatalf("unexpected cost comparison summary: %+v", output.CostComparison)
	}
	if !output.ComparisonSummary.CandidateBetterAC ||
		output.ComparisonSummary.CandidateWorseAC ||
		output.ComparisonSummary.CandidateSameAC ||
		!output.ComparisonSummary.CandidateMoreExpensive ||
		output.ComparisonSummary.CandidateCheaper ||
		output.ComparisonSummary.CandidateSameCost ||
		!output.ComparisonSummary.CandidateSlower ||
		output.ComparisonSummary.CandidateFaster ||
		output.ComparisonSummary.CandidateSameLatency ||
		output.ComparisonSummary.TradeoffType != "improved_with_higher_cost" {
		t.Fatalf("unexpected comparison summary: %+v", output.ComparisonSummary)
	}
	if output.ImprovedCount != 1 || output.RegressedCount != 0 || output.ChangedNonACCount != 1 {
		t.Fatalf("unexpected change summary: %+v", output)
	}
	if len(output.ProblemSummaries) != 3 || output.ProblemSummaries[0].ChangeType != "improved" || !output.ProblemSummaries[0].Changed {
		t.Fatalf("unexpected problem summaries: %+v", output.ProblemSummaries)
	}
	if output.ProblemSummaries[1].ChangeType != "same" || output.ProblemSummaries[1].Changed {
		t.Fatalf("unexpected second problem summary: %+v", output.ProblemSummaries[1])
	}
	if output.ProblemSummaries[2].ChangeType != "changed_non_ac" || !output.ProblemSummaries[2].Changed {
		t.Fatalf("unexpected third problem summary: %+v", output.ProblemSummaries[2])
	}
	if len(output.HighlightedProblems) != 3 {
		t.Fatalf("unexpected highlighted problems length: %+v", output.HighlightedProblems)
	}
	if output.HighlightedProblems[0].ProblemID != 1 || output.HighlightedProblems[0].ChangeType != "improved" {
		t.Fatalf("unexpected first highlighted problem: %+v", output.HighlightedProblems[0])
	}
	if output.HighlightedProblems[1].ProblemID != 3 || output.HighlightedProblems[1].ChangeType != "changed_non_ac" {
		t.Fatalf("unexpected second highlighted problem: %+v", output.HighlightedProblems[1])
	}
	if output.HighlightedProblems[2].ProblemID != 2 || output.HighlightedProblems[2].ChangeType != "same" {
		t.Fatalf("unexpected third highlighted problem: %+v", output.HighlightedProblems[2])
	}
}

func TestExperimentCompareServiceCompareCandidateFailure(t *testing.T) {
	repo := &fakeExperimentCompareRepository{}
	runner := &fakeExperimentRunner{
		runOutputs: []*ExperimentOutput{
			{ID: 10, Name: "baseline", Model: "mock-a", TotalCount: 2, SuccessCount: 2, ACCount: 1, FailedCount: 0, VerdictDistribution: VerdictDistribution{ACCount: 1, WACount: 1}},
			nil,
		},
		runErrors: []error{
			nil,
			errors.New("candidate run failed"),
		},
		getMap: map[uint]*ExperimentOutput{
			10: {ID: 10, Name: "baseline", Model: "mock-a", TotalCount: 2, SuccessCount: 2, ACCount: 1, FailedCount: 0, VerdictDistribution: VerdictDistribution{ACCount: 1, WACount: 1}},
		},
	}
	service := NewExperimentCompareService(repo, runner, "mock-default")

	_, err := service.Compare(context.Background(), CompareExperimentInput{
		Name:                "compare-2",
		ProblemIDs:          []uint{1, 2},
		BaselineModel:       "mock-a",
		CandidateModel:      "mock-b",
		BaselinePromptName:  prompt.DefaultSolvePromptName,
		CandidatePromptName: prompt.StrictCPP17SolvePromptName,
		BaselineAgentName:   agent.DirectCodegenAgentName,
		CandidateAgentName:  agent.AnalyzeThenCodegenAgentName,
	})
	if err == nil {
		t.Fatal("expected compare to return error when candidate run fails")
	}

	if repo.updated == nil || repo.updated.Status != model.ExperimentCompareStatusFailed {
		t.Fatalf("expected failed compare to be persisted, got %+v", repo.updated)
	}
	if len(runner.runInputs) != 2 || runner.runInputs[0].Model != "mock-a" || runner.runInputs[1].Model != "mock-b" {
		t.Fatalf("expected compare to pass distinct models before candidate failure, got %+v", runner.runInputs)
	}
	if runner.runInputs[0].PromptName != prompt.DefaultSolvePromptName || runner.runInputs[1].PromptName != prompt.StrictCPP17SolvePromptName {
		t.Fatalf("expected compare to preserve distinct prompt names before candidate failure, got %+v", runner.runInputs)
	}
	if runner.runInputs[0].AgentName != agent.DirectCodegenAgentName || runner.runInputs[1].AgentName != agent.AnalyzeThenCodegenAgentName {
		t.Fatalf("expected compare to preserve distinct agent names before candidate failure, got %+v", runner.runInputs)
	}
}

func TestExperimentCompareServiceGet(t *testing.T) {
	repo := &fakeExperimentCompareRepository{
		getByID: &model.ExperimentCompare{
			BaseModel:             model.BaseModel{ID: 5, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
			Name:                  "compare-3",
			CompareDimension:      ExperimentCompareDimensionModel,
			BaselineValue:         "mock-a",
			CandidateValue:        "mock-b",
			BaselinePromptName:    prompt.DefaultSolvePromptName,
			CandidatePromptName:   prompt.StrictCPP17SolvePromptName,
			BaselineAgentName:     agent.DirectCodegenAgentName,
			CandidateAgentName:    agent.AnalyzeThenCodegenAgentName,
			ProblemIDs:            "[1,2]",
			BaselineExperimentID:  uintPtr(10),
			CandidateExperimentID: uintPtr(20),
			DeltaACCount:          1,
			DeltaFailedCount:      -1,
			Status:                model.ExperimentCompareStatusCompleted,
		},
	}
	runner := &fakeExperimentRunner{
		getMap: map[uint]*ExperimentOutput{
			10: {
				ID:                  10,
				Name:                "baseline",
				PromptName:          prompt.DefaultSolvePromptName,
				AgentName:           agent.DirectCodegenAgentName,
				VerdictDistribution: VerdictDistribution{ACCount: 1, WACount: 1},
				CostSummary: ExperimentCostSummary{
					TotalTokenInput:       210,
					TotalTokenOutput:      90,
					TotalTokens:           300,
					AverageTokenInput:     105,
					AverageTokenOutput:    45,
					AverageTotalTokens:    150,
					TotalLLMLatencyMS:     260,
					TotalLatencyMS:        480,
					AverageLLMLatencyMS:   130,
					AverageTotalLatencyMS: 240,
					RunCount:              2,
				},
				Runs: []ExperimentRunOutput{
					{ProblemID: 1, Verdict: "WA", Status: ExperimentRunStatusSuccess, SubmissionID: uintPtr(101)},
					{ProblemID: 2, Verdict: "AC", Status: ExperimentRunStatusSuccess, SubmissionID: uintPtr(102)},
				},
			},
			20: {
				ID:                  20,
				Name:                "candidate",
				PromptName:          prompt.StrictCPP17SolvePromptName,
				AgentName:           agent.AnalyzeThenCodegenAgentName,
				VerdictDistribution: VerdictDistribution{ACCount: 1, CECount: 1},
				CostSummary: ExperimentCostSummary{
					TotalTokenInput:       180,
					TotalTokenOutput:      60,
					TotalTokens:           240,
					AverageTokenInput:     90,
					AverageTokenOutput:    30,
					AverageTotalTokens:    120,
					TotalLLMLatencyMS:     220,
					TotalLatencyMS:        390,
					AverageLLMLatencyMS:   110,
					AverageTotalLatencyMS: 195,
					RunCount:              2,
				},
				Runs: []ExperimentRunOutput{
					{ProblemID: 1, Verdict: "RE", Status: ExperimentRunStatusSuccess, SubmissionID: uintPtr(201)},
					{ProblemID: 2, Verdict: "AC", Status: ExperimentRunStatusSuccess, SubmissionID: uintPtr(202)},
				},
			},
		},
	}
	service := NewExperimentCompareService(repo, runner, "mock-default")

	output, err := service.Get(context.Background(), 5)
	if err != nil {
		t.Fatalf("get compare returned error: %v", err)
	}

	if output.BaselineExperimentID != 10 || output.CandidateExperimentID != 20 {
		t.Fatalf("unexpected experiment ids: %+v", output)
	}
	if output.BaselinePromptName != prompt.DefaultSolvePromptName || output.CandidatePromptName != prompt.StrictCPP17SolvePromptName {
		t.Fatalf("unexpected prompt names on get: %+v", output)
	}
	if output.BaselineAgentName != agent.DirectCodegenAgentName || output.CandidateAgentName != agent.AnalyzeThenCodegenAgentName {
		t.Fatalf("unexpected agent names on get: %+v", output)
	}

	if len(output.ProblemIDs) != 2 || output.ProblemIDs[0] != 1 {
		t.Fatalf("unexpected problem ids: %+v", output.ProblemIDs)
	}
	if output.DeltaDistribution.CECount != 1 || output.DeltaDistribution.WACount != -1 {
		t.Fatalf("unexpected delta distribution on get: %+v", output.DeltaDistribution)
	}
	if output.CostComparison.BaselineRunCount != 2 ||
		output.CostComparison.CandidateRunCount != 2 ||
		output.CostComparison.DeltaTotalTokens != -60 ||
		output.CostComparison.DeltaAverageTotalTokens != -30 ||
		output.CostComparison.DeltaTotalLatencyMS != -90 ||
		output.CostComparison.DeltaAverageTotalLatencyMS != -45 {
		t.Fatalf("unexpected cost comparison on get: %+v", output.CostComparison)
	}
	if output.ComparisonSummary.CandidateBetterAC ||
		output.ComparisonSummary.CandidateWorseAC ||
		!output.ComparisonSummary.CandidateSameAC ||
		output.ComparisonSummary.CandidateMoreExpensive ||
		!output.ComparisonSummary.CandidateCheaper ||
		output.ComparisonSummary.CandidateSameCost ||
		output.ComparisonSummary.CandidateSlower ||
		!output.ComparisonSummary.CandidateFaster ||
		output.ComparisonSummary.CandidateSameLatency ||
		output.ComparisonSummary.TradeoffType != "same_outcome_lower_cost" {
		t.Fatalf("unexpected comparison summary on get: %+v", output.ComparisonSummary)
	}
	if output.ImprovedCount != 0 || output.RegressedCount != 0 || output.ChangedNonACCount != 1 {
		t.Fatalf("unexpected per-problem change counts: %+v", output)
	}
	if len(output.ProblemSummaries) != 2 || output.ProblemSummaries[0].ChangeType != "changed_non_ac" || output.ProblemSummaries[1].ChangeType != "same" {
		t.Fatalf("unexpected problem summaries on get: %+v", output.ProblemSummaries)
	}
	if len(output.HighlightedProblems) != 2 || output.HighlightedProblems[0].ProblemID != 1 || output.HighlightedProblems[0].ChangeType != "changed_non_ac" {
		t.Fatalf("unexpected highlighted problems on get: %+v", output.HighlightedProblems)
	}
}

func TestExperimentCompareServiceCompareUsesPromptDimensionForSameModelDifferentPrompt(t *testing.T) {
	repo := &fakeExperimentCompareRepository{}
	runner := &fakeExperimentRunner{
		runOutputs: []*ExperimentOutput{
			{ID: 10, Model: "mock-a", PromptName: prompt.DefaultSolvePromptName, AgentName: agent.DirectCodegenAgentName},
			{ID: 20, Model: "mock-a", PromptName: prompt.StrictCPP17SolvePromptName, AgentName: agent.AnalyzeThenCodegenAgentName},
		},
	}
	service := NewExperimentCompareService(repo, runner, "mock-default")

	output, err := service.Compare(context.Background(), CompareExperimentInput{
		Name:                "compare-prompt",
		ProblemIDs:          []uint{1},
		BaselineModel:       "mock-a",
		CandidateModel:      "mock-a",
		BaselinePromptName:  prompt.DefaultSolvePromptName,
		CandidatePromptName: prompt.StrictCPP17SolvePromptName,
		BaselineAgentName:   agent.DirectCodegenAgentName,
		CandidateAgentName:  agent.AnalyzeThenCodegenAgentName,
	})
	if err != nil {
		t.Fatalf("compare returned error: %v", err)
	}
	if output.CompareDimension != ExperimentCompareDimensionPrompt {
		t.Fatalf("expected compare dimension prompt, got %s", output.CompareDimension)
	}
	if output.BaselineValue != prompt.DefaultSolvePromptName || output.CandidateValue != prompt.StrictCPP17SolvePromptName {
		t.Fatalf("expected prompt names to be stored as compare values, got %+v", output)
	}
}

func TestExperimentCompareServiceCompareUsesAgentDimensionForSameModelAndPrompt(t *testing.T) {
	repo := &fakeExperimentCompareRepository{}
	runner := &fakeExperimentRunner{
		runOutputs: []*ExperimentOutput{
			{ID: 10, Model: "mock-a", PromptName: prompt.DefaultSolvePromptName, AgentName: agent.DirectCodegenAgentName},
			{ID: 20, Model: "mock-a", PromptName: prompt.DefaultSolvePromptName, AgentName: agent.AnalyzeThenCodegenAgentName},
		},
	}
	service := NewExperimentCompareService(repo, runner, "mock-default")

	output, err := service.Compare(context.Background(), CompareExperimentInput{
		Name:                "compare-agent",
		ProblemIDs:          []uint{1},
		BaselineModel:       "mock-a",
		CandidateModel:      "mock-a",
		BaselinePromptName:  prompt.DefaultSolvePromptName,
		CandidatePromptName: prompt.DefaultSolvePromptName,
		BaselineAgentName:   agent.DirectCodegenAgentName,
		CandidateAgentName:  agent.AnalyzeThenCodegenAgentName,
	})
	if err != nil {
		t.Fatalf("compare returned error: %v", err)
	}
	if output.CompareDimension != ExperimentCompareDimensionAgent {
		t.Fatalf("expected compare dimension agent, got %s", output.CompareDimension)
	}
	if output.BaselineValue != agent.DirectCodegenAgentName || output.CandidateValue != agent.AnalyzeThenCodegenAgentName {
		t.Fatalf("expected agent names to be stored as compare values, got %+v", output)
	}
}

func TestBuildExperimentCompareSummarySameOutcomeSameCost(t *testing.T) {
	summary := buildExperimentCompareSummary(
		&ExperimentOutput{ACCount: 2},
		&ExperimentOutput{ACCount: 2},
		ExperimentCompareCostComparison{
			BaselineTotalTokens:            100,
			CandidateTotalTokens:           100,
			BaselineAverageTotalLatencyMS:  50,
			CandidateAverageTotalLatencyMS: 50,
		},
	)

	if summary.CandidateBetterAC || summary.CandidateWorseAC || !summary.CandidateSameAC {
		t.Fatalf("unexpected ac flags: %+v", summary)
	}
	if summary.CandidateMoreExpensive || summary.CandidateCheaper || !summary.CandidateSameCost {
		t.Fatalf("unexpected cost flags: %+v", summary)
	}
	if summary.CandidateSlower || summary.CandidateFaster || !summary.CandidateSameLatency {
		t.Fatalf("unexpected latency flags: %+v", summary)
	}
	if summary.TradeoffType != "same_outcome_same_cost" {
		t.Fatalf("unexpected tradeoff type: %+v", summary)
	}
}

func uintPtr(value uint) *uint {
	return &value
}
