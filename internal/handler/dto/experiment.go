package dto

import (
	"time"

	"ai-for-oj/internal/service"
)

type RunExperimentRequest struct {
	Name       string `json:"name"`
	ProblemIDs []uint `json:"problem_ids" binding:"required,min=1"`
	Model      string `json:"model"`
	PromptName string `json:"prompt_name"`
	AgentName  string `json:"agent_name"`
}

type ExperimentRunResponse struct {
	ID           uint      `json:"id"`
	ProblemID    uint      `json:"problem_id"`
	AISolveRunID *uint     `json:"ai_solve_run_id,omitempty"`
	SubmissionID *uint     `json:"submission_id,omitempty"`
	AttemptNo    int       `json:"attempt_no"`
	Verdict      string    `json:"verdict,omitempty"`
	Status       string    `json:"status"`
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type ExperimentResponse struct {
	ID                  uint                          `json:"id"`
	Name                string                        `json:"name"`
	Model               string                        `json:"model"`
	PromptName          string                        `json:"prompt_name"`
	AgentName           string                        `json:"agent_name"`
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
