package service

import (
	"context"
	"testing"
	"time"

	"ai-for-oj/internal/judge"
	"ai-for-oj/internal/model"
	"ai-for-oj/internal/repository"
	"ai-for-oj/internal/sandbox"
)

type fakeProblemRepository struct {
	problem *model.Problem
	list    []model.Problem
	err     error
}

func (r fakeProblemRepository) Create(context.Context, *model.Problem) error {
	return nil
}

func (r fakeProblemRepository) List(context.Context) ([]model.Problem, error) {
	return r.list, r.err
}

func (r fakeProblemRepository) GetByID(context.Context, uint) (*model.Problem, error) {
	return r.problem, r.err
}

func (r fakeProblemRepository) GetByIDWithTestCases(context.Context, uint) (*model.Problem, error) {
	return r.problem, r.err
}

type fakeSubmissionRepository struct {
	submission        *model.Submission
	judgeResult       *model.JudgeResult
	testCaseResults   []model.SubmissionTestCaseResult
	submissionCounter uint
	list              []model.Submission
	getResult         *model.Submission
	stats             []repository.SubmissionProblemStatsRow
	err               error
}

func (r *fakeSubmissionRepository) Create(_ context.Context, submission *model.Submission) error {
	if r.err != nil {
		return r.err
	}
	r.submissionCounter++
	submission.ID = r.submissionCounter
	r.submission = submission
	return nil
}

func (r *fakeSubmissionRepository) CreateJudgeResult(_ context.Context, result *model.JudgeResult) error {
	if r.err != nil {
		return r.err
	}
	r.judgeResult = result
	return nil
}

func (r *fakeSubmissionRepository) CreateTestCaseResults(_ context.Context, results []model.SubmissionTestCaseResult) error {
	if r.err != nil {
		return r.err
	}
	r.testCaseResults = append(r.testCaseResults, results...)
	return nil
}

func (r *fakeSubmissionRepository) List(context.Context, repository.SubmissionListQuery) ([]model.Submission, int64, error) {
	return r.list, int64(len(r.list)), r.err
}

func (r *fakeSubmissionRepository) GetByID(context.Context, uint) (*model.Submission, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.getResult == nil {
		return nil, repository.ErrSubmissionNotFound
	}
	return r.getResult, nil
}

func (r *fakeSubmissionRepository) AggregateByProblem(context.Context) ([]repository.SubmissionProblemStatsRow, error) {
	return r.stats, r.err
}

type fakeJudgeEngine struct {
	result judge.Result
	err    error
}

func (e fakeJudgeEngine) Judge(context.Context, judge.Request) (judge.Result, error) {
	return e.result, e.err
}

func TestJudgeSubmissionServiceSubmit(t *testing.T) {
	problemRepo := fakeProblemRepository{
		problem: &model.Problem{
			BaseModel: model.BaseModel{ID: 42},
			TestCases: []model.TestCase{
				{CreatedModel: model.CreatedModel{ID: 1}, Input: "1", ExpectedOutput: "1"},
			},
		},
	}
	submissionRepo := &fakeSubmissionRepository{}
	engine := fakeJudgeEngine{
		result: judge.Result{
			Verdict:     judge.VerdictAccepted,
			RuntimeMS:   1,
			MemoryKB:    1024,
			PassedCount: 1,
			TotalCount:  1,
			TestCaseResults: []judge.TestCaseResult{
				{
					TestCaseID: 1,
					CaseIndex:  1,
					Verdict:    judge.VerdictAccepted,
					RuntimeMS:  1,
					Stdout:     "1",
					ExitCode:   0,
				},
			},
		},
	}

	service := NewJudgeSubmissionService(problemRepo, submissionRepo, engine)
	output, err := service.Submit(context.Background(), JudgeSubmissionInput{
		ProblemID:  42,
		SourceCode: "int main() { return 0; }",
		Language:   model.LanguageCPP17,
	})
	if err != nil {
		t.Fatalf("submit returned error: %v", err)
	}

	if output.SubmissionID == 0 {
		t.Fatal("expected submission id to be assigned")
	}

	if submissionRepo.submission == nil {
		t.Fatal("expected submission to be persisted")
	}

	if submissionRepo.submission.SourceType != model.SourceTypeHuman {
		t.Fatalf("expected source type %s, got %s", model.SourceTypeHuman, submissionRepo.submission.SourceType)
	}

	if submissionRepo.judgeResult == nil {
		t.Fatal("expected judge result to be persisted")
	}

	if submissionRepo.judgeResult.Verdict != judge.VerdictAccepted {
		t.Fatalf("expected verdict %s, got %s", judge.VerdictAccepted, submissionRepo.judgeResult.Verdict)
	}

	if len(submissionRepo.testCaseResults) != 1 {
		t.Fatalf("expected 1 testcase result to be persisted, got %d", len(submissionRepo.testCaseResults))
	}

	if submissionRepo.testCaseResults[0].CaseIndex != 1 || submissionRepo.testCaseResults[0].Verdict != judge.VerdictAccepted {
		t.Fatalf("expected testcase result summary to be persisted, got %+v", submissionRepo.testCaseResults[0])
	}
}

func TestJudgeSubmissionServiceSubmitRejectsUnsupportedLanguage(t *testing.T) {
	service := NewJudgeSubmissionService(
		fakeProblemRepository{},
		&fakeSubmissionRepository{},
		fakeJudgeEngine{},
	)

	_, err := service.Submit(context.Background(), JudgeSubmissionInput{
		ProblemID:  1,
		SourceCode: "print(1)",
		Language:   "python",
	})
	if err != ErrUnsupportedLanguage {
		t.Fatalf("expected err %v, got %v", ErrUnsupportedLanguage, err)
	}
}

func TestJudgeSubmissionServiceSubmitReturnsProblemNotFound(t *testing.T) {
	service := NewJudgeSubmissionService(
		fakeProblemRepository{err: repository.ErrProblemNotFound},
		&fakeSubmissionRepository{},
		fakeJudgeEngine{},
	)

	_, err := service.Submit(context.Background(), JudgeSubmissionInput{
		ProblemID:  999,
		SourceCode: "int main() { return 0; }",
		Language:   model.LanguageCPP17,
	})
	if err != repository.ErrProblemNotFound {
		t.Fatalf("expected err %v, got %v", repository.ErrProblemNotFound, err)
	}
}

func TestJudgeSubmissionServiceSubmitReturnsUnjudgeableWhenProblemHasNoTestCases(t *testing.T) {
	problemRepo := fakeProblemRepository{
		problem: &model.Problem{
			BaseModel: model.BaseModel{ID: 7},
			TestCases: nil,
		},
	}
	submissionRepo := &fakeSubmissionRepository{}
	service := NewJudgeSubmissionService(
		problemRepo,
		submissionRepo,
		judge.NewEngine(sandbox.NewMockSandbox()),
	)

	output, err := service.Submit(context.Background(), JudgeSubmissionInput{
		ProblemID:  7,
		SourceCode: "int main() { return 0; }",
		Language:   model.LanguageCPP17,
	})
	if err != nil {
		t.Fatalf("submit returned error: %v", err)
	}

	if output.Verdict != judge.VerdictUnjudgeable {
		t.Fatalf("expected verdict %s, got %s", judge.VerdictUnjudgeable, output.Verdict)
	}

	if output.ErrorMessage == "" {
		t.Fatal("expected clear error message for unjudgeable submission")
	}

	if submissionRepo.judgeResult == nil {
		t.Fatal("expected judge result to be persisted")
	}

	if submissionRepo.judgeResult.Verdict != judge.VerdictUnjudgeable {
		t.Fatalf("expected persisted verdict %s, got %s", judge.VerdictUnjudgeable, submissionRepo.judgeResult.Verdict)
	}

	if submissionRepo.judgeResult.ExecStage != "validate" {
		t.Fatalf("expected exec stage validate, got %s", submissionRepo.judgeResult.ExecStage)
	}

	if len(submissionRepo.testCaseResults) != 0 {
		t.Fatalf("expected no testcase results to be persisted, got %+v", submissionRepo.testCaseResults)
	}
}

func TestFakeSubmissionRepositorySupportsStatsShape(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeSubmissionRepository{
		stats: []repository.SubmissionProblemStatsRow{
			{
				ProblemID:          1,
				ProblemTitle:       "Echo",
				TotalSubmissions:   5,
				ACCount:            2,
				WACount:            1,
				CECount:            1,
				RECount:            1,
				TLECount:           0,
				LatestSubmissionAt: &now,
			},
		},
	}

	stats, err := repo.AggregateByProblem(context.Background())
	if err != nil {
		t.Fatalf("aggregate returned error: %v", err)
	}
	if len(stats) != 1 || stats[0].ProblemID != 1 {
		t.Fatalf("expected one stats row for problem 1, got %+v", stats)
	}
}
