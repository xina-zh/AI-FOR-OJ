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
	"ai-for-oj/internal/tooling"
)

type RepeatExperimentInput struct {
	Name          string
	ProblemIDs    []uint
	Model         string
	PromptName    string
	AgentName     string
	ToolingConfig string
	RepeatCount   int
}

type ExperimentRepeatRoundSummary struct {
	RoundNo             int                 `json:"round_no"`
	ExperimentID        uint                `json:"experiment_id"`
	ACCount             int                 `json:"ac_count"`
	FailedCount         int                 `json:"failed_count"`
	VerdictDistribution VerdictDistribution `json:"verdict_distribution"`
	Status              string              `json:"status"`
}

type ExperimentRepeatProblemSummary struct {
	ProblemID           uint                `json:"problem_id"`
	TotalRounds         int                 `json:"total_rounds"`
	ACCount             int                 `json:"ac_count"`
	FailedCount         int                 `json:"failed_count"`
	ACRate              float64             `json:"ac_rate"`
	VerdictDistribution VerdictDistribution `json:"verdict_distribution"`
	LatestVerdict       string              `json:"latest_verdict,omitempty"`
}

type ExperimentRepeatUnstableProblem struct {
	ProblemID           uint                `json:"problem_id"`
	TotalRounds         int                 `json:"total_rounds"`
	ACCount             int                 `json:"ac_count"`
	FailedCount         int                 `json:"failed_count"`
	ACRate              float64             `json:"ac_rate"`
	VerdictDistribution VerdictDistribution `json:"verdict_distribution"`
	LatestVerdict       string              `json:"latest_verdict,omitempty"`
	InstabilityScore    int                 `json:"instability_score"`
	VerdictKindCount    int                 `json:"verdict_kind_count"`
}

type ExperimentRepeatCostSummary struct {
	TotalTokenInput               int64   `json:"total_token_input"`
	TotalTokenOutput              int64   `json:"total_token_output"`
	TotalTokens                   int64   `json:"total_tokens"`
	AverageTokenInputPerRound     float64 `json:"average_token_input_per_round"`
	AverageTokenOutputPerRound    float64 `json:"average_token_output_per_round"`
	AverageTotalTokensPerRound    float64 `json:"average_total_tokens_per_round"`
	TotalLLMLatencyMS             int     `json:"total_llm_latency_ms"`
	TotalLatencyMS                int     `json:"total_latency_ms"`
	AverageLLMLatencyMSPerRound   float64 `json:"average_llm_latency_ms_per_round"`
	AverageTotalLatencyMSPerRound float64 `json:"average_total_latency_ms_per_round"`
	RoundCount                    int     `json:"round_count"`
}

type ExperimentRepeatOutput struct {
	ID                         uint                              `json:"id"`
	Name                       string                            `json:"name"`
	Model                      string                            `json:"model"`
	PromptName                 string                            `json:"prompt_name"`
	AgentName                  string                            `json:"agent_name"`
	ToolingConfig              string                            `json:"tooling_config"`
	ProblemIDs                 []uint                            `json:"problem_ids"`
	RepeatCount                int                               `json:"repeat_count"`
	ExperimentIDs              []uint                            `json:"experiment_ids"`
	TotalProblemCount          int                               `json:"total_problem_count"`
	TotalRunCount              int                               `json:"total_run_count"`
	OverallACCount             int                               `json:"overall_ac_count"`
	OverallFailedCount         int                               `json:"overall_failed_count"`
	AverageACCountPerRound     float64                           `json:"average_ac_count_per_round"`
	AverageFailedCountPerRound float64                           `json:"average_failed_count_per_round"`
	OverallACRate              float64                           `json:"overall_ac_rate"`
	BestRoundACCount           int                               `json:"best_round_ac_count"`
	WorstRoundACCount          int                               `json:"worst_round_ac_count"`
	CostSummary                ExperimentRepeatCostSummary       `json:"cost_summary"`
	Status                     string                            `json:"status"`
	ErrorMessage               string                            `json:"error_message,omitempty"`
	RoundSummaries             []ExperimentRepeatRoundSummary    `json:"round_summaries"`
	ProblemSummaries           []ExperimentRepeatProblemSummary  `json:"problem_summaries"`
	MostUnstableProblems       []ExperimentRepeatUnstableProblem `json:"most_unstable_problems"`
	CreatedAt                  time.Time                         `json:"created_at"`
	UpdatedAt                  time.Time                         `json:"updated_at"`
}

type ExperimentRepeatListInput struct {
	Page     int
	PageSize int
}

type ExperimentRepeatListOutput struct {
	Items      []ExperimentRepeatOutput `json:"items"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"page_size"`
	Total      int64                    `json:"total"`
	TotalPages int                      `json:"total_pages"`
}

type ExperimentRepeatService struct {
	repeats      repository.ExperimentRepeatRepository
	experiments  ExperimentRunner
	defaultModel string
}

func NewExperimentRepeatService(
	repeats repository.ExperimentRepeatRepository,
	experiments ExperimentRunner,
	defaultModel string,
) *ExperimentRepeatService {
	return &ExperimentRepeatService{
		repeats:      repeats,
		experiments:  experiments,
		defaultModel: defaultModel,
	}
}

func (s *ExperimentRepeatService) Repeat(ctx context.Context, input RepeatExperimentInput) (*ExperimentRepeatOutput, error) {
	problemIDsJSON, err := json.Marshal(input.ProblemIDs)
	if err != nil {
		return nil, fmt.Errorf("marshal problem ids: %w", err)
	}
	resolvedPromptName, err := prompt.ResolveSolvePromptName(input.PromptName)
	if err != nil {
		return nil, err
	}
	resolvedAgentName, err := agent.ResolveSolveAgentName(input.AgentName)
	if err != nil {
		return nil, err
	}
	_, canonicalToolingConfig, err := tooling.ResolveConfig(input.ToolingConfig)
	if err != nil {
		return nil, err
	}

	repeat := &model.ExperimentRepeat{
		Name:          defaultRepeatName(input.Name),
		ModelName:     firstNonEmpty(input.Model, s.defaultModel),
		PromptName:    resolvedPromptName,
		AgentName:     resolvedAgentName,
		ToolingConfig: canonicalToolingConfig,
		ProblemIDs:    string(problemIDsJSON),
		RepeatCount:   input.RepeatCount,
		Status:        model.ExperimentRepeatStatusRunning,
	}
	if err := s.repeats.Create(ctx, repeat); err != nil {
		return nil, fmt.Errorf("create experiment repeat: %w", err)
	}

	experimentIDs := make([]uint, 0, input.RepeatCount)
	rounds := make([]*ExperimentOutput, 0, input.RepeatCount)
	for round := 1; round <= input.RepeatCount; round++ {
		experiment, err := s.experiments.Run(ctx, RunExperimentInput{
			Name:          fmt.Sprintf("%s-round-%d", repeat.Name, round),
			ProblemIDs:    input.ProblemIDs,
			Model:         repeat.ModelName,
			PromptName:    repeat.PromptName,
			AgentName:     repeat.AgentName,
			ToolingConfig: repeat.ToolingConfig,
		})
		if err != nil {
			return s.failRepeat(ctx, repeat, experimentIDs, err)
		}
		experimentIDs = append(experimentIDs, experiment.ID)
		rounds = append(rounds, experiment)
	}

	experimentIDsJSON, err := json.Marshal(experimentIDs)
	if err != nil {
		return nil, fmt.Errorf("marshal experiment ids: %w", err)
	}
	repeat.ExperimentIDs = string(experimentIDsJSON)
	repeat.Status = model.ExperimentRepeatStatusCompleted
	if err := s.repeats.Update(ctx, repeat); err != nil {
		return nil, fmt.Errorf("update experiment repeat: %w", err)
	}

	return buildExperimentRepeatOutput(repeat, input.ProblemIDs, experimentIDs, rounds), nil
}

func (s *ExperimentRepeatService) Get(ctx context.Context, repeatID uint) (*ExperimentRepeatOutput, error) {
	repeat, err := s.repeats.GetByID(ctx, repeatID)
	if err != nil {
		return nil, err
	}

	problemIDs, err := decodeUintSlice(repeat.ProblemIDs)
	if err != nil {
		return nil, fmt.Errorf("decode repeat problem ids: %w", err)
	}
	experimentIDs, err := decodeUintSlice(repeat.ExperimentIDs)
	if err != nil {
		return nil, fmt.Errorf("decode repeat experiment ids: %w", err)
	}

	rounds := make([]*ExperimentOutput, 0, len(experimentIDs))
	for _, experimentID := range experimentIDs {
		experiment, err := s.experiments.Get(ctx, experimentID)
		if err != nil {
			return nil, err
		}
		rounds = append(rounds, experiment)
	}

	return buildExperimentRepeatOutput(repeat, problemIDs, experimentIDs, rounds), nil
}

func (s *ExperimentRepeatService) List(ctx context.Context, input ExperimentRepeatListInput) (*ExperimentRepeatListOutput, error) {
	repeats, total, err := s.repeats.List(ctx, repository.ExperimentRepeatListQuery{
		Page:     input.Page,
		PageSize: input.PageSize,
	})
	if err != nil {
		return nil, fmt.Errorf("list experiment repeats: %w", err)
	}

	items := make([]ExperimentRepeatOutput, 0, len(repeats))
	for _, repeat := range repeats {
		problemIDs, err := decodeUintSlice(repeat.ProblemIDs)
		if err != nil {
			return nil, fmt.Errorf("decode repeat problem ids: %w", err)
		}
		experimentIDs, err := decodeUintSlice(repeat.ExperimentIDs)
		if err != nil {
			return nil, fmt.Errorf("decode repeat experiment ids: %w", err)
		}
		items = append(items, *buildExperimentRepeatOutput(&repeat, problemIDs, experimentIDs, nil))
	}

	return &ExperimentRepeatListOutput{
		Items:      items,
		Page:       input.Page,
		PageSize:   input.PageSize,
		Total:      total,
		TotalPages: totalPages(total, input.PageSize),
	}, nil
}

func buildExperimentRepeatOutput(
	repeat *model.ExperimentRepeat,
	problemIDs []uint,
	experimentIDs []uint,
	rounds []*ExperimentOutput,
) *ExperimentRepeatOutput {
	output := &ExperimentRepeatOutput{
		ID:                   repeat.ID,
		Name:                 repeat.Name,
		Model:                repeat.ModelName,
		PromptName:           prompt.DisplaySolvePromptName(repeat.PromptName),
		AgentName:            agent.DisplaySolveAgentName(repeat.AgentName),
		ToolingConfig:        repeat.ToolingConfig,
		ProblemIDs:           append([]uint(nil), problemIDs...),
		RepeatCount:          repeat.RepeatCount,
		ExperimentIDs:        append([]uint(nil), experimentIDs...),
		TotalProblemCount:    len(problemIDs),
		TotalRunCount:        len(problemIDs) * repeat.RepeatCount,
		CostSummary:          buildExperimentRepeatCostSummary(rounds),
		Status:               repeat.Status,
		ErrorMessage:         repeat.ErrorMessage,
		RoundSummaries:       make([]ExperimentRepeatRoundSummary, 0, len(rounds)),
		ProblemSummaries:     make([]ExperimentRepeatProblemSummary, 0, len(problemIDs)),
		MostUnstableProblems: make([]ExperimentRepeatUnstableProblem, 0, len(problemIDs)),
		CreatedAt:            repeat.CreatedAt,
		UpdatedAt:            repeat.UpdatedAt,
	}
	if len(rounds) == 0 {
		return output
	}

	output.BestRoundACCount = rounds[0].ACCount
	output.WorstRoundACCount = rounds[0].ACCount

	for index, round := range rounds {
		output.OverallACCount += round.ACCount
		output.OverallFailedCount += round.FailedCount
		if round.ACCount > output.BestRoundACCount {
			output.BestRoundACCount = round.ACCount
		}
		if round.ACCount < output.WorstRoundACCount {
			output.WorstRoundACCount = round.ACCount
		}
		output.RoundSummaries = append(output.RoundSummaries, ExperimentRepeatRoundSummary{
			RoundNo:             index + 1,
			ExperimentID:        round.ID,
			ACCount:             round.ACCount,
			FailedCount:         round.FailedCount,
			VerdictDistribution: round.VerdictDistribution,
			Status:              round.Status,
		})
	}

	output.AverageACCountPerRound = float64(output.OverallACCount) / float64(len(rounds))
	output.AverageFailedCountPerRound = float64(output.OverallFailedCount) / float64(len(rounds))
	if output.TotalRunCount > 0 {
		output.OverallACRate = float64(output.OverallACCount) / float64(output.TotalRunCount)
	}
	output.ProblemSummaries = buildRepeatProblemSummaries(problemIDs, rounds)
	output.MostUnstableProblems = buildMostUnstableProblems(output.ProblemSummaries)
	return output
}

func buildExperimentRepeatCostSummary(rounds []*ExperimentOutput) ExperimentRepeatCostSummary {
	var summary ExperimentRepeatCostSummary
	for _, round := range rounds {
		if round == nil || round.CostSummary.RunCount == 0 {
			continue
		}

		summary.RoundCount++
		summary.TotalTokenInput += round.CostSummary.TotalTokenInput
		summary.TotalTokenOutput += round.CostSummary.TotalTokenOutput
		summary.TotalLLMLatencyMS += round.CostSummary.TotalLLMLatencyMS
		summary.TotalLatencyMS += round.CostSummary.TotalLatencyMS
	}

	summary.TotalTokens = summary.TotalTokenInput + summary.TotalTokenOutput
	if summary.RoundCount == 0 {
		return summary
	}

	roundCount := float64(summary.RoundCount)
	summary.AverageTokenInputPerRound = float64(summary.TotalTokenInput) / roundCount
	summary.AverageTokenOutputPerRound = float64(summary.TotalTokenOutput) / roundCount
	summary.AverageTotalTokensPerRound = float64(summary.TotalTokens) / roundCount
	summary.AverageLLMLatencyMSPerRound = float64(summary.TotalLLMLatencyMS) / roundCount
	summary.AverageTotalLatencyMSPerRound = float64(summary.TotalLatencyMS) / roundCount
	return summary
}

func buildRepeatProblemSummaries(problemIDs []uint, rounds []*ExperimentOutput) []ExperimentRepeatProblemSummary {
	type accumulator struct {
		totalRounds         int
		acCount             int
		failedCount         int
		verdictDistribution VerdictDistribution
		latestVerdict       string
	}

	accs := make(map[uint]*accumulator, len(problemIDs))
	for _, problemID := range problemIDs {
		accs[problemID] = &accumulator{}
	}

	for _, round := range rounds {
		for _, run := range round.Runs {
			acc, ok := accs[run.ProblemID]
			if !ok {
				continue
			}
			acc.totalRounds++
			acc.latestVerdict = run.Verdict
			if run.Verdict == modelVerdictAccepted {
				acc.acCount++
			} else {
				acc.failedCount++
			}
			if run.Verdict != "" {
				acc.verdictDistribution.Add(run.Verdict)
			} else {
				acc.verdictDistribution.Add("")
			}
		}
	}

	summaries := make([]ExperimentRepeatProblemSummary, 0, len(problemIDs))
	for _, problemID := range problemIDs {
		acc := accs[problemID]
		summary := ExperimentRepeatProblemSummary{
			ProblemID:           problemID,
			TotalRounds:         acc.totalRounds,
			ACCount:             acc.acCount,
			FailedCount:         acc.failedCount,
			VerdictDistribution: acc.verdictDistribution,
			LatestVerdict:       acc.latestVerdict,
		}
		if acc.totalRounds > 0 {
			summary.ACRate = float64(acc.acCount) / float64(acc.totalRounds)
		}
		summaries = append(summaries, summary)
	}

	return summaries
}

func buildMostUnstableProblems(problemSummaries []ExperimentRepeatProblemSummary) []ExperimentRepeatUnstableProblem {
	unstable := make([]ExperimentRepeatUnstableProblem, 0, len(problemSummaries))
	for _, summary := range problemSummaries {
		unstable = append(unstable, ExperimentRepeatUnstableProblem{
			ProblemID:           summary.ProblemID,
			TotalRounds:         summary.TotalRounds,
			ACCount:             summary.ACCount,
			FailedCount:         summary.FailedCount,
			ACRate:              summary.ACRate,
			VerdictDistribution: summary.VerdictDistribution,
			LatestVerdict:       summary.LatestVerdict,
			InstabilityScore:    minInt(summary.ACCount, summary.FailedCount),
			VerdictKindCount:    verdictKindCount(summary.VerdictDistribution),
		})
	}

	sort.SliceStable(unstable, func(i, j int) bool {
		if unstable[i].InstabilityScore != unstable[j].InstabilityScore {
			return unstable[i].InstabilityScore > unstable[j].InstabilityScore
		}
		if unstable[i].VerdictKindCount != unstable[j].VerdictKindCount {
			return unstable[i].VerdictKindCount > unstable[j].VerdictKindCount
		}
		return unstable[i].ProblemID < unstable[j].ProblemID
	})

	return unstable
}

func verdictKindCount(dist VerdictDistribution) int {
	count := 0
	if dist.ACCount > 0 {
		count++
	}
	if dist.WACount > 0 {
		count++
	}
	if dist.CECount > 0 {
		count++
	}
	if dist.RECount > 0 {
		count++
	}
	if dist.TLECount > 0 {
		count++
	}
	if dist.UnjudgeableCount > 0 {
		count++
	}
	if dist.UnknownCount > 0 {
		count++
	}
	return count
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *ExperimentRepeatService) failRepeat(
	ctx context.Context,
	repeat *model.ExperimentRepeat,
	experimentIDs []uint,
	runErr error,
) (*ExperimentRepeatOutput, error) {
	if len(experimentIDs) > 0 {
		if raw, err := json.Marshal(experimentIDs); err == nil {
			repeat.ExperimentIDs = string(raw)
		}
	}
	repeat.Status = model.ExperimentRepeatStatusFailed
	repeat.ErrorMessage = runErr.Error()
	if err := s.repeats.Update(ctx, repeat); err != nil {
		return nil, fmt.Errorf("update experiment repeat: %w", err)
	}
	return nil, runErr
}

func defaultRepeatName(name string) string {
	if name = firstNonEmpty(name); name != "" {
		return name
	}
	return "repeat-experiment-" + time.Now().UTC().Format("20060102-150405")
}

func decodeUintSlice(raw string) ([]uint, error) {
	if raw == "" {
		return nil, nil
	}
	var values []uint
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil, err
	}
	return values, nil
}
