package judge

import (
	"context"
	"testing"

	"ai-for-oj/internal/model"
	"ai-for-oj/internal/sandbox"
)

func TestJudgeAccepted(t *testing.T) {
	engine := NewEngine(sandbox.NewMockSandbox())

	result, err := engine.Judge(context.Background(), Request{
		Problem: &model.Problem{
			TimeLimitMS:   1000,
			MemoryLimitMB: 256,
		},
		TestCases: []model.TestCase{
			{Input: "1 2", ExpectedOutput: "1 2"},
			{Input: "hello", ExpectedOutput: "hello"},
		},
		Language:   model.LanguageCPP17,
		SourceCode: "int main() { return 0; }",
	})
	if err != nil {
		t.Fatalf("judge returned error: %v", err)
	}

	if result.Verdict != VerdictAccepted {
		t.Fatalf("expected verdict %s, got %s", VerdictAccepted, result.Verdict)
	}

	if result.PassedCount != 2 || result.TotalCount != 2 {
		t.Fatalf("expected passed/total 2/2, got %d/%d", result.PassedCount, result.TotalCount)
	}

	if len(result.TestCaseResults) != 2 || result.TestCaseResults[0].Verdict != VerdictAccepted {
		t.Fatalf("expected testcase results to be recorded for AC, got %+v", result.TestCaseResults)
	}
}

func TestJudgeReturnsUnjudgeableWhenProblemHasNoTestCases(t *testing.T) {
	engine := NewEngine(sandbox.NewMockSandbox())

	result, err := engine.Judge(context.Background(), Request{
		Problem: &model.Problem{
			TimeLimitMS:   1000,
			MemoryLimitMB: 256,
		},
		TestCases:  nil,
		Language:   model.LanguageCPP17,
		SourceCode: "int main() { return 0; }",
	})
	if err != nil {
		t.Fatalf("judge returned error: %v", err)
	}

	if result.Verdict != VerdictUnjudgeable {
		t.Fatalf("expected verdict %s, got %s", VerdictUnjudgeable, result.Verdict)
	}

	if result.ExecStage != "validate" {
		t.Fatalf("expected exec stage validate, got %s", result.ExecStage)
	}

	if result.ErrorMessage == "" {
		t.Fatal("expected clear error message for unjudgeable submission")
	}

	if result.PassedCount != 0 || result.TotalCount != 0 {
		t.Fatalf("expected passed/total 0/0, got %d/%d", result.PassedCount, result.TotalCount)
	}

	if len(result.TestCaseResults) != 0 {
		t.Fatalf("expected no testcase results, got %+v", result.TestCaseResults)
	}
}

func TestJudgeCompileError(t *testing.T) {
	engine := NewEngine(sandbox.NewMockSandbox())

	result, err := engine.Judge(context.Background(), Request{
		Problem: &model.Problem{
			TimeLimitMS:   1000,
			MemoryLimitMB: 256,
		},
		TestCases: []model.TestCase{
			{Input: "1", ExpectedOutput: "1"},
		},
		Language:   model.LanguageCPP17,
		SourceCode: "MOCK_CE",
	})
	if err != nil {
		t.Fatalf("judge returned error: %v", err)
	}

	if result.Verdict != VerdictCompileError {
		t.Fatalf("expected verdict %s, got %s", VerdictCompileError, result.Verdict)
	}

	if result.CompileStderr == "" || result.ExecStage != "compile" || result.ExitCode == 0 {
		t.Fatalf("expected compile observability fields to be populated, got %+v", result)
	}
}

func TestJudgeWrongAnswer(t *testing.T) {
	engine := NewEngine(sandbox.NewMockSandbox())

	result, err := engine.Judge(context.Background(), Request{
		Problem: &model.Problem{
			TimeLimitMS:   1000,
			MemoryLimitMB: 256,
		},
		TestCases: []model.TestCase{
			{Input: "1", ExpectedOutput: "1"},
		},
		Language:   model.LanguageCPP17,
		SourceCode: "MOCK_WA",
	})
	if err != nil {
		t.Fatalf("judge returned error: %v", err)
	}

	if result.Verdict != VerdictWrongAnswer {
		t.Fatalf("expected verdict %s, got %s", VerdictWrongAnswer, result.Verdict)
	}

	if len(result.TestCaseResults) != 1 || result.TestCaseResults[0].Verdict != VerdictWrongAnswer {
		t.Fatalf("expected testcase WA summary, got %+v", result.TestCaseResults)
	}
}

func TestJudgeTimeLimitExceeded(t *testing.T) {
	engine := NewEngine(sandbox.NewMockSandbox())

	result, err := engine.Judge(context.Background(), Request{
		Problem: &model.Problem{
			TimeLimitMS:   1000,
			MemoryLimitMB: 256,
		},
		TestCases: []model.TestCase{
			{Input: "1", ExpectedOutput: "1"},
		},
		Language:   model.LanguageCPP17,
		SourceCode: "MOCK_TLE",
	})
	if err != nil {
		t.Fatalf("judge returned error: %v", err)
	}

	if result.Verdict != VerdictTimeLimitExceeded {
		t.Fatalf("expected verdict %s, got %s", VerdictTimeLimitExceeded, result.Verdict)
	}

	if !result.TimedOut || result.ExecStage != "run" {
		t.Fatalf("expected timeout observability fields, got %+v", result)
	}

	if len(result.TestCaseResults) != 1 || !result.TestCaseResults[0].TimedOut {
		t.Fatalf("expected testcase timeout summary, got %+v", result.TestCaseResults)
	}
}

func TestJudgeMemoryLimitExceeded(t *testing.T) {
	engine := NewEngine(sandbox.NewMockSandbox())

	result, err := engine.Judge(context.Background(), Request{
		Problem: &model.Problem{
			TimeLimitMS:   1000,
			MemoryLimitMB: 64,
		},
		TestCases: []model.TestCase{
			{Input: "1", ExpectedOutput: "1"},
		},
		Language:   model.LanguageCPP17,
		SourceCode: "MOCK_MLE",
	})
	if err != nil {
		t.Fatalf("judge returned error: %v", err)
	}

	if result.Verdict != VerdictMemoryLimitExceeded || !result.MemoryExceeded {
		t.Fatalf("expected MLE result, got %+v", result)
	}
	if result.MemoryKB < 64*1024 {
		t.Fatalf("expected memory usage at least limit, got %d", result.MemoryKB)
	}
	if len(result.TestCaseResults) != 1 ||
		result.TestCaseResults[0].Verdict != VerdictMemoryLimitExceeded ||
		!result.TestCaseResults[0].MemoryExceeded {
		t.Fatalf("expected testcase MLE summary, got %+v", result.TestCaseResults)
	}
}

func TestJudgeRuntimeError(t *testing.T) {
	engine := NewEngine(sandbox.NewMockSandbox())

	result, err := engine.Judge(context.Background(), Request{
		Problem: &model.Problem{
			TimeLimitMS:   1000,
			MemoryLimitMB: 256,
		},
		TestCases: []model.TestCase{
			{Input: "1", ExpectedOutput: "1"},
		},
		Language:   model.LanguageCPP17,
		SourceCode: "MOCK_RE",
	})
	if err != nil {
		t.Fatalf("judge returned error: %v", err)
	}

	if result.Verdict != VerdictRuntimeError {
		t.Fatalf("expected verdict %s, got %s", VerdictRuntimeError, result.Verdict)
	}

	if result.RunStderr == "" || result.ExitCode == 0 {
		t.Fatalf("expected runtime observability fields, got %+v", result)
	}

	if len(result.TestCaseResults) != 1 || result.TestCaseResults[0].Verdict != VerdictRuntimeError {
		t.Fatalf("expected testcase runtime error summary, got %+v", result.TestCaseResults)
	}
}
