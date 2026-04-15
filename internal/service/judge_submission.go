package service

import (
	"context"
	"errors"
	"fmt"

	"ai-for-oj/internal/judge"
	"ai-for-oj/internal/model"
	"ai-for-oj/internal/repository"
)

var ErrUnsupportedLanguage = errors.New("unsupported language")

type JudgeSubmissionInput struct {
	ProblemID  uint
	SourceCode string
	Language   string
	SourceType string
}

type JudgeSubmissionOutput struct {
	SubmissionID    uint                          `json:"submission_id"`
	ProblemID       uint                          `json:"problem_id"`
	Language        string                        `json:"language"`
	SourceType      string                        `json:"source_type"`
	Verdict         string                        `json:"verdict"`
	RuntimeMS       int                           `json:"runtime_ms"`
	MemoryKB        int                           `json:"memory_kb"`
	PassedCount     int                           `json:"passed_count"`
	TotalCount      int                           `json:"total_count"`
	ErrorMessage    string                        `json:"error_message,omitempty"`
	CompileStderr   string                        `json:"compile_stderr,omitempty"`
	RunStdout       string                        `json:"run_stdout,omitempty"`
	RunStderr       string                        `json:"run_stderr,omitempty"`
	ExitCode        int                           `json:"exit_code"`
	TimedOut        bool                          `json:"timed_out"`
	ExecStage       string                        `json:"exec_stage,omitempty"`
	TestCaseResults []JudgeSubmissionCaseFeedback `json:"test_case_results,omitempty"`
}

type JudgeSubmissionCaseFeedback struct {
	CaseIndex int    `json:"case_index"`
	Verdict   string `json:"verdict"`
	RuntimeMS int    `json:"runtime_ms"`
	Stdout    string `json:"stdout,omitempty"`
	Stderr    string `json:"stderr,omitempty"`
	ExitCode  int    `json:"exit_code"`
	TimedOut  bool   `json:"timed_out"`
}

type JudgeSubmissionService struct {
	problems    repository.ProblemRepository
	submissions repository.SubmissionRepository
	engine      judge.Engine
}

func NewJudgeSubmissionService(
	problems repository.ProblemRepository,
	submissions repository.SubmissionRepository,
	engine judge.Engine,
) *JudgeSubmissionService {
	return &JudgeSubmissionService{
		problems:    problems,
		submissions: submissions,
		engine:      engine,
	}
}

func (s *JudgeSubmissionService) Submit(ctx context.Context, input JudgeSubmissionInput) (*JudgeSubmissionOutput, error) {
	if input.Language != model.LanguageCPP17 {
		return nil, ErrUnsupportedLanguage
	}

	problem, err := s.problems.GetByIDWithTestCases(ctx, input.ProblemID)
	if err != nil {
		return nil, err
	}

	submission := &model.Submission{
		ProblemID:  input.ProblemID,
		SourceCode: input.SourceCode,
		Language:   input.Language,
		SourceType: defaultSourceType(input.SourceType),
	}
	if err := s.submissions.Create(ctx, submission); err != nil {
		return nil, fmt.Errorf("create submission: %w", err)
	}

	judgeResult, err := s.engine.Judge(ctx, judge.Request{
		Problem:    problem,
		TestCases:  problem.TestCases,
		Language:   input.Language,
		SourceCode: input.SourceCode,
	})
	if err != nil {
		return nil, fmt.Errorf("judge submission: %w", err)
	}

	persistedResult := &model.JudgeResult{
		SubmissionID:  submission.ID,
		Verdict:       judgeResult.Verdict,
		RuntimeMS:     judgeResult.RuntimeMS,
		MemoryKB:      judgeResult.MemoryKB,
		PassedCount:   judgeResult.PassedCount,
		TotalCount:    judgeResult.TotalCount,
		CompileStderr: judgeResult.CompileStderr,
		RunStdout:     judgeResult.RunStdout,
		RunStderr:     judgeResult.RunStderr,
		ExitCode:      judgeResult.ExitCode,
		TimedOut:      judgeResult.TimedOut,
		ExecStage:     judgeResult.ExecStage,
		ErrorMessage:  judgeResult.ErrorMessage,
	}
	if err := s.submissions.CreateJudgeResult(ctx, persistedResult); err != nil {
		return nil, fmt.Errorf("create judge result: %w", err)
	}

	testCaseResults := make([]model.SubmissionTestCaseResult, 0, len(judgeResult.TestCaseResults))
	outputCaseResults := make([]JudgeSubmissionCaseFeedback, 0, len(judgeResult.TestCaseResults))
	for _, item := range judgeResult.TestCaseResults {
		testCaseResults = append(testCaseResults, model.SubmissionTestCaseResult{
			SubmissionID: submission.ID,
			TestCaseID:   item.TestCaseID,
			CaseIndex:    item.CaseIndex,
			Verdict:      item.Verdict,
			RuntimeMS:    item.RuntimeMS,
			Stdout:       item.Stdout,
			Stderr:       item.Stderr,
			ExitCode:     item.ExitCode,
			TimedOut:     item.TimedOut,
		})
		outputCaseResults = append(outputCaseResults, JudgeSubmissionCaseFeedback{
			CaseIndex: item.CaseIndex,
			Verdict:   item.Verdict,
			RuntimeMS: item.RuntimeMS,
			Stdout:    item.Stdout,
			Stderr:    item.Stderr,
			ExitCode:  item.ExitCode,
			TimedOut:  item.TimedOut,
		})
	}
	if err := s.submissions.CreateTestCaseResults(ctx, testCaseResults); err != nil {
		return nil, fmt.Errorf("create submission test case results: %w", err)
	}

	return &JudgeSubmissionOutput{
		SubmissionID:    submission.ID,
		ProblemID:       input.ProblemID,
		Language:        input.Language,
		SourceType:      submission.SourceType,
		Verdict:         judgeResult.Verdict,
		RuntimeMS:       judgeResult.RuntimeMS,
		MemoryKB:        judgeResult.MemoryKB,
		PassedCount:     judgeResult.PassedCount,
		TotalCount:      judgeResult.TotalCount,
		ErrorMessage:    judgeResult.ErrorMessage,
		CompileStderr:   judgeResult.CompileStderr,
		RunStdout:       judgeResult.RunStdout,
		RunStderr:       judgeResult.RunStderr,
		ExitCode:        judgeResult.ExitCode,
		TimedOut:        judgeResult.TimedOut,
		ExecStage:       judgeResult.ExecStage,
		TestCaseResults: outputCaseResults,
	}, nil
}

func defaultSourceType(sourceType string) string {
	if sourceType == "" {
		return model.SourceTypeHuman
	}
	return sourceType
}
