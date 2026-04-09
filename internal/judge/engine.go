package judge

import (
	"context"
	"fmt"
	"strings"

	"ai-for-oj/internal/sandbox"
)

type Sandbox interface {
	sandbox.Executor
}

type JudgeEngine struct {
	sandbox Sandbox
}

func NewEngine(sandbox Sandbox) *JudgeEngine {
	return &JudgeEngine{sandbox: sandbox}
}

func (e *JudgeEngine) Judge(ctx context.Context, req Request) (Result, error) {
	result := Result{
		Verdict:    VerdictAccepted,
		TotalCount: len(req.TestCases),
	}

	compileResult, err := e.sandbox.Compile(ctx, sandbox.CompileRequest{
		Language:   req.Language,
		SourceCode: req.SourceCode,
	})
	if err != nil {
		return Result{}, fmt.Errorf("compile submission: %w", err)
	}

	if !compileResult.Success {
		result.Verdict = VerdictCompileError
		result.ExecStage = "compile"
		result.CompileStderr = compileResult.Stderr
		result.ExitCode = compileResult.ExitCode
		result.ErrorMessage = pickErrorMessage(compileResult.ErrorMessage, compileResult.Stderr)
		return result, nil
	}
	if cleaner, ok := e.sandbox.(sandbox.Cleaner); ok {
		defer func() {
			_ = cleaner.Cleanup(ctx, compileResult.ArtifactID)
		}()
	}

	for index, testCase := range req.TestCases {
		runResult, err := e.sandbox.Run(ctx, sandbox.RunRequest{
			Language:      req.Language,
			ArtifactID:    compileResult.ArtifactID,
			Input:         testCase.Input,
			TimeLimitMS:   req.Problem.TimeLimitMS,
			MemoryLimitMB: req.Problem.MemoryLimitMB,
		})
		if err != nil {
			return Result{}, fmt.Errorf("run test case %d: %w", testCase.ID, err)
		}

		result.ExecStage = "run"
		result.RunStdout = runResult.Stdout
		result.RunStderr = runResult.Stderr
		result.ExitCode = runResult.ExitCode
		result.TimedOut = runResult.TimedOut
		result.RuntimeMS = max(result.RuntimeMS, runResult.RuntimeMS)
		result.MemoryKB = max(result.MemoryKB, runResult.MemoryKB)

		caseResult := TestCaseResult{
			TestCaseID: testCase.ID,
			CaseIndex:  index + 1,
			RuntimeMS:  runResult.RuntimeMS,
			Stdout:     runResult.Stdout,
			Stderr:     runResult.Stderr,
			ExitCode:   runResult.ExitCode,
			TimedOut:   runResult.TimedOut,
		}

		switch {
		case runResult.TimedOut:
			result.Verdict = VerdictTimeLimitExceeded
			caseResult.Verdict = VerdictTimeLimitExceeded
			result.TestCaseResults = append(result.TestCaseResults, caseResult)
			result.ErrorMessage = pickErrorMessage("time limit exceeded", runResult.ErrorMessage, runResult.Stderr)
			return result, nil
		case runResult.RuntimeError:
			result.Verdict = VerdictRuntimeError
			caseResult.Verdict = VerdictRuntimeError
			result.TestCaseResults = append(result.TestCaseResults, caseResult)
			result.ErrorMessage = pickErrorMessage(runResult.ErrorMessage, runResult.Stderr)
			return result, nil
		case !equalOutput(runResult.Stdout, testCase.ExpectedOutput):
			result.Verdict = VerdictWrongAnswer
			caseResult.Verdict = VerdictWrongAnswer
			result.TestCaseResults = append(result.TestCaseResults, caseResult)
			result.ErrorMessage = "wrong answer: output mismatch"
			return result, nil
		default:
			caseResult.Verdict = VerdictAccepted
			result.TestCaseResults = append(result.TestCaseResults, caseResult)
			result.PassedCount++
		}
	}

	return result, nil
}

func equalOutput(actual, expected string) bool {
	return normalizeOutput(actual) == normalizeOutput(expected)
}

func normalizeOutput(value string) string {
	lines := strings.Split(value, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func pickErrorMessage(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}
