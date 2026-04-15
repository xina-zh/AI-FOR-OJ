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

type fakeExperimentRepeatRepository struct {
	created *model.ExperimentRepeat
	updated *model.ExperimentRepeat
	getByID *model.ExperimentRepeat
	nextID  uint
	err     error
}

func (r *fakeExperimentRepeatRepository) Create(_ context.Context, repeat *model.ExperimentRepeat) error {
	if r.err != nil {
		return r.err
	}
	if r.nextID == 0 {
		r.nextID = 1
	}
	repeat.ID = r.nextID
	repeat.CreatedAt = time.Now().UTC()
	repeat.UpdatedAt = repeat.CreatedAt
	copied := *repeat
	r.created = &copied
	return nil
}

func (r *fakeExperimentRepeatRepository) Update(_ context.Context, repeat *model.ExperimentRepeat) error {
	if r.err != nil {
		return r.err
	}
	repeat.UpdatedAt = time.Now().UTC()
	copied := *repeat
	r.updated = &copied
	r.getByID = &copied
	return nil
}

func (r *fakeExperimentRepeatRepository) GetByID(_ context.Context, repeatID uint) (*model.ExperimentRepeat, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.getByID == nil || r.getByID.ID != repeatID {
		return nil, repository.ErrExperimentRepeatNotFound
	}
	return r.getByID, nil
}

func TestExperimentRepeatServiceRepeat(t *testing.T) {
	repo := &fakeExperimentRepeatRepository{}
	runner := &fakeExperimentRunner{
		runOutputs: []*ExperimentOutput{
			{
				ID:                  10,
				Name:                "round-1",
				Model:               "mock-cpp17",
				PromptName:          prompt.CPP17MinimalSolvePromptName,
				AgentName:           agent.AnalyzeThenCodegenAgentName,
				TotalCount:          2,
				SuccessCount:        2,
				ACCount:             1,
				FailedCount:         1,
				VerdictDistribution: VerdictDistribution{ACCount: 1, WACount: 1},
				CostSummary: ExperimentCostSummary{
					TotalTokenInput:   100,
					TotalTokenOutput:  40,
					TotalTokens:       140,
					TotalLLMLatencyMS: 200,
					TotalLatencyMS:    320,
					RunCount:          2,
				},
				Status: ExperimentStatusCompleted,
				Runs: []ExperimentRunOutput{
					{ProblemID: 1, Verdict: "AC", Status: ExperimentRunStatusSuccess},
					{ProblemID: 2, Verdict: "WA", Status: ExperimentRunStatusSuccess},
				},
			},
			{
				ID:                  11,
				Name:                "round-2",
				Model:               "mock-cpp17",
				PromptName:          prompt.CPP17MinimalSolvePromptName,
				AgentName:           agent.AnalyzeThenCodegenAgentName,
				TotalCount:          2,
				SuccessCount:        2,
				ACCount:             2,
				FailedCount:         0,
				VerdictDistribution: VerdictDistribution{ACCount: 2},
				CostSummary: ExperimentCostSummary{
					TotalTokenInput:   120,
					TotalTokenOutput:  60,
					TotalTokens:       180,
					TotalLLMLatencyMS: 220,
					TotalLatencyMS:    360,
					RunCount:          2,
				},
				Status: ExperimentStatusCompleted,
				Runs: []ExperimentRunOutput{
					{ProblemID: 1, Verdict: "AC", Status: ExperimentRunStatusSuccess},
					{ProblemID: 2, Verdict: "AC", Status: ExperimentRunStatusSuccess},
				},
			},
			{
				ID:                  12,
				Name:                "round-3",
				Model:               "mock-cpp17",
				PromptName:          prompt.CPP17MinimalSolvePromptName,
				AgentName:           agent.AnalyzeThenCodegenAgentName,
				TotalCount:          2,
				SuccessCount:        1,
				ACCount:             0,
				FailedCount:         1,
				VerdictDistribution: VerdictDistribution{RECount: 1, UnknownCount: 1},
				CostSummary:         ExperimentCostSummary{},
				Status:              ExperimentStatusCompleted,
				Runs: []ExperimentRunOutput{
					{ProblemID: 1, Verdict: "RE", Status: ExperimentRunStatusSuccess},
					{ProblemID: 2, Verdict: "", Status: ExperimentRunStatusFailed},
				},
			},
		},
	}
	service := NewExperimentRepeatService(repo, runner, "mock-default")

	output, err := service.Repeat(context.Background(), RepeatExperimentInput{
		Name:        "repeat-1",
		ProblemIDs:  []uint{1, 2},
		Model:       "mock-cpp17",
		PromptName:  prompt.CPP17MinimalSolvePromptName,
		AgentName:   agent.AnalyzeThenCodegenAgentName,
		RepeatCount: 3,
	})
	if err != nil {
		t.Fatalf("repeat returned error: %v", err)
	}

	if len(runner.runInputs) != 3 {
		t.Fatalf("expected 3 rounds, got %d", len(runner.runInputs))
	}
	for _, input := range runner.runInputs {
		if input.Model != "mock-cpp17" {
			t.Fatalf("expected repeat model to be passed to every round, got %+v", runner.runInputs)
		}
		if input.PromptName != prompt.CPP17MinimalSolvePromptName {
			t.Fatalf("expected repeat prompt to be passed to every round, got %+v", runner.runInputs)
		}
		if input.AgentName != agent.AnalyzeThenCodegenAgentName {
			t.Fatalf("expected repeat agent to be passed to every round, got %+v", runner.runInputs)
		}
	}
	if output.PromptName != prompt.CPP17MinimalSolvePromptName {
		t.Fatalf("expected repeat prompt name in output, got %q", output.PromptName)
	}
	if output.AgentName != agent.AnalyzeThenCodegenAgentName {
		t.Fatalf("expected repeat agent name in output, got %q", output.AgentName)
	}
	if output.TotalProblemCount != 2 || output.TotalRunCount != 6 {
		t.Fatalf("unexpected total counts: %+v", output)
	}
	if output.OverallACCount != 3 || output.OverallFailedCount != 2 {
		t.Fatalf("unexpected overall counts: %+v", output)
	}
	if output.BestRoundACCount != 2 || output.WorstRoundACCount != 0 {
		t.Fatalf("unexpected stability summary: %+v", output)
	}
	if output.CostSummary.RoundCount != 2 ||
		output.CostSummary.TotalTokenInput != 220 ||
		output.CostSummary.TotalTokenOutput != 100 ||
		output.CostSummary.TotalTokens != 320 ||
		output.CostSummary.TotalLLMLatencyMS != 420 ||
		output.CostSummary.TotalLatencyMS != 680 {
		t.Fatalf("unexpected repeat cost summary totals: %+v", output.CostSummary)
	}
	if output.CostSummary.AverageTokenInputPerRound != 110 ||
		output.CostSummary.AverageTokenOutputPerRound != 50 ||
		output.CostSummary.AverageTotalTokensPerRound != 160 ||
		output.CostSummary.AverageLLMLatencyMSPerRound != 210 ||
		output.CostSummary.AverageTotalLatencyMSPerRound != 340 {
		t.Fatalf("unexpected repeat cost summary averages: %+v", output.CostSummary)
	}
	if len(output.RoundSummaries) != 3 || output.RoundSummaries[2].VerdictDistribution.RECount != 1 {
		t.Fatalf("unexpected round summaries: %+v", output.RoundSummaries)
	}
	if len(output.ProblemSummaries) != 2 {
		t.Fatalf("expected 2 problem summaries, got %+v", output.ProblemSummaries)
	}
	if output.ProblemSummaries[0].ProblemID != 1 || output.ProblemSummaries[0].ACCount != 2 || output.ProblemSummaries[0].FailedCount != 1 {
		t.Fatalf("unexpected problem summary for problem 1: %+v", output.ProblemSummaries[0])
	}
	if output.ProblemSummaries[0].VerdictDistribution.ACCount != 2 || output.ProblemSummaries[0].VerdictDistribution.RECount != 1 {
		t.Fatalf("unexpected verdict distribution for problem 1: %+v", output.ProblemSummaries[0].VerdictDistribution)
	}
	if output.ProblemSummaries[1].ProblemID != 2 || output.ProblemSummaries[1].ACCount != 1 || output.ProblemSummaries[1].FailedCount != 2 {
		t.Fatalf("unexpected problem summary for problem 2: %+v", output.ProblemSummaries[1])
	}
	if output.ProblemSummaries[1].VerdictDistribution.WACount != 1 || output.ProblemSummaries[1].VerdictDistribution.ACCount != 1 || output.ProblemSummaries[1].VerdictDistribution.UnknownCount != 1 {
		t.Fatalf("unexpected verdict distribution for problem 2: %+v", output.ProblemSummaries[1].VerdictDistribution)
	}
	if len(output.MostUnstableProblems) != 2 {
		t.Fatalf("expected most unstable problems, got %+v", output.MostUnstableProblems)
	}
	if output.MostUnstableProblems[0].ProblemID != 2 || output.MostUnstableProblems[0].InstabilityScore != 1 {
		t.Fatalf("unexpected first unstable problem: %+v", output.MostUnstableProblems[0])
	}
	if output.MostUnstableProblems[1].ProblemID != 1 || output.MostUnstableProblems[1].InstabilityScore != 1 {
		t.Fatalf("unexpected second unstable problem: %+v", output.MostUnstableProblems[1])
	}
	if output.MostUnstableProblems[0].VerdictKindCount != 3 || output.MostUnstableProblems[1].VerdictKindCount != 2 {
		t.Fatalf("unexpected unstable verdict kind counts: %+v", output.MostUnstableProblems)
	}
}

func TestExperimentRepeatServiceRepeatMarksFailedOnRoundError(t *testing.T) {
	repo := &fakeExperimentRepeatRepository{}
	runner := &fakeExperimentRunner{
		runOutputs: []*ExperimentOutput{
			{ID: 10, Name: "round-1", Model: "mock-cpp17", TotalCount: 1, ACCount: 1, Status: ExperimentStatusCompleted},
			nil,
		},
		runErrors: []error{
			nil,
			errors.New("round 2 failed"),
		},
	}
	service := NewExperimentRepeatService(repo, runner, "mock-default")

	_, err := service.Repeat(context.Background(), RepeatExperimentInput{
		Name:        "repeat-2",
		ProblemIDs:  []uint{1},
		Model:       "mock-cpp17",
		PromptName:  prompt.StrictCPP17SolvePromptName,
		AgentName:   agent.AnalyzeThenCodegenAgentName,
		RepeatCount: 2,
	})
	if err == nil {
		t.Fatal("expected repeat to fail on hard round error")
	}
	if repo.updated == nil || repo.updated.Status != model.ExperimentRepeatStatusFailed || repo.updated.ErrorMessage == "" {
		t.Fatalf("expected failed repeat to be persisted, got %+v", repo.updated)
	}
	if len(runner.runInputs) != 2 || runner.runInputs[0].Model != "mock-cpp17" || runner.runInputs[1].Model != "mock-cpp17" {
		t.Fatalf("expected repeat to preserve model across rounds before failure, got %+v", runner.runInputs)
	}
	if runner.runInputs[0].PromptName != prompt.StrictCPP17SolvePromptName || runner.runInputs[1].PromptName != prompt.StrictCPP17SolvePromptName {
		t.Fatalf("expected repeat to preserve prompt across rounds before failure, got %+v", runner.runInputs)
	}
	if runner.runInputs[0].AgentName != agent.AnalyzeThenCodegenAgentName || runner.runInputs[1].AgentName != agent.AnalyzeThenCodegenAgentName {
		t.Fatalf("expected repeat to preserve agent across rounds before failure, got %+v", runner.runInputs)
	}
}

func TestExperimentRepeatServiceGet(t *testing.T) {
	repo := &fakeExperimentRepeatRepository{
		getByID: &model.ExperimentRepeat{
			BaseModel:     model.BaseModel{ID: 7, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
			Name:          "repeat-3",
			ModelName:     "mock-cpp17",
			PromptName:    prompt.DefaultSolvePromptName,
			AgentName:     agent.DirectCodegenAgentName,
			ProblemIDs:    "[1,2]",
			ExperimentIDs: "[10,11]",
			RepeatCount:   2,
			Status:        model.ExperimentRepeatStatusCompleted,
		},
	}
	runner := &fakeExperimentRunner{
		getMap: map[uint]*ExperimentOutput{
			10: {
				ID:                  10,
				PromptName:          prompt.DefaultSolvePromptName,
				AgentName:           agent.DirectCodegenAgentName,
				ACCount:             1,
				FailedCount:         1,
				VerdictDistribution: VerdictDistribution{ACCount: 1, WACount: 1},
				CostSummary: ExperimentCostSummary{
					TotalTokenInput:   90,
					TotalTokenOutput:  30,
					TotalTokens:       120,
					TotalLLMLatencyMS: 150,
					TotalLatencyMS:    260,
					RunCount:          2,
				},
				Status: ExperimentStatusCompleted,
				Runs: []ExperimentRunOutput{
					{ProblemID: 1, Verdict: "AC", Status: ExperimentRunStatusSuccess},
					{ProblemID: 2, Verdict: "WA", Status: ExperimentRunStatusSuccess},
				},
			},
			11: {
				ID:                  11,
				PromptName:          prompt.DefaultSolvePromptName,
				AgentName:           agent.DirectCodegenAgentName,
				ACCount:             2,
				FailedCount:         0,
				VerdictDistribution: VerdictDistribution{ACCount: 2},
				CostSummary:         ExperimentCostSummary{},
				Status:              ExperimentStatusCompleted,
				Runs: []ExperimentRunOutput{
					{ProblemID: 1, Verdict: "AC", Status: ExperimentRunStatusSuccess},
					{ProblemID: 2, Verdict: "AC", Status: ExperimentRunStatusSuccess},
				},
			},
		},
	}
	service := NewExperimentRepeatService(repo, runner, "mock-default")

	output, err := service.Get(context.Background(), 7)
	if err != nil {
		t.Fatalf("get repeat returned error: %v", err)
	}
	if len(output.ExperimentIDs) != 2 || output.ExperimentIDs[0] != 10 {
		t.Fatalf("unexpected experiment ids: %+v", output.ExperimentIDs)
	}
	if output.PromptName != prompt.DefaultSolvePromptName {
		t.Fatalf("unexpected prompt name on get: %q", output.PromptName)
	}
	if output.AgentName != agent.DirectCodegenAgentName {
		t.Fatalf("unexpected agent name on get: %q", output.AgentName)
	}
	if output.OverallACCount != 3 || output.OverallFailedCount != 1 {
		t.Fatalf("unexpected aggregate output: %+v", output)
	}
	if output.CostSummary.RoundCount != 1 ||
		output.CostSummary.TotalTokenInput != 90 ||
		output.CostSummary.TotalTokenOutput != 30 ||
		output.CostSummary.TotalTokens != 120 ||
		output.CostSummary.TotalLLMLatencyMS != 150 ||
		output.CostSummary.TotalLatencyMS != 260 ||
		output.CostSummary.AverageTokenInputPerRound != 90 ||
		output.CostSummary.AverageTokenOutputPerRound != 30 ||
		output.CostSummary.AverageTotalTokensPerRound != 120 ||
		output.CostSummary.AverageLLMLatencyMSPerRound != 150 ||
		output.CostSummary.AverageTotalLatencyMSPerRound != 260 {
		t.Fatalf("unexpected repeat cost summary on get: %+v", output.CostSummary)
	}
	if len(output.ProblemSummaries) != 2 || output.ProblemSummaries[0].ProblemID != 1 || output.ProblemSummaries[1].ProblemID != 2 {
		t.Fatalf("unexpected problem summaries: %+v", output.ProblemSummaries)
	}
	if output.ProblemSummaries[1].ACRate != 0.5 {
		t.Fatalf("unexpected problem 2 ac rate: %+v", output.ProblemSummaries[1])
	}
	if len(output.MostUnstableProblems) != 2 || output.MostUnstableProblems[0].ProblemID != 2 {
		t.Fatalf("unexpected most unstable problems: %+v", output.MostUnstableProblems)
	}
}
