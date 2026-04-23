package tooling

import (
	"context"
	"errors"
	"fmt"

	"ai-for-oj/internal/judge"
	"ai-for-oj/internal/model"
)

const SampleJudgeToolName = "sample_judge"

var ErrSampleJudgeProblemRequired = errors.New("sample judge problem is required")

type SampleJudgeTool struct {
	engine judge.Engine
}

func NewSampleJudgeTool(engine judge.Engine) *SampleJudgeTool {
	return &SampleJudgeTool{engine: engine}
}

func (t *SampleJudgeTool) Name() string {
	return SampleJudgeToolName
}

func (t *SampleJudgeTool) Run(ctx context.Context, input CallInput) (CallOutput, error) {
	if input.Problem == nil {
		return CallOutput{}, ErrSampleJudgeProblemRequired
	}

	samples := make([]model.TestCase, 0, len(input.Problem.TestCases))
	for _, testCase := range input.Problem.TestCases {
		if testCase.IsSample {
			samples = append(samples, testCase)
		}
	}
	if len(samples) == 0 {
		return CallOutput{
			ToolName: SampleJudgeToolName,
			Status:   CallStatusSkipped,
			Summary:  "no sample test cases available",
			Metadata: map[string]any{
				"verdict":     judge.VerdictUnjudgeable,
				"total_count": 0,
			},
		}, nil
	}

	result, err := t.engine.Judge(ctx, judge.Request{
		Problem:    input.Problem,
		TestCases:  samples,
		Language:   model.LanguageCPP17,
		SourceCode: input.SourceCode,
	})
	if err != nil {
		return CallOutput{}, fmt.Errorf("run sample judge: %w", err)
	}

	status := CallStatusOK
	if result.Verdict != judge.VerdictAccepted {
		status = CallStatusFailed
	}
	return CallOutput{
		ToolName: SampleJudgeToolName,
		Status:   status,
		Summary:  fmt.Sprintf("sample judge verdict %s (%d/%d)", result.Verdict, result.PassedCount, result.TotalCount),
		Metadata: map[string]any{
			"verdict":         result.Verdict,
			"passed_count":    result.PassedCount,
			"total_count":     result.TotalCount,
			"compile_stderr":  result.CompileStderr,
			"stdout":          result.RunStdout,
			"stderr":          result.RunStderr,
			"runtime_ms":      result.RuntimeMS,
			"memory_kb":       result.MemoryKB,
			"exit_code":       result.ExitCode,
			"timed_out":       result.TimedOut,
			"memory_exceeded": result.MemoryExceeded,
			"exec_stage":      result.ExecStage,
			"error_message":   result.ErrorMessage,
		},
	}, nil
}
