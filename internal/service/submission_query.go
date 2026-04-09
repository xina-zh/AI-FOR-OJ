package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"ai-for-oj/internal/model"
	"ai-for-oj/internal/repository"
)

type SubmissionJudgeResultOutput struct {
	ID            uint      `json:"id"`
	Verdict       string    `json:"verdict"`
	RuntimeMS     int       `json:"runtime_ms"`
	MemoryKB      int       `json:"memory_kb"`
	PassedCount   int       `json:"passed_count"`
	TotalCount    int       `json:"total_count"`
	CompileStderr string    `json:"compile_stderr,omitempty"`
	RunStdout     string    `json:"run_stdout,omitempty"`
	RunStderr     string    `json:"run_stderr,omitempty"`
	ExitCode      int       `json:"exit_code"`
	TimedOut      bool      `json:"timed_out"`
	ExecStage     string    `json:"exec_stage,omitempty"`
	ErrorMessage  string    `json:"error_message,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type SubmissionTestCaseResultOutput struct {
	TestCaseID uint   `json:"testcase_id"`
	CaseIndex  int    `json:"index"`
	Verdict    string `json:"verdict"`
	RuntimeMS  int    `json:"runtime_ms"`
	Stdout     string `json:"stdout,omitempty"`
	Stderr     string `json:"stderr,omitempty"`
	ExitCode   int    `json:"exit_code"`
	TimedOut   bool   `json:"timed_out"`
}

type SubmissionSummaryOutput struct {
	ID           uint      `json:"id"`
	ProblemID    uint      `json:"problem_id"`
	ProblemTitle string    `json:"problem_title"`
	Language     string    `json:"language"`
	SourceType   string    `json:"source_type"`
	Verdict      string    `json:"verdict"`
	RuntimeMS    int       `json:"runtime_ms"`
	PassedCount  int       `json:"passed_count"`
	TotalCount   int       `json:"total_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type SubmissionDetailOutput struct {
	ID              uint                             `json:"id"`
	ProblemID       uint                             `json:"problem_id"`
	ProblemTitle    string                           `json:"problem_title"`
	Language        string                           `json:"language"`
	SourceType      string                           `json:"source_type"`
	SourceCode      string                           `json:"source_code"`
	Verdict         string                           `json:"verdict"`
	RuntimeMS       int                              `json:"runtime_ms"`
	MemoryKB        int                              `json:"memory_kb"`
	PassedCount     int                              `json:"passed_count"`
	TotalCount      int                              `json:"total_count"`
	CompileStderr   string                           `json:"compile_stderr,omitempty"`
	RunStdout       string                           `json:"run_stdout,omitempty"`
	RunStderr       string                           `json:"run_stderr,omitempty"`
	ExitCode        int                              `json:"exit_code"`
	TimedOut        bool                             `json:"timed_out"`
	ExecStage       string                           `json:"exec_stage,omitempty"`
	ErrorMessage    string                           `json:"error_message,omitempty"`
	CreatedAt       time.Time                        `json:"created_at"`
	UpdatedAt       time.Time                        `json:"updated_at"`
	JudgeResult     *SubmissionJudgeResultOutput     `json:"judge_result,omitempty"`
	TestCaseResults []SubmissionTestCaseResultOutput `json:"testcase_results,omitempty"`
}

type SubmissionQueryService struct {
	submissions repository.SubmissionRepository
}

type SubmissionListInput struct {
	Page      int
	PageSize  int
	ProblemID *uint
}

type SubmissionListOutput struct {
	Items      []SubmissionSummaryOutput `json:"items"`
	Page       int                       `json:"page"`
	PageSize   int                       `json:"page_size"`
	Total      int64                     `json:"total"`
	TotalPages int                       `json:"total_pages"`
}

type SubmissionProblemStatsOutput struct {
	ProblemID          uint       `json:"problem_id"`
	ProblemTitle       string     `json:"problem_title"`
	TotalSubmissions   int64      `json:"total_submissions"`
	ACCount            int64      `json:"ac_count"`
	WACount            int64      `json:"wa_count"`
	CECount            int64      `json:"ce_count"`
	RECount            int64      `json:"re_count"`
	TLECount           int64      `json:"tle_count"`
	LatestSubmissionAt *time.Time `json:"latest_submission_at,omitempty"`
}

func NewSubmissionQueryService(submissions repository.SubmissionRepository) *SubmissionQueryService {
	return &SubmissionQueryService{submissions: submissions}
}

func (s *SubmissionQueryService) List(ctx context.Context, input SubmissionListInput) (*SubmissionListOutput, error) {
	submissions, total, err := s.submissions.List(ctx, repository.SubmissionListQuery{
		Page:      input.Page,
		PageSize:  input.PageSize,
		ProblemID: input.ProblemID,
	})
	if err != nil {
		return nil, fmt.Errorf("list submissions: %w", err)
	}

	outputs := make([]SubmissionSummaryOutput, 0, len(submissions))
	for _, submission := range submissions {
		outputs = append(outputs, toSubmissionSummaryOutput(submission))
	}

	totalPages := 0
	if input.PageSize > 0 && total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(input.PageSize)))
	}

	return &SubmissionListOutput{
		Items:      outputs,
		Page:       input.Page,
		PageSize:   input.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *SubmissionQueryService) Get(ctx context.Context, submissionID uint) (*SubmissionDetailOutput, error) {
	submission, err := s.submissions.GetByID(ctx, submissionID)
	if err != nil {
		return nil, err
	}
	output := toSubmissionDetailOutput(*submission)
	return &output, nil
}

func (s *SubmissionQueryService) AggregateByProblem(ctx context.Context) ([]SubmissionProblemStatsOutput, error) {
	rows, err := s.submissions.AggregateByProblem(ctx)
	if err != nil {
		return nil, fmt.Errorf("aggregate submissions by problem: %w", err)
	}

	outputs := make([]SubmissionProblemStatsOutput, 0, len(rows))
	for _, row := range rows {
		outputs = append(outputs, SubmissionProblemStatsOutput{
			ProblemID:          row.ProblemID,
			ProblemTitle:       row.ProblemTitle,
			TotalSubmissions:   row.TotalSubmissions,
			ACCount:            row.ACCount,
			WACount:            row.WACount,
			CECount:            row.CECount,
			RECount:            row.RECount,
			TLECount:           row.TLECount,
			LatestSubmissionAt: row.LatestSubmissionAt,
		})
	}
	return outputs, nil
}

func toSubmissionSummaryOutput(submission model.Submission) SubmissionSummaryOutput {
	output := SubmissionSummaryOutput{
		ID:           submission.ID,
		ProblemID:    submission.ProblemID,
		ProblemTitle: submission.Problem.Title,
		Language:     submission.Language,
		SourceType:   submission.SourceType,
		CreatedAt:    submission.CreatedAt,
		UpdatedAt:    submission.UpdatedAt,
	}
	if submission.JudgeResult != nil {
		output.Verdict = submission.JudgeResult.Verdict
		output.RuntimeMS = submission.JudgeResult.RuntimeMS
		output.PassedCount = submission.JudgeResult.PassedCount
		output.TotalCount = submission.JudgeResult.TotalCount
	}
	return output
}

func toSubmissionDetailOutput(submission model.Submission) SubmissionDetailOutput {
	output := SubmissionDetailOutput{
		ID:           submission.ID,
		ProblemID:    submission.ProblemID,
		ProblemTitle: submission.Problem.Title,
		Language:     submission.Language,
		SourceType:   submission.SourceType,
		SourceCode:   submission.SourceCode,
		CreatedAt:    submission.CreatedAt,
		UpdatedAt:    submission.UpdatedAt,
	}
	if submission.JudgeResult != nil {
		output.Verdict = submission.JudgeResult.Verdict
		output.RuntimeMS = submission.JudgeResult.RuntimeMS
		output.MemoryKB = submission.JudgeResult.MemoryKB
		output.PassedCount = submission.JudgeResult.PassedCount
		output.TotalCount = submission.JudgeResult.TotalCount
		output.CompileStderr = submission.JudgeResult.CompileStderr
		output.RunStdout = submission.JudgeResult.RunStdout
		output.RunStderr = submission.JudgeResult.RunStderr
		output.ExitCode = submission.JudgeResult.ExitCode
		output.TimedOut = submission.JudgeResult.TimedOut
		output.ExecStage = submission.JudgeResult.ExecStage
		output.ErrorMessage = submission.JudgeResult.ErrorMessage
		output.JudgeResult = &SubmissionJudgeResultOutput{
			ID:            submission.JudgeResult.ID,
			Verdict:       submission.JudgeResult.Verdict,
			RuntimeMS:     submission.JudgeResult.RuntimeMS,
			MemoryKB:      submission.JudgeResult.MemoryKB,
			PassedCount:   submission.JudgeResult.PassedCount,
			TotalCount:    submission.JudgeResult.TotalCount,
			CompileStderr: submission.JudgeResult.CompileStderr,
			RunStdout:     submission.JudgeResult.RunStdout,
			RunStderr:     submission.JudgeResult.RunStderr,
			ExitCode:      submission.JudgeResult.ExitCode,
			TimedOut:      submission.JudgeResult.TimedOut,
			ExecStage:     submission.JudgeResult.ExecStage,
			ErrorMessage:  submission.JudgeResult.ErrorMessage,
			CreatedAt:     submission.JudgeResult.CreatedAt,
			UpdatedAt:     submission.JudgeResult.UpdatedAt,
		}
	}
	if len(submission.TestCaseResults) > 0 {
		output.TestCaseResults = make([]SubmissionTestCaseResultOutput, 0, len(submission.TestCaseResults))
		for _, item := range submission.TestCaseResults {
			output.TestCaseResults = append(output.TestCaseResults, SubmissionTestCaseResultOutput{
				TestCaseID: item.TestCaseID,
				CaseIndex:  item.CaseIndex,
				Verdict:    item.Verdict,
				RuntimeMS:  item.RuntimeMS,
				Stdout:     item.Stdout,
				Stderr:     item.Stderr,
				ExitCode:   item.ExitCode,
				TimedOut:   item.TimedOut,
			})
		}
	}
	return output
}
