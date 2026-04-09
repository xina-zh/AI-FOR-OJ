package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"ai-for-oj/internal/model"
	"ai-for-oj/internal/repository"
)

const ExperimentCompareDimensionModel = "model"

type ExperimentRunner interface {
	Run(ctx context.Context, input RunExperimentInput) (*ExperimentOutput, error)
	Get(ctx context.Context, experimentID uint) (*ExperimentOutput, error)
}

type CompareExperimentInput struct {
	Name           string
	ProblemIDs     []uint
	BaselineModel  string
	CandidateModel string
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

type ExperimentCompareOutput struct {
	ID                    uint                                  `json:"id"`
	Name                  string                                `json:"name"`
	CompareDimension      string                                `json:"compare_dimension"`
	BaselineValue         string                                `json:"baseline_value"`
	CandidateValue        string                                `json:"candidate_value"`
	ProblemIDs            []uint                                `json:"problem_ids"`
	BaselineExperimentID  uint                                  `json:"baseline_experiment_id"`
	CandidateExperimentID uint                                  `json:"candidate_experiment_id"`
	BaselineSummary       *ExperimentOutput                     `json:"baseline_summary,omitempty"`
	CandidateSummary      *ExperimentOutput                     `json:"candidate_summary,omitempty"`
	BaselineDistribution  VerdictDistribution                   `json:"baseline_verdict_distribution"`
	CandidateDistribution VerdictDistribution                   `json:"candidate_verdict_distribution"`
	DeltaDistribution     VerdictDistribution                   `json:"delta_verdict_distribution"`
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

	compare := &model.ExperimentCompare{
		Name:             defaultCompareName(input.Name),
		CompareDimension: ExperimentCompareDimensionModel,
		BaselineValue:    baselineModel,
		CandidateValue:   candidateModel,
		ProblemIDs:       string(problemIDsJSON),
		Status:           model.ExperimentCompareStatusRunning,
	}
	if err := s.compares.Create(ctx, compare); err != nil {
		return nil, fmt.Errorf("create experiment compare: %w", err)
	}

	baseline, err := s.experiments.Run(ctx, RunExperimentInput{
		Name:       compare.Name + "-baseline",
		ProblemIDs: input.ProblemIDs,
		Model:      baselineModel,
	})
	if err != nil {
		return s.failCompare(ctx, compare, err)
	}
	compare.BaselineExperimentID = &baseline.ID

	candidate, err := s.experiments.Run(ctx, RunExperimentInput{
		Name:       compare.Name + "-candidate",
		ProblemIDs: input.ProblemIDs,
		Model:      candidateModel,
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

	return &ExperimentCompareOutput{
		ID:                    compare.ID,
		Name:                  compare.Name,
		CompareDimension:      compare.CompareDimension,
		BaselineValue:         compare.BaselineValue,
		CandidateValue:        compare.CandidateValue,
		ProblemIDs:            append([]uint(nil), input.ProblemIDs...),
		BaselineExperimentID:  baseline.ID,
		CandidateExperimentID: candidate.ID,
		BaselineSummary:       baseline,
		CandidateSummary:      candidate,
		BaselineDistribution:  baseline.VerdictDistribution,
		CandidateDistribution: candidate.VerdictDistribution,
		DeltaDistribution:     DiffVerdictDistribution(candidate.VerdictDistribution, baseline.VerdictDistribution),
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

	return &ExperimentCompareOutput{
		ID:                    compare.ID,
		Name:                  compare.Name,
		CompareDimension:      compare.CompareDimension,
		BaselineValue:         compare.BaselineValue,
		CandidateValue:        compare.CandidateValue,
		ProblemIDs:            problemIDs,
		BaselineExperimentID:  derefUint(compare.BaselineExperimentID),
		CandidateExperimentID: derefUint(compare.CandidateExperimentID),
		BaselineSummary:       baseline,
		CandidateSummary:      candidate,
		BaselineDistribution:  verdictDistributionOf(baseline),
		CandidateDistribution: verdictDistributionOf(candidate),
		DeltaDistribution:     DiffVerdictDistribution(verdictDistributionOf(candidate), verdictDistributionOf(baseline)),
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
