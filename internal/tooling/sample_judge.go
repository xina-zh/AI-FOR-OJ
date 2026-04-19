package tooling

import (
	"context"
	"encoding/json"
	"fmt"

	"ai-for-oj/internal/judge"
	"ai-for-oj/internal/model"
	"ai-for-oj/internal/repository"
)

const SampleJudgeToolName = "sample_judge"

type SampleJudgeTool struct {
	problems repository.ProblemRepository
	engine   judge.Engine
}

type sampleJudgeMetadata struct {
	Verdict       string `json:"verdict,omitempty"`
	PassedCount   int    `json:"passed_count"`
	TotalCount    int    `json:"total_count"`
	CompileStderr string `json:"compile_stderr,omitempty"`
	RunStdout     string `json:"run_stdout,omitempty"`
	RunStderr     string `json:"run_stderr,omitempty"`
	ErrorMessage  string `json:"error_message,omitempty"`
}

func NewSampleJudgeTool(problems repository.ProblemRepository, engine judge.Engine) *SampleJudgeTool {
	return &SampleJudgeTool{problems: problems, engine: engine}
}

func (t *SampleJudgeTool) Name() string {
	return SampleJudgeToolName
}

func (t *SampleJudgeTool) Execute(ctx context.Context, input CallInput) (CallResult, error) {
	problem, err := t.problemWithTestCases(ctx, input)
	if err != nil {
		return CallResult{}, err
	}

	samples := sampleCases(problem.TestCases)
	if len(samples) == 0 {
		return CallResult{
			ToolName: SampleJudgeToolName,
			Status:   CallStatusSkipped,
			Summary:  "sample_judge skipped: problem has no sample test cases",
		}, nil
	}

	result, err := t.engine.Judge(ctx, judge.Request{
		Problem:    problem,
		TestCases:  samples,
		Language:   model.LanguageCPP17,
		SourceCode: input.SourceCode,
	})
	if err != nil {
		return CallResult{}, err
	}

	metadata := sampleJudgeMetadata{
		Verdict:       result.Verdict,
		PassedCount:   result.PassedCount,
		TotalCount:    result.TotalCount,
		CompileStderr: result.CompileStderr,
		RunStdout:     result.RunStdout,
		RunStderr:     result.RunStderr,
		ErrorMessage:  result.ErrorMessage,
	}
	data, err := json.Marshal(metadata)
	if err != nil {
		return CallResult{}, fmt.Errorf("marshal sample judge metadata: %w", err)
	}

	return CallResult{
		ToolName: SampleJudgeToolName,
		Status:   CallStatusOK,
		Summary:  fmt.Sprintf("sample_judge verdict=%s passed=%d/%d", result.Verdict, result.PassedCount, result.TotalCount),
		Metadata: string(data),
	}, nil
}

func (t *SampleJudgeTool) problemWithTestCases(ctx context.Context, input CallInput) (*model.Problem, error) {
	if input.Problem != nil && len(input.Problem.TestCases) > 0 {
		return input.Problem, nil
	}
	return t.problems.GetByIDWithTestCases(ctx, input.ProblemID)
}

func sampleCases(testCases []model.TestCase) []model.TestCase {
	samples := make([]model.TestCase, 0, len(testCases))
	for _, item := range testCases {
		if item.IsSample {
			samples = append(samples, item)
		}
	}
	return samples
}
