package tooling

import (
	"context"
	"testing"

	"ai-for-oj/internal/judge"
	"ai-for-oj/internal/model"
)

type fakeSampleJudgeEngine struct {
	request judge.Request
	result  judge.Result
	err     error
}

func (e *fakeSampleJudgeEngine) Judge(_ context.Context, req judge.Request) (judge.Result, error) {
	e.request = req
	return e.result, e.err
}

func TestSampleJudgeToolRunsOnlySampleTestCases(t *testing.T) {
	engine := &fakeSampleJudgeEngine{
		result: judge.Result{
			Verdict:      judge.VerdictWrongAnswer,
			RuntimeMS:    7,
			PassedCount:  0,
			TotalCount:   1,
			RunStdout:    "actual",
			RunStderr:    "stderr",
			ExitCode:     0,
			ErrorMessage: "wrong answer: output mismatch",
		},
	}
	tool := NewSampleJudgeTool(engine)
	problem := &model.Problem{
		BaseModel: model.BaseModel{ID: 9},
		TestCases: []model.TestCase{
			{CreatedModel: model.CreatedModel{ID: 1}, Input: "sample", ExpectedOutput: "expected", IsSample: true},
			{CreatedModel: model.CreatedModel{ID: 2}, Input: "hidden", ExpectedOutput: "secret", IsSample: false},
		},
	}

	output, err := tool.Run(context.Background(), CallInput{
		Problem:    problem,
		SourceCode: "int main(){return 0;}",
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if engine.request.Problem != problem {
		t.Fatalf("expected original problem in judge request")
	}
	if len(engine.request.TestCases) != 1 || !engine.request.TestCases[0].IsSample {
		t.Fatalf("expected only sample testcases, got %+v", engine.request.TestCases)
	}
	if engine.request.SourceCode != "int main(){return 0;}" || engine.request.Language != model.LanguageCPP17 {
		t.Fatalf("unexpected judge request: %+v", engine.request)
	}
	if output.ToolName != SampleJudgeToolName || output.Status != CallStatusFailed {
		t.Fatalf("unexpected output status: %+v", output)
	}
	if output.Metadata["verdict"] != judge.VerdictWrongAnswer ||
		output.Metadata["stdout"] != "actual" ||
		output.Metadata["stderr"] != "stderr" ||
		output.Metadata["runtime_ms"] != 7 ||
		output.Metadata["error_message"] != "wrong answer: output mismatch" {
		t.Fatalf("expected judge metadata, got %+v", output.Metadata)
	}
}

func TestSampleJudgeToolSkipsWhenNoSamples(t *testing.T) {
	engine := &fakeSampleJudgeEngine{}
	tool := NewSampleJudgeTool(engine)

	output, err := tool.Run(context.Background(), CallInput{
		Problem: &model.Problem{
			TestCases: []model.TestCase{{Input: "hidden", ExpectedOutput: "secret", IsSample: false}},
		},
		SourceCode: "int main(){return 0;}",
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(engine.request.TestCases) != 0 {
		t.Fatalf("expected judge not to be called, got %+v", engine.request)
	}
	if output.Status != CallStatusSkipped {
		t.Fatalf("expected skipped result, got %+v", output)
	}
}
