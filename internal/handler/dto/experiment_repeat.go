package dto

import (
	"time"

	"ai-for-oj/internal/service"
)

type RepeatExperimentRequest struct {
	Name        string `json:"name"`
	ProblemIDs  []uint `json:"problem_ids" binding:"required,min=1"`
	Model       string `json:"model"`
	RepeatCount int    `json:"repeat_count" binding:"required,gte=1,lte=10"`
}

type ExperimentRepeatRoundSummaryResponse struct {
	RoundNo             int                         `json:"round_no"`
	ExperimentID        uint                        `json:"experiment_id"`
	ACCount             int                         `json:"ac_count"`
	FailedCount         int                         `json:"failed_count"`
	VerdictDistribution service.VerdictDistribution `json:"verdict_distribution"`
	Status              string                      `json:"status"`
}

type ExperimentRepeatProblemSummaryResponse struct {
	ProblemID           uint                        `json:"problem_id"`
	TotalRounds         int                         `json:"total_rounds"`
	ACCount             int                         `json:"ac_count"`
	FailedCount         int                         `json:"failed_count"`
	ACRate              float64                     `json:"ac_rate"`
	VerdictDistribution service.VerdictDistribution `json:"verdict_distribution"`
	LatestVerdict       string                      `json:"latest_verdict,omitempty"`
}

type ExperimentRepeatUnstableProblemResponse struct {
	ProblemID           uint                        `json:"problem_id"`
	TotalRounds         int                         `json:"total_rounds"`
	ACCount             int                         `json:"ac_count"`
	FailedCount         int                         `json:"failed_count"`
	ACRate              float64                     `json:"ac_rate"`
	VerdictDistribution service.VerdictDistribution `json:"verdict_distribution"`
	LatestVerdict       string                      `json:"latest_verdict,omitempty"`
	InstabilityScore    int                         `json:"instability_score"`
	VerdictKindCount    int                         `json:"verdict_kind_count"`
}

type ExperimentRepeatResponse struct {
	ID                         uint                                      `json:"id"`
	Name                       string                                    `json:"name"`
	Model                      string                                    `json:"model"`
	ProblemIDs                 []uint                                    `json:"problem_ids"`
	RepeatCount                int                                       `json:"repeat_count"`
	ExperimentIDs              []uint                                    `json:"experiment_ids"`
	TotalProblemCount          int                                       `json:"total_problem_count"`
	TotalRunCount              int                                       `json:"total_run_count"`
	OverallACCount             int                                       `json:"overall_ac_count"`
	OverallFailedCount         int                                       `json:"overall_failed_count"`
	AverageACCountPerRound     float64                                   `json:"average_ac_count_per_round"`
	AverageFailedCountPerRound float64                                   `json:"average_failed_count_per_round"`
	OverallACRate              float64                                   `json:"overall_ac_rate"`
	BestRoundACCount           int                                       `json:"best_round_ac_count"`
	WorstRoundACCount          int                                       `json:"worst_round_ac_count"`
	CostSummary                service.ExperimentRepeatCostSummary       `json:"cost_summary"`
	Status                     string                                    `json:"status"`
	ErrorMessage               string                                    `json:"error_message,omitempty"`
	RoundSummaries             []ExperimentRepeatRoundSummaryResponse    `json:"round_summaries"`
	ProblemSummaries           []ExperimentRepeatProblemSummaryResponse  `json:"problem_summaries"`
	MostUnstableProblems       []ExperimentRepeatUnstableProblemResponse `json:"most_unstable_problems"`
	CreatedAt                  time.Time                                 `json:"created_at"`
	UpdatedAt                  time.Time                                 `json:"updated_at"`
}
