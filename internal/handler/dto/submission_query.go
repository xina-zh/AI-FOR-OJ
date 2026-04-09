package dto

import "time"

type SubmissionListResponse struct {
	Items      []SubmissionSummaryResponse `json:"items"`
	Page       int                         `json:"page"`
	PageSize   int                         `json:"page_size"`
	Total      int64                       `json:"total"`
	TotalPages int                         `json:"total_pages"`
}

type SubmissionProblemStatsResponse struct {
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

type SubmissionJudgeResultResponse struct {
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

type SubmissionTestCaseResultResponse struct {
	TestCaseID uint   `json:"testcase_id"`
	CaseIndex  int    `json:"index"`
	Verdict    string `json:"verdict"`
	RuntimeMS  int    `json:"runtime_ms"`
	Stdout     string `json:"stdout,omitempty"`
	Stderr     string `json:"stderr,omitempty"`
	ExitCode   int    `json:"exit_code"`
	TimedOut   bool   `json:"timed_out"`
}

type SubmissionSummaryResponse struct {
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

type SubmissionDetailResponse struct {
	ID              uint                               `json:"id"`
	ProblemID       uint                               `json:"problem_id"`
	ProblemTitle    string                             `json:"problem_title"`
	Language        string                             `json:"language"`
	SourceType      string                             `json:"source_type"`
	SourceCode      string                             `json:"source_code"`
	Verdict         string                             `json:"verdict"`
	RuntimeMS       int                                `json:"runtime_ms"`
	MemoryKB        int                                `json:"memory_kb"`
	PassedCount     int                                `json:"passed_count"`
	TotalCount      int                                `json:"total_count"`
	CompileStderr   string                             `json:"compile_stderr,omitempty"`
	RunStdout       string                             `json:"run_stdout,omitempty"`
	RunStderr       string                             `json:"run_stderr,omitempty"`
	ExitCode        int                                `json:"exit_code"`
	TimedOut        bool                               `json:"timed_out"`
	ExecStage       string                             `json:"exec_stage,omitempty"`
	ErrorMessage    string                             `json:"error_message,omitempty"`
	CreatedAt       time.Time                          `json:"created_at"`
	UpdatedAt       time.Time                          `json:"updated_at"`
	JudgeResult     *SubmissionJudgeResultResponse     `json:"judge_result,omitempty"`
	TestCaseResults []SubmissionTestCaseResultResponse `json:"testcase_results,omitempty"`
}
