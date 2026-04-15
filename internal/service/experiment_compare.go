package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"ai-for-oj/internal/agent"
	"ai-for-oj/internal/model"
	"ai-for-oj/internal/prompt"
	"ai-for-oj/internal/repository"
)

const (
	ExperimentCompareDimensionModel  = "model"
	ExperimentCompareDimensionPrompt = "prompt"
	ExperimentCompareDimensionAgent  = "agent"
)

type ExperimentRunner interface {
	Run(ctx context.Context, input RunExperimentInput) (*ExperimentOutput, error)
	Get(ctx context.Context, experimentID uint) (*ExperimentOutput, error)
}

type CompareExperimentInput struct {
	Name                string
	ProblemIDs          []uint
	BaselineModel       string
	CandidateModel      string
	BaselinePromptName  string
	CandidatePromptName string
	BaselineAgentName   string
	CandidateAgentName  string
}

type ExperimentCompareProblemSummary struct {
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

type ExperimentCompareHighlightedProblem struct {
	ProblemID             uint   `json:"problem_id"`
	BaselineVerdict       string `json:"baseline_verdict,omitempty"`
	CandidateVerdict      string `json:"candidate_verdict,omitempty"`
	Changed               bool   `json:"changed"`
	ChangeType            string `json:"change_type"`
	BaselineSubmissionID  *uint  `json:"baseline_submission_id,omitempty"`
	CandidateSubmissionID *uint  `json:"candidate_submission_id,omitempty"`
}

type ExperimentCompareCostComparison struct {
	BaselineTotalTokenInput        int64   `json:"baseline_total_token_input"`
	BaselineTotalTokenOutput       int64   `json:"baseline_total_token_output"`
	BaselineTotalTokens            int64   `json:"baseline_total_tokens"`
	BaselineAverageTokenInput      float64 `json:"baseline_average_token_input"`
	BaselineAverageTokenOutput     float64 `json:"baseline_average_token_output"`
	BaselineAverageTotalTokens     float64 `json:"baseline_average_total_tokens"`
	BaselineTotalLLMLatencyMS      int     `json:"baseline_total_llm_latency_ms"`
	BaselineTotalLatencyMS         int     `json:"baseline_total_latency_ms"`
	BaselineAverageLLMLatencyMS    float64 `json:"baseline_average_llm_latency_ms"`
	BaselineAverageTotalLatencyMS  float64 `json:"baseline_average_total_latency_ms"`
	BaselineRunCount               int     `json:"baseline_run_count"`
	CandidateTotalTokenInput       int64   `json:"candidate_total_token_input"`
	CandidateTotalTokenOutput      int64   `json:"candidate_total_token_output"`
	CandidateTotalTokens           int64   `json:"candidate_total_tokens"`
	CandidateAverageTokenInput     float64 `json:"candidate_average_token_input"`
	CandidateAverageTokenOutput    float64 `json:"candidate_average_token_output"`
	CandidateAverageTotalTokens    float64 `json:"candidate_average_total_tokens"`
	CandidateTotalLLMLatencyMS     int     `json:"candidate_total_llm_latency_ms"`
	CandidateTotalLatencyMS        int     `json:"candidate_total_latency_ms"`
	CandidateAverageLLMLatencyMS   float64 `json:"candidate_average_llm_latency_ms"`
	CandidateAverageTotalLatencyMS float64 `json:"candidate_average_total_latency_ms"`
	CandidateRunCount              int     `json:"candidate_run_count"`
	DeltaTotalTokens               int64   `json:"delta_total_tokens"`
	DeltaAverageTotalTokens        float64 `json:"delta_average_total_tokens"`
	DeltaTotalLatencyMS            int     `json:"delta_total_latency_ms"`
	DeltaAverageTotalLatencyMS     float64 `json:"delta_average_total_latency_ms"`
}

type ExperimentCompareSummary struct {
	CandidateBetterAC      bool   `json:"candidate_better_ac"`
	CandidateWorseAC       bool   `json:"candidate_worse_ac"`
	CandidateSameAC        bool   `json:"candidate_same_ac"`
	CandidateMoreExpensive bool   `json:"candidate_more_expensive"`
	CandidateCheaper       bool   `json:"candidate_cheaper"`
	CandidateSameCost      bool   `json:"candidate_same_cost"`
	CandidateSlower        bool   `json:"candidate_slower"`
	CandidateFaster        bool   `json:"candidate_faster"`
	CandidateSameLatency   bool   `json:"candidate_same_latency"`
	TradeoffType           string `json:"tradeoff_type"`
}

type ExperimentCompareOutput struct {
	ID                    uint                                  `json:"id"`
	Name                  string                                `json:"name"`
	CompareDimension      string                                `json:"compare_dimension"`
	BaselineValue         string                                `json:"baseline_value"`
	CandidateValue        string                                `json:"candidate_value"`
	BaselinePromptName    string                                `json:"baseline_prompt_name"`
	CandidatePromptName   string                                `json:"candidate_prompt_name"`
	BaselineAgentName     string                                `json:"baseline_agent_name"`
	CandidateAgentName    string                                `json:"candidate_agent_name"`
	ProblemIDs            []uint                                `json:"problem_ids"`
	BaselineExperimentID  uint                                  `json:"baseline_experiment_id"`
	CandidateExperimentID uint                                  `json:"candidate_experiment_id"`
	BaselineSummary       *ExperimentOutput                     `json:"baseline_summary,omitempty"`
	CandidateSummary      *ExperimentOutput                     `json:"candidate_summary,omitempty"`
	BaselineDistribution  VerdictDistribution                   `json:"baseline_verdict_distribution"`
	CandidateDistribution VerdictDistribution                   `json:"candidate_verdict_distribution"`
	DeltaDistribution     VerdictDistribution                   `json:"delta_verdict_distribution"`
	CostComparison        ExperimentCompareCostComparison       `json:"cost_comparison"`
	ComparisonSummary     ExperimentCompareSummary              `json:"comparison_summary"`
	ImprovedCount         int                                   `json:"improved_count"`
	RegressedCount        int                                   `json:"regressed_count"`
	ChangedNonACCount     int                                   `json:"changed_non_ac_count"`
	ProblemSummaries      []ExperimentCompareProblemSummary     `json:"problem_summaries"`
	HighlightedProblems   []ExperimentCompareHighlightedProblem `json:"highlighted_problems"`
	DeltaACCount          int                                   `json:"delta_ac_count"`
	DeltaFailedCount      int                                   `json:"delta_failed_count"`
	Status                string                                `json:"status"`
	ErrorMessage          string                                `json:"error_message,omitempty"`
	CreatedAt             time.Time                             `json:"created_at"`
	UpdatedAt             time.Time                             `json:"updated_at"`
}

type ExperimentCompareService struct {
	compares     repository.ExperimentCompareRepository
	experiments  ExperimentRunner
	defaultModel string
}

func NewExperimentCompareService(
	compares repository.ExperimentCompareRepository,
	experiments ExperimentRunner,
	defaultModel string,
) *ExperimentCompareService {
	return &ExperimentCompareService{
		compares:     compares,
		experiments:  experiments,
		defaultModel: defaultModel,
	}
}

func (s *ExperimentCompareService) Compare(ctx context.Context, input CompareExperimentInput) (*ExperimentCompareOutput, error) {
	problemIDsJSON, err := json.Marshal(input.ProblemIDs)
	if err != nil {
		return nil, fmt.Errorf("marshal problem ids: %w", err)
	}

	baselineModel := firstNonEmpty(input.BaselineModel, s.defaultModel)
	candidateModel := firstNonEmpty(input.CandidateModel, s.defaultModel)
	baselinePromptName, err := prompt.ResolveSolvePromptName(input.BaselinePromptName)
	if err != nil {
		return nil, err
	}
	candidatePromptName, err := prompt.ResolveSolvePromptName(input.CandidatePromptName)
	if err != nil {
		return nil, err
	}
	baselineAgentName, err := agent.ResolveSolveAgentName(input.BaselineAgentName)
	if err != nil {
		return nil, err
	}
	candidateAgentName, err := agent.ResolveSolveAgentName(input.CandidateAgentName)
	if err != nil {
		return nil, err
	}
	compareDimension, baselineValue, candidateValue := resolveCompareDimensionAndValues(
		baselineModel,
		candidateModel,
		baselinePromptName,
		candidatePromptName,
		baselineAgentName,
		candidateAgentName,
	)

	compare := &model.ExperimentCompare{
		Name:                defaultCompareName(input.Name),
		CompareDimension:    compareDimension,
		BaselineValue:       baselineValue,
		CandidateValue:      candidateValue,
		BaselinePromptName:  baselinePromptName,
		CandidatePromptName: candidatePromptName,
		BaselineAgentName:   baselineAgentName,
		CandidateAgentName:  candidateAgentName,
		ProblemIDs:          string(problemIDsJSON),
		Status:              model.ExperimentCompareStatusRunning,
	}
	if err := s.compares.Create(ctx, compare); err != nil {
		return nil, fmt.Errorf("create experiment compare: %w", err)
	}

	baseline, err := s.experiments.Run(ctx, RunExperimentInput{
		Name:       compare.Name + "-baseline",
		ProblemIDs: input.ProblemIDs,
		Model:      baselineModel,
		PromptName: baselinePromptName,
		AgentName:  baselineAgentName,
	})
	if err != nil {
		return s.failCompare(ctx, compare, err)
	}
	compare.BaselineExperimentID = &baseline.ID

	candidate, err := s.experiments.Run(ctx, RunExperimentInput{
		Name:       compare.Name + "-candidate",
		ProblemIDs: input.ProblemIDs,
		Model:      candidateModel,
		PromptName: candidatePromptName,
		AgentName:  candidateAgentName,
	})
	if err != nil {
		return s.failCompare(ctx, compare, err)
	}
	compare.CandidateExperimentID = &candidate.ID

	compare.DeltaACCount = candidate.ACCount - baseline.ACCount
	compare.DeltaFailedCount = candidate.FailedCount - baseline.FailedCount
	compare.Status = model.ExperimentCompareStatusCompleted
	if err := s.compares.Update(ctx, compare); err != nil {
		return nil, fmt.Errorf("update experiment compare: %w", err)
	}

	problemSummaries := buildCompareProblemSummaries(input.ProblemIDs, baseline, candidate)
	improvedCount, regressedCount, changedNonACCount := summarizeCompareProblemChanges(problemSummaries)
	highlightedProblems := buildHighlightedCompareProblems(problemSummaries)
	costComparison := buildExperimentCompareCostComparison(baseline, candidate)

	return &ExperimentCompareOutput{
		ID:                    compare.ID,
		Name:                  compare.Name,
		CompareDimension:      compare.CompareDimension,
		BaselineValue:         compare.BaselineValue,
		CandidateValue:        compare.CandidateValue,
		BaselinePromptName:    prompt.DisplaySolvePromptName(compare.BaselinePromptName),
		CandidatePromptName:   prompt.DisplaySolvePromptName(compare.CandidatePromptName),
		BaselineAgentName:     agent.DisplaySolveAgentName(compare.BaselineAgentName),
		CandidateAgentName:    agent.DisplaySolveAgentName(compare.CandidateAgentName),
		ProblemIDs:            append([]uint(nil), input.ProblemIDs...),
		BaselineExperimentID:  baseline.ID,
		CandidateExperimentID: candidate.ID,
		BaselineSummary:       baseline,
		CandidateSummary:      candidate,
		BaselineDistribution:  baseline.VerdictDistribution,
		CandidateDistribution: candidate.VerdictDistribution,
		DeltaDistribution:     DiffVerdictDistribution(candidate.VerdictDistribution, baseline.VerdictDistribution),
		CostComparison:        costComparison,
		ComparisonSummary:     buildExperimentCompareSummary(baseline, candidate, costComparison),
		ImprovedCount:         improvedCount,
		RegressedCount:        regressedCount,
		ChangedNonACCount:     changedNonACCount,
		ProblemSummaries:      problemSummaries,
		HighlightedProblems:   highlightedProblems,
		DeltaACCount:          compare.DeltaACCount,
		DeltaFailedCount:      compare.DeltaFailedCount,
		Status:                compare.Status,
		ErrorMessage:          compare.ErrorMessage,
		CreatedAt:             compare.CreatedAt,
		UpdatedAt:             compare.UpdatedAt,
	}, nil
}

func (s *ExperimentCompareService) Get(ctx context.Context, compareID uint) (*ExperimentCompareOutput, error) {
	compare, err := s.compares.GetByID(ctx, compareID)
	if err != nil {
		return nil, err
	}

	problemIDs, err := decodeProblemIDs(compare.ProblemIDs)
	if err != nil {
		return nil, fmt.Errorf("decode compare problem ids: %w", err)
	}

	var baseline *ExperimentOutput
	if compare.BaselineExperimentID != nil {
		baseline, err = s.experiments.Get(ctx, *compare.BaselineExperimentID)
		if err != nil {
			return nil, err
		}
	}

	var candidate *ExperimentOutput
	if compare.CandidateExperimentID != nil {
		candidate, err = s.experiments.Get(ctx, *compare.CandidateExperimentID)
		if err != nil {
			return nil, err
		}
	}

	problemSummaries := buildCompareProblemSummaries(problemIDs, baseline, candidate)
	improvedCount, regressedCount, changedNonACCount := summarizeCompareProblemChanges(problemSummaries)
	highlightedProblems := buildHighlightedCompareProblems(problemSummaries)
	costComparison := buildExperimentCompareCostComparison(baseline, candidate)

	return &ExperimentCompareOutput{
		ID:                    compare.ID,
		Name:                  compare.Name,
		CompareDimension:      compare.CompareDimension,
		BaselineValue:         compare.BaselineValue,
		CandidateValue:        compare.CandidateValue,
		BaselinePromptName:    prompt.DisplaySolvePromptName(compare.BaselinePromptName),
		CandidatePromptName:   prompt.DisplaySolvePromptName(compare.CandidatePromptName),
		BaselineAgentName:     agent.DisplaySolveAgentName(compare.BaselineAgentName),
		CandidateAgentName:    agent.DisplaySolveAgentName(compare.CandidateAgentName),
		ProblemIDs:            problemIDs,
		BaselineExperimentID:  derefUint(compare.BaselineExperimentID),
		CandidateExperimentID: derefUint(compare.CandidateExperimentID),
		BaselineSummary:       baseline,
		CandidateSummary:      candidate,
		BaselineDistribution:  verdictDistributionOf(baseline),
		CandidateDistribution: verdictDistributionOf(candidate),
		DeltaDistribution:     DiffVerdictDistribution(verdictDistributionOf(candidate), verdictDistributionOf(baseline)),
		CostComparison:        costComparison,
		ComparisonSummary:     buildExperimentCompareSummary(baseline, candidate, costComparison),
		ImprovedCount:         improvedCount,
		RegressedCount:        regressedCount,
		ChangedNonACCount:     changedNonACCount,
		ProblemSummaries:      problemSummaries,
		HighlightedProblems:   highlightedProblems,
		DeltaACCount:          compare.DeltaACCount,
		DeltaFailedCount:      compare.DeltaFailedCount,
		Status:                compare.Status,
		ErrorMessage:          compare.ErrorMessage,
		CreatedAt:             compare.CreatedAt,
		UpdatedAt:             compare.UpdatedAt,
	}, nil
}

func verdictDistributionOf(output *ExperimentOutput) VerdictDistribution {
	if output == nil {
		return VerdictDistribution{}
	}
	return output.VerdictDistribution
}

func costSummaryOf(output *ExperimentOutput) ExperimentCostSummary {
	if output == nil {
		return ExperimentCostSummary{}
	}
	return output.CostSummary
}

func resolveCompareDimensionAndValues(
	baselineModel, candidateModel, baselinePromptName, candidatePromptName, baselineAgentName, candidateAgentName string,
) (dimension, baselineValue, candidateValue string) {
	if baselineModel == candidateModel && baselinePromptName != candidatePromptName {
		return ExperimentCompareDimensionPrompt, baselinePromptName, candidatePromptName
	}
	if baselineModel == candidateModel && baselinePromptName == candidatePromptName && baselineAgentName != candidateAgentName {
		return ExperimentCompareDimensionAgent, baselineAgentName, candidateAgentName
	}
	return ExperimentCompareDimensionModel, baselineModel, candidateModel
}

func buildExperimentCompareCostComparison(baseline, candidate *ExperimentOutput) ExperimentCompareCostComparison {
	baselineCost := costSummaryOf(baseline)
	candidateCost := costSummaryOf(candidate)

	return ExperimentCompareCostComparison{
		BaselineTotalTokenInput:        baselineCost.TotalTokenInput,
		BaselineTotalTokenOutput:       baselineCost.TotalTokenOutput,
		BaselineTotalTokens:            baselineCost.TotalTokens,
		BaselineAverageTokenInput:      baselineCost.AverageTokenInput,
		BaselineAverageTokenOutput:     baselineCost.AverageTokenOutput,
		BaselineAverageTotalTokens:     baselineCost.AverageTotalTokens,
		BaselineTotalLLMLatencyMS:      baselineCost.TotalLLMLatencyMS,
		BaselineTotalLatencyMS:         baselineCost.TotalLatencyMS,
		BaselineAverageLLMLatencyMS:    baselineCost.AverageLLMLatencyMS,
		BaselineAverageTotalLatencyMS:  baselineCost.AverageTotalLatencyMS,
		BaselineRunCount:               baselineCost.RunCount,
		CandidateTotalTokenInput:       candidateCost.TotalTokenInput,
		CandidateTotalTokenOutput:      candidateCost.TotalTokenOutput,
		CandidateTotalTokens:           candidateCost.TotalTokens,
		CandidateAverageTokenInput:     candidateCost.AverageTokenInput,
		CandidateAverageTokenOutput:    candidateCost.AverageTokenOutput,
		CandidateAverageTotalTokens:    candidateCost.AverageTotalTokens,
		CandidateTotalLLMLatencyMS:     candidateCost.TotalLLMLatencyMS,
		CandidateTotalLatencyMS:        candidateCost.TotalLatencyMS,
		CandidateAverageLLMLatencyMS:   candidateCost.AverageLLMLatencyMS,
		CandidateAverageTotalLatencyMS: candidateCost.AverageTotalLatencyMS,
		CandidateRunCount:              candidateCost.RunCount,
		DeltaTotalTokens:               candidateCost.TotalTokens - baselineCost.TotalTokens,
		DeltaAverageTotalTokens:        candidateCost.AverageTotalTokens - baselineCost.AverageTotalTokens,
		DeltaTotalLatencyMS:            candidateCost.TotalLatencyMS - baselineCost.TotalLatencyMS,
		DeltaAverageTotalLatencyMS:     candidateCost.AverageTotalLatencyMS - baselineCost.AverageTotalLatencyMS,
	}
}

func buildExperimentCompareSummary(
	baseline, candidate *ExperimentOutput,
	costComparison ExperimentCompareCostComparison,
) ExperimentCompareSummary {
	baselineAC := acCountOf(baseline)
	candidateAC := acCountOf(candidate)
	acCmp := compareInt(candidateAC, baselineAC)
	costCmp := compareInt64(costComparison.CandidateTotalTokens, costComparison.BaselineTotalTokens)
	latencyCmp := compareFloat64(costComparison.CandidateAverageTotalLatencyMS, costComparison.BaselineAverageTotalLatencyMS)

	return ExperimentCompareSummary{
		CandidateBetterAC:      acCmp > 0,
		CandidateWorseAC:       acCmp < 0,
		CandidateSameAC:        acCmp == 0,
		CandidateMoreExpensive: costCmp > 0,
		CandidateCheaper:       costCmp < 0,
		CandidateSameCost:      costCmp == 0,
		CandidateSlower:        latencyCmp > 0,
		CandidateFaster:        latencyCmp < 0,
		CandidateSameLatency:   latencyCmp == 0,
		TradeoffType:           compareTradeoffType(acCmp, costCmp),
	}
}

func acCountOf(output *ExperimentOutput) int {
	if output == nil {
		return 0
	}
	return output.ACCount
}

func compareTradeoffType(acCmp, costCmp int) string {
	switch {
	case acCmp > 0 && costCmp > 0:
		return "improved_with_higher_cost"
	case acCmp > 0 && costCmp < 0:
		return "improved_with_lower_cost"
	case acCmp == 0 && costCmp > 0:
		return "same_outcome_higher_cost"
	case acCmp == 0 && costCmp < 0:
		return "same_outcome_lower_cost"
	case acCmp == 0 && costCmp == 0:
		return "same_outcome_same_cost"
	case acCmp < 0 && costCmp > 0:
		return "regressed_with_higher_cost"
	case acCmp < 0 && costCmp < 0:
		return "regressed_with_lower_cost"
	default:
		return "mixed"
	}
}

func compareInt(left, right int) int {
	switch {
	case left > right:
		return 1
	case left < right:
		return -1
	default:
		return 0
	}
}

func compareInt64(left, right int64) int {
	switch {
	case left > right:
		return 1
	case left < right:
		return -1
	default:
		return 0
	}
}

func compareFloat64(left, right float64) int {
	switch {
	case left > right:
		return 1
	case left < right:
		return -1
	default:
		return 0
	}
}

func (s *ExperimentCompareService) failCompare(ctx context.Context, compare *model.ExperimentCompare, runErr error) (*ExperimentCompareOutput, error) {
	compare.Status = model.ExperimentCompareStatusFailed
	compare.ErrorMessage = runErr.Error()
	if err := s.compares.Update(ctx, compare); err != nil {
		return nil, fmt.Errorf("update experiment compare: %w", err)
	}
	return nil, runErr
}

func decodeProblemIDs(raw string) ([]uint, error) {
	if raw == "" {
		return nil, nil
	}
	var values []uint
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil, err
	}
	return values, nil
}

func defaultCompareName(name string) string {
	if name = firstNonEmpty(name); name != "" {
		return name
	}
	return "compare-experiment-" + time.Now().UTC().Format("20060102-150405")
}

func derefUint(value *uint) uint {
	if value == nil {
		return 0
	}
	return *value
}

func buildCompareProblemSummaries(problemIDs []uint, baseline, candidate *ExperimentOutput) []ExperimentCompareProblemSummary {
	baselineRuns := mapExperimentRunsByProblemID(baseline)
	candidateRuns := mapExperimentRunsByProblemID(candidate)
	summaries := make([]ExperimentCompareProblemSummary, 0, len(problemIDs))
	for _, problemID := range problemIDs {
		baseRun := baselineRuns[problemID]
		candRun := candidateRuns[problemID]
		summaries = append(summaries, ExperimentCompareProblemSummary{
			ProblemID:             problemID,
			BaselineVerdict:       verdictOfRun(baseRun),
			CandidateVerdict:      verdictOfRun(candRun),
			Changed:               verdictOfRun(baseRun) != verdictOfRun(candRun),
			ChangeType:            compareChangeType(verdictOfRun(baseRun), verdictOfRun(candRun)),
			BaselineStatus:        statusOfRun(baseRun),
			CandidateStatus:       statusOfRun(candRun),
			BaselineSubmissionID:  submissionIDOfRun(baseRun),
			CandidateSubmissionID: submissionIDOfRun(candRun),
		})
	}
	return summaries
}

func summarizeCompareProblemChanges(summaries []ExperimentCompareProblemSummary) (improved, regressed, changedNonAC int) {
	for _, summary := range summaries {
		switch summary.ChangeType {
		case "improved":
			improved++
		case "regressed":
			regressed++
		case "changed_non_ac":
			changedNonAC++
		}
	}
	return
}

func mapExperimentRunsByProblemID(output *ExperimentOutput) map[uint]ExperimentRunOutput {
	if output == nil {
		return nil
	}
	indexed := make(map[uint]ExperimentRunOutput, len(output.Runs))
	for _, run := range output.Runs {
		indexed[run.ProblemID] = run
	}
	return indexed
}

func compareChangeType(baselineVerdict, candidateVerdict string) string {
	if baselineVerdict == candidateVerdict {
		return "same"
	}
	if baselineVerdict != modelVerdictAccepted && candidateVerdict == modelVerdictAccepted {
		return "improved"
	}
	if baselineVerdict == modelVerdictAccepted && candidateVerdict != modelVerdictAccepted {
		return "regressed"
	}
	return "changed_non_ac"
}

func verdictOfRun(run ExperimentRunOutput) string {
	return run.Verdict
}

func statusOfRun(run ExperimentRunOutput) string {
	return run.Status
}

func submissionIDOfRun(run ExperimentRunOutput) *uint {
	return run.SubmissionID
}

func buildHighlightedCompareProblems(problemSummaries []ExperimentCompareProblemSummary) []ExperimentCompareHighlightedProblem {
	highlighted := make([]ExperimentCompareHighlightedProblem, 0, len(problemSummaries))
	for _, problem := range problemSummaries {
		highlighted = append(highlighted, ExperimentCompareHighlightedProblem{
			ProblemID:             problem.ProblemID,
			BaselineVerdict:       problem.BaselineVerdict,
			CandidateVerdict:      problem.CandidateVerdict,
			Changed:               problem.Changed,
			ChangeType:            problem.ChangeType,
			BaselineSubmissionID:  problem.BaselineSubmissionID,
			CandidateSubmissionID: problem.CandidateSubmissionID,
		})
	}

	sort.SliceStable(highlighted, func(i, j int) bool {
		left := compareHighlightPriority(highlighted[i].ChangeType)
		right := compareHighlightPriority(highlighted[j].ChangeType)
		if left != right {
			return left < right
		}
		return highlighted[i].ProblemID < highlighted[j].ProblemID
	})

	return highlighted
}

func compareHighlightPriority(changeType string) int {
	switch changeType {
	case "regressed":
		return 0
	case "improved":
		return 1
	case "changed_non_ac":
		return 2
	case "same":
		return 3
	default:
		return 4
	}
}
