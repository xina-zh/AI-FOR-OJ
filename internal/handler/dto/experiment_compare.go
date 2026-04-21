package dto

import (
	"time"

	"ai-for-oj/internal/service"
)

type CompareExperimentRequest struct {
	Name                string `json:"name"`
	ProblemIDs          []uint `json:"problem_ids" binding:"required,min=1"`
	BaselineModel       string `json:"baseline_model"`
	CandidateModel      string `json:"candidate_model"`
	BaselinePromptName  string `json:"baseline_prompt_name"`
	CandidatePromptName string `json:"candidate_prompt_name"`
	BaselineAgentName   string `json:"baseline_agent_name"`
	CandidateAgentName  string `json:"candidate_agent_name"`
}

type ExperimentCompareProblemSummaryResponse struct {
	ProblemID             uint   `json:"problem_id"`
	BaselineVerdict       string `json:"baseline_verdict,omitempty"`
	CandidateVerdict      string `json:"candidate_verdict,omitempty"`
	Changed               bool   `json:"changed"`
	ChangeType            string `json:"change_type"`
	BaselineStatus        string `json:"baseline_status,omitempty"`
	CandidateStatus       string `json:"candidate_status,omitempty"`
	BaselineSubmissionID  *uint  `json:"baseline_submission_id,omitempty"`
	CandidateSubmissionID *uint  `json:"candidate_submission_id,omitempty"`
}

type ExperimentCompareHighlightedProblemResponse struct {
	ProblemID             uint   `json:"problem_id"`
	BaselineVerdict       string `json:"baseline_verdict,omitempty"`
	CandidateVerdict      string `json:"candidate_verdict,omitempty"`
	Changed               bool   `json:"changed"`
	ChangeType            string `json:"change_type"`
	BaselineSubmissionID  *uint  `json:"baseline_submission_id,omitempty"`
	CandidateSubmissionID *uint  `json:"candidate_submission_id,omitempty"`
}

type ExperimentCompareResponse struct {
	ID                    uint                                          `json:"id"`
	Name                  string                                        `json:"name"`
	CompareDimension      string                                        `json:"compare_dimension"`
	BaselineValue         string                                        `json:"baseline_value"`
	CandidateValue        string                                        `json:"candidate_value"`
	BaselinePromptName    string                                        `json:"baseline_prompt_name"`
	CandidatePromptName   string                                        `json:"candidate_prompt_name"`
	BaselineAgentName     string                                        `json:"baseline_agent_name"`
	CandidateAgentName    string                                        `json:"candidate_agent_name"`
	ProblemIDs            []uint                                        `json:"problem_ids"`
	BaselineExperimentID  uint                                          `json:"baseline_experiment_id"`
	CandidateExperimentID uint                                          `json:"candidate_experiment_id"`
	BaselineSummary       *ExperimentResponse                           `json:"baseline_summary,omitempty"`
	CandidateSummary      *ExperimentResponse                           `json:"candidate_summary,omitempty"`
	BaselineDistribution  service.VerdictDistribution                   `json:"baseline_verdict_distribution"`
	CandidateDistribution service.VerdictDistribution                   `json:"candidate_verdict_distribution"`
	DeltaDistribution     service.VerdictDistribution                   `json:"delta_verdict_distribution"`
	CostComparison        service.ExperimentCompareCostComparison       `json:"cost_comparison"`
	ComparisonSummary     service.ExperimentCompareSummary              `json:"comparison_summary"`
	ImprovedCount         int                                           `json:"improved_count"`
	RegressedCount        int                                           `json:"regressed_count"`
	ChangedNonACCount     int                                           `json:"changed_non_ac_count"`
	ProblemSummaries      []ExperimentCompareProblemSummaryResponse     `json:"problem_summaries"`
	HighlightedProblems   []ExperimentCompareHighlightedProblemResponse `json:"highlighted_problems"`
	DeltaACCount          int                                           `json:"delta_ac_count"`
	DeltaFailedCount      int                                           `json:"delta_failed_count"`
	Status                string                                        `json:"status"`
	ErrorMessage          string                                        `json:"error_message,omitempty"`
	CreatedAt             time.Time                                     `json:"created_at"`
	UpdatedAt             time.Time                                     `json:"updated_at"`
}

type ExperimentCompareListResponse struct {
	Items      []ExperimentCompareResponse `json:"items"`
	Page       int                         `json:"page"`
	PageSize   int                         `json:"page_size"`
	Total      int64                       `json:"total"`
	TotalPages int                         `json:"total_pages"`
}
