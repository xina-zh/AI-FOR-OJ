package tooling

import (
	"context"
	"strings"
	"testing"

	"ai-for-oj/internal/judge"
	"ai-for-oj/internal/model"
	"ai-for-oj/internal/repository"
)

type fakeSampleProblemRepository struct {
	problem *model.Problem
	err     error
}

func (r fakeSampleProblemRepository) Create(context.Context, *model.Problem) error {
	return nil
}

func (r fakeSampleProblemRepository) List(context.Context) ([]model.Problem, error) {
	return nil, nil
}

func (r fakeSampleProblemRepository) GetByID(context.Context, uint) (*model.Problem, error) {
	return r.problem, r.err
}

func (r fakeSampleProblemRepository) GetByIDWithTestCases(context.Context, uint) (*model.Problem, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.problem, nil
}

type fakeSampleJudgeEngine struct {
	req judge.Request
	res judge.Result
	err error
}

func (e *fakeSampleJudgeEngine) Judge(_ context.Context, req judge.Request) (judge.Result, error) {
	e.req = req
	return e.res, e.err
}

func TestSampleJudgeToolRunsOnlySampleCases(t *testing.T) {
	engine := &fakeSampleJudgeEngine{
		res: judge.Result{Verdict: "AC", PassedCount: 1, TotalCount: 1},
	}
	tool := NewSampleJudgeTool(fakeSampleProblemRepository{
		problem: &model.Problem{
			BaseModel: model.BaseModel{ID: 7},
			TestCases: []model.TestCase{
				{CreatedModel: model.CreatedModel{ID: 1}, Input: "1\n", ExpectedOutput: "1\n", IsSample: true},
				{CreatedModel: model.CreatedModel{ID: 2}, Input: "2\n", ExpectedOutput: "2\n", IsSample: false},
			},
		},
	}, engine)

	result, err := tool.Execute(context.Background(), CallInput{ProblemID: 7, SourceCode: "int main(){}"})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if result.Status != CallStatusOK {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(engine.req.TestCases) != 1 || !engine.req.TestCases[0].IsSample {
		t.Fatalf("expected only sample cases, got %+v", engine.req.TestCases)
	}
	if !strings.Contains(result.Summary, "verdict=AC") {
		t.Fatalf("expected verdict summary, got %q", result.Summary)
	}
}

func TestSampleJudgeToolSkipsWhenNoSamples(t *testing.T) {
	tool := NewSampleJudgeTool(fakeSampleProblemRepository{
		problem: &model.Problem{BaseModel: model.BaseModel{ID: 7}},
	}, &fakeSampleJudgeEngine{})

	result, err := tool.Execute(context.Background(), CallInput{ProblemID: 7, SourceCode: "int main(){}"})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if result.Status != CallStatusSkipped {
		t.Fatalf("expected skipped result, got %+v", result)
	}
}

func TestSampleJudgeToolReturnsRepositoryError(t *testing.T) {
	tool := NewSampleJudgeTool(fakeSampleProblemRepository{err: repository.ErrProblemNotFound}, &fakeSampleJudgeEngine{})
	_, err := tool.Execute(context.Background(), CallInput{ProblemID: 7, SourceCode: "int main(){}"})
	if err == nil {
		t.Fatal("expected repository error")
	}
}
