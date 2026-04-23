package dto

import (
	"time"

	"ai-for-oj/internal/service"
)

type RunExperimentRequest struct {
	Name          string `json:"name"`
	ProblemIDs    []uint `json:"problem_ids" binding:"required,min=1"`
	Model         string `json:"model"`
	PromptName    string `json:"prompt_name"`
	AgentName     string `json:"agent_name"`
	ToolingConfig string `json:"tooling_config"`
}

type ExperimentRunResponse struct {
	ID            uint      `json:"id"`
	ProblemID     uint      `json:"problem_id"`
	AISolveRunID  *uint     `json:"ai_solve_run_id,omitempty"`
	SubmissionID  *uint     `json:"submission_id,omitempty"`
	AttemptNo     int       `json:"attempt_no"`
	Verdict       string    `json:"verdict,omitempty"`
	Status        string    `json:"status"`
	ErrorMessage  string    `json:"error_message,omitempty"`
	AttemptCount  int       `json:"attempt_count"`
	FailureType   string    `json:"failure_type,omitempty"`
	StrategyPath  string    `json:"strategy_path,omitempty"`
	ToolingConfig string    `json:"tooling_config"`
	ToolCallCount int       `json:"tool_call_count"`
	CreatedAt     time.Time `json:"created_at"`
}

type ExperimentResponse struct {
	ID                  uint                          `json:"id"`
	Name                string                        `json:"name"`
	Model               string                        `json:"model"`
	PromptName          string                        `json:"prompt_name"`
	AgentName           string                        `json:"agent_name"`
	ToolingConfig       string                        `json:"tooling_config"`
	Status              string                        `json:"status"`
	TotalCount          int                           `json:"total_count"`
	SuccessCount        int                           `json:"success_count"`
	ACCount             int                           `json:"ac_count"`
	FailedCount         int                           `json:"failed_count"`
	VerdictDistribution service.VerdictDistribution   `json:"verdict_distribution"`
	CostSummary         service.ExperimentCostSummary `json:"cost_summary"`
	CreatedAt           time.Time                     `json:"created_at"`
	UpdatedAt           time.Time                     `json:"updated_at"`
	Runs                []ExperimentRunResponse       `json:"runs"`
}

type ExperimentListResponse struct {
	Items      []ExperimentResponse `json:"items"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	Total      int64                `json:"total"`
	TotalPages int                  `json:"total_pages"`
}

type ExperimentRunTraceEventResponse struct {
	ID         uint      `json:"id"`
	SequenceNo int       `json:"sequence_no"`
	StepType   string    `json:"step_type"`
	Content    string    `json:"content"`
	Metadata   string    `json:"metadata"`
	CreatedAt  time.Time `json:"created_at"`
}

type ExperimentTraceAISolveRunResponse struct {
	ID             uint   `json:"id"`
	Status         string `json:"status"`
	Verdict        string `json:"verdict"`
	AttemptCount   int    `json:"attempt_count"`
	FailureType    string `json:"failure_type,omitempty"`
	StrategyPath   string `json:"strategy_path,omitempty"`
	ToolingConfig  string `json:"tooling_config"`
	ToolCallCount  int    `json:"tool_call_count"`
	TokenInput     int64  `json:"token_input"`
	TokenOutput    int64  `json:"token_output"`
	LLMLatencyMS   int    `json:"llm_latency_ms"`
	TotalLatencyMS int    `json:"total_latency_ms"`
}

type ExperimentTraceSubmissionResponse struct {
	ID          uint   `json:"id"`
	ProblemID   uint   `json:"problem_id"`
	Language    string `json:"language"`
	SourceType  string `json:"source_type"`
	Verdict     string `json:"verdict,omitempty"`
	RuntimeMS   int    `json:"runtime_ms"`
	MemoryKB    int    `json:"memory_kb"`
	PassedCount int    `json:"passed_count"`
	TotalCount  int    `json:"total_count"`
}

type ExperimentRunTraceResponse struct {
	ExperimentRunID uint                               `json:"experiment_run_id"`
	ExperimentID    uint                               `json:"experiment_id"`
	ProblemID       uint                               `json:"problem_id"`
	AISolveRunID    *uint                              `json:"ai_solve_run_id,omitempty"`
	SubmissionID    *uint                              `json:"submission_id,omitempty"`
	AttemptNo       int                                `json:"attempt_no"`
	Verdict         string                             `json:"verdict,omitempty"`
	Status          string                             `json:"status"`
	ErrorMessage    string                             `json:"error_message,omitempty"`
	Timeline        []ExperimentRunTraceEventResponse  `json:"timeline"`
	AISolveRun      *ExperimentTraceAISolveRunResponse `json:"ai_solve_run,omitempty"`
	Submission      *ExperimentTraceSubmissionResponse `json:"submission,omitempty"`
}
