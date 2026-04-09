package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"ai-for-oj/internal/model"
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
				TotalCount:          3,
				SuccessCount:        3,
				ACCount:             1,
				FailedCount:         0,
				VerdictDistribution: VerdictDistribution{ACCount: 1, WACount: 1, CECount: 1},
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
				TotalCount:          3,
				SuccessCount:        3,
				ACCount:             2,
				FailedCount:         0,
				VerdictDistribution: VerdictDistribution{ACCount: 2, RECount: 1},
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
				TotalCount:          3,
				SuccessCount:        3,
				ACCount:             1,
				FailedCount:         0,
				VerdictDistribution: VerdictDistribution{ACCount: 1, WACount: 1, CECount: 1},
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
				TotalCount:          3,
				SuccessCount:        3,
				ACCount:             2,
				FailedCount:         0,
				VerdictDistribution: VerdictDistribution{ACCount: 2, RECount: 1},
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
		Name:           "compare-1",
		ProblemIDs:     []uint{1, 2, 3},
		BaselineModel:  "mock-a",
		CandidateModel: "mock-b",
	})
	if err != nil {
		t.Fatalf("compare returned error: %v", err)
	}

	if len(runner.runInputs) != 2 {
		t.Fatalf("expected baseline and candidate to run, got %d calls", len(runner.runInputs))
	}

	if output.CompareDimension != ExperimentCompareDimensionModel {
		t.Fatalf("expected compare dimension model, got %s", output.CompareDimension)
	}

	if output.DeltaACCount != 1 || output.DeltaFailedCount != 0 {
		t.Fatalf("unexpected delta summary: %+v", output)
	}
	if output.BaselineDistribution.WACount != 1 || output.CandidateDistribution.ACCount != 2 || output.DeltaDistribution.ACCount != 1 || output.DeltaDistribution.WACount != -1 {
		t.Fatalf("unexpected verdict distribution delta: %+v", output)
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
		Name:           "compare-2",
		ProblemIDs:     []uint{1, 2},
		BaselineModel:  "mock-a",
		CandidateModel: "mock-b",
	})
	if err == nil {
		t.Fatal("expected compare to return error when candidate run fails")
	}

	if repo.updated == nil || repo.updated.Status != model.ExperimentCompareStatusFailed {
		t.Fatalf("expected failed compare to be persisted, got %+v", repo.updated)
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
				VerdictDistribution: VerdictDistribution{ACCount: 1, WACount: 1},
				Runs: []ExperimentRunOutput{
					{ProblemID: 1, Verdict: "WA", Status: ExperimentRunStatusSuccess, SubmissionID: uintPtr(101)},
					{ProblemID: 2, Verdict: "AC", Status: ExperimentRunStatusSuccess, SubmissionID: uintPtr(102)},
				},
			},
			20: {
				ID:                  20,
				Name:                "candidate",
				VerdictDistribution: VerdictDistribution{ACCount: 1, CECount: 1},
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

	if len(output.ProblemIDs) != 2 || output.ProblemIDs[0] != 1 {
		t.Fatalf("unexpected problem ids: %+v", output.ProblemIDs)
	}
	if output.DeltaDistribution.CECount != 1 || output.DeltaDistribution.WACount != -1 {
		t.Fatalf("unexpected delta distribution on get: %+v", output.DeltaDistribution)
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

func uintPtr(value uint) *uint {
	return &value
}
