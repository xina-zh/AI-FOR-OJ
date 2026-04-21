package service

import (
	"context"
	"testing"

	"ai-for-oj/internal/model"
	"ai-for-oj/internal/repository"
)

type fakeProblemWriteRepository struct {
	problem   *model.Problem
	list      []model.Problem
	err       error
	deletedID uint
}

func (r *fakeProblemWriteRepository) Create(_ context.Context, problem *model.Problem) error {
	problem.ID = 1
	r.problem = problem
	return r.err
}

func (r *fakeProblemWriteRepository) List(context.Context) ([]model.Problem, error) {
	return r.list, r.err
}

func (r *fakeProblemWriteRepository) GetByID(_ context.Context, problemID uint) (*model.Problem, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.problem == nil || r.problem.ID != problemID {
		return nil, repository.ErrProblemNotFound
	}
	return r.problem, nil
}

func (r *fakeProblemWriteRepository) GetByIDWithTestCases(_ context.Context, problemID uint) (*model.Problem, error) {
	return r.GetByID(context.Background(), problemID)
}

func (r *fakeProblemWriteRepository) Delete(_ context.Context, problemID uint) error {
	if r.err != nil {
		return r.err
	}
	if r.problem == nil || r.problem.ID != problemID {
		return repository.ErrProblemNotFound
	}
	r.deletedID = problemID
	r.problem = nil
	return nil
}

type fakeTestCaseRepository struct {
	testCase *model.TestCase
	list     []model.TestCase
	err      error
}

func (r *fakeTestCaseRepository) Create(_ context.Context, testCase *model.TestCase) error {
	testCase.ID = 1
	r.testCase = testCase
	return r.err
}

func (r *fakeTestCaseRepository) ListByProblemID(context.Context, uint) ([]model.TestCase, error) {
	return r.list, r.err
}

func TestProblemServiceCreate(t *testing.T) {
	problemRepo := &fakeProblemWriteRepository{}
	testCaseRepo := &fakeTestCaseRepository{}
	service := NewProblemService(problemRepo, testCaseRepo)

	output, err := service.Create(context.Background(), CreateProblemInput{
		Title:         "A + B",
		Description:   "desc",
		InputSpec:     "input",
		OutputSpec:    "output",
		Samples:       "[]",
		TimeLimitMS:   1000,
		MemoryLimitMB: 256,
		Difficulty:    "easy",
		Tags:          "math",
	})
	if err != nil {
		t.Fatalf("create problem returned error: %v", err)
	}

	if output.ID == 0 {
		t.Fatal("expected created problem id")
	}

	if problemRepo.problem == nil || problemRepo.problem.Title != "A + B" {
		t.Fatal("expected problem to be persisted")
	}
}

func TestProblemServiceList(t *testing.T) {
	problemRepo := &fakeProblemWriteRepository{
		list: []model.Problem{
			{BaseModel: model.BaseModel{ID: 1}, Title: "A"},
			{BaseModel: model.BaseModel{ID: 2}, Title: "B"},
		},
	}
	service := NewProblemService(problemRepo, &fakeTestCaseRepository{})

	outputs, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("list problems returned error: %v", err)
	}

	if len(outputs) != 2 {
		t.Fatalf("expected 2 problems, got %d", len(outputs))
	}
}

func TestProblemServiceCreateTestCase(t *testing.T) {
	problemRepo := &fakeProblemWriteRepository{
		problem: &model.Problem{BaseModel: model.BaseModel{ID: 42}},
	}
	testCaseRepo := &fakeTestCaseRepository{}
	service := NewProblemService(problemRepo, testCaseRepo)

	output, err := service.CreateTestCase(context.Background(), CreateTestCaseInput{
		ProblemID:      42,
		Input:          "1 2",
		ExpectedOutput: "3",
		IsSample:       true,
	})
	if err != nil {
		t.Fatalf("create test case returned error: %v", err)
	}

	if output.ID == 0 {
		t.Fatal("expected created test case id")
	}

	if testCaseRepo.testCase == nil || testCaseRepo.testCase.ProblemID != 42 {
		t.Fatal("expected test case to be persisted")
	}
}

func TestProblemServiceListTestCasesReturnsProblemNotFound(t *testing.T) {
	service := NewProblemService(
		&fakeProblemWriteRepository{err: repository.ErrProblemNotFound},
		&fakeTestCaseRepository{},
	)

	_, err := service.ListTestCases(context.Background(), 999)
	if err != repository.ErrProblemNotFound {
		t.Fatalf("expected err %v, got %v", repository.ErrProblemNotFound, err)
	}
}

func TestProblemServiceDelete(t *testing.T) {
	problemRepo := &fakeProblemWriteRepository{
		problem: &model.Problem{BaseModel: model.BaseModel{ID: 42}, Title: "delete me"},
	}
	service := NewProblemService(problemRepo, &fakeTestCaseRepository{})

	err := service.Delete(context.Background(), 42)
	if err != nil {
		t.Fatalf("delete problem returned error: %v", err)
	}

	if problemRepo.deletedID != 42 {
		t.Fatalf("expected deleted problem 42, got %d", problemRepo.deletedID)
	}
}

func TestProblemServiceDeleteReturnsProblemNotFound(t *testing.T) {
	service := NewProblemService(&fakeProblemWriteRepository{}, &fakeTestCaseRepository{})

	err := service.Delete(context.Background(), 999)
	if err != repository.ErrProblemNotFound {
		t.Fatalf("expected err %v, got %v", repository.ErrProblemNotFound, err)
	}
}
