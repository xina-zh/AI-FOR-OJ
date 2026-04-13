package service

import (
	"context"
	"fmt"
	"time"

	"ai-for-oj/internal/model"
	"ai-for-oj/internal/repository"
)

const (
	ExperimentStatusRunning    = "running"
	ExperimentStatusCompleted  = "completed"
	ExperimentRunStatusSuccess = "success"
	ExperimentRunStatusFailed  = "failed"
)

type AISolver interface {
	Solve(ctx context.Context, input AISolveInput) (*AISolveOutput, error)
}

type RunExperimentInput struct {
	Name       string
	ProblemIDs []uint
	Model      string
}

type ExperimentRunOutput struct {
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

type ExperimentCostSummary struct {
	TotalTokenInput       int64   `json:"total_token_input"`
	TotalTokenOutput      int64   `json:"total_token_output"`
	TotalTokens           int64   `json:"total_tokens"`
	AverageTokenInput     float64 `json:"average_token_input"`
	AverageTokenOutput    float64 `json:"average_token_output"`
	AverageTotalTokens    float64 `json:"average_total_tokens"`
	TotalLLMLatencyMS     int     `json:"total_llm_latency_ms"`
	TotalLatencyMS        int     `json:"total_latency_ms"`
	AverageLLMLatencyMS   float64 `json:"average_llm_latency_ms"`
	AverageTotalLatencyMS float64 `json:"average_total_latency_ms"`
	RunCount              int     `json:"run_count"`
}

type ExperimentOutput struct {
	ID                  uint                  `json:"id"`
	Name                string                `json:"name"`
	Model               string                `json:"model"`
	Status              string                `json:"status"`
	TotalCount          int                   `json:"total_count"`
	SuccessCount        int                   `json:"success_count"`
	ACCount             int                   `json:"ac_count"`
	FailedCount         int                   `json:"failed_count"`
	VerdictDistribution VerdictDistribution   `json:"verdict_distribution"`
	CostSummary         ExperimentCostSummary `json:"cost_summary"`
	CreatedAt           time.Time             `json:"created_at"`
	UpdatedAt           time.Time             `json:"updated_at"`
	Runs                []ExperimentRunOutput `json:"runs"`
}

type ExperimentService struct {
	experiments  repository.ExperimentRepository
	aiSolver     AISolver
	defaultModel string
}

func NewExperimentService(
	experiments repository.ExperimentRepository,
	aiSolver AISolver,
	defaultModel string,
) *ExperimentService {
	return &ExperimentService{
		experiments:  experiments,
		aiSolver:     aiSolver,
		defaultModel: defaultModel,
	}
}

func (s *ExperimentService) Run(ctx context.Context, input RunExperimentInput) (*ExperimentOutput, error) {
	experiment := &model.Experiment{
		Name:       defaultExperimentName(input.Name),
		ModelName:  firstNonEmpty(input.Model, s.defaultModel),
		Status:     ExperimentStatusRunning,
		TotalCount: len(input.ProblemIDs),
	}
	if err := s.experiments.Create(ctx, experiment); err != nil {
		return nil, fmt.Errorf("create experiment: %w", err)
	}

	for index, problemID := range input.ProblemIDs {
		aiOutput, err := s.aiSolver.Solve(ctx, AISolveInput{
			ProblemID: problemID,
			Model:     input.Model,
		})

		run := &model.ExperimentRun{
			ExperimentID: experiment.ID,
			ProblemID:    problemID,
			AttemptNo:    index + 1,
			Status:       ExperimentRunStatusSuccess,
		}

		if aiOutput != nil {
			if aiOutput.AISolveRunID != 0 {
				run.AISolveRunID = &aiOutput.AISolveRunID
				run.AISolveRun = &model.AISolveRun{
					BaseModel: model.BaseModel{
						ID: aiOutput.AISolveRunID,
					},
					TokenInput:     aiOutput.TokenInput,
					TokenOutput:    aiOutput.TokenOutput,
					LLMLatencyMS:   aiOutput.LLMLatencyMS,
					TotalLatencyMS: aiOutput.TotalLatencyMS,
				}
			}
			if aiOutput.SubmissionID != 0 {
				run.SubmissionID = &aiOutput.SubmissionID
			}
			run.FinalVerdict = aiOutput.Verdict
			run.ErrorMessage = aiOutput.ErrorMessage
		}

		if err != nil {
			run.Status = ExperimentRunStatusFailed
			run.ErrorMessage = firstNonEmpty(run.ErrorMessage, err.Error())
			experiment.FailedCount++
		} else {
			experiment.SuccessCount++
			if run.FinalVerdict == modelVerdictAccepted {
				experiment.ACCount++
			}
		}

		if err := s.experiments.CreateRun(ctx, run); err != nil {
			return nil, fmt.Errorf("create experiment run: %w", err)
		}
	}

	experiment.Status = ExperimentStatusCompleted
	if err := s.experiments.Update(ctx, experiment); err != nil {
		return nil, fmt.Errorf("update experiment: %w", err)
	}

	return s.Get(ctx, experiment.ID)
}

func (s *ExperimentService) Get(ctx context.Context, experimentID uint) (*ExperimentOutput, error) {
	experiment, err := s.experiments.GetByIDWithRuns(ctx, experimentID)
	if err != nil {
		return nil, err
	}
	return toExperimentOutput(experiment), nil
}

func toExperimentOutput(experiment *model.Experiment) *ExperimentOutput {
	output := &ExperimentOutput{
		ID:           experiment.ID,
		Name:         experiment.Name,
		Model:        experiment.ModelName,
		Status:       experiment.Status,
		TotalCount:   experiment.TotalCount,
		SuccessCount: experiment.SuccessCount,
		ACCount:      experiment.ACCount,
		FailedCount:  experiment.FailedCount,
		CostSummary:  buildExperimentCostSummary(experiment.Runs),
		CreatedAt:    experiment.CreatedAt,
		UpdatedAt:    experiment.UpdatedAt,
		Runs:         make([]ExperimentRunOutput, 0, len(experiment.Runs)),
	}
	for _, run := range experiment.Runs {
		if run.FinalVerdict != "" {
			output.VerdictDistribution.Add(run.FinalVerdict)
		} else if run.Status == ExperimentRunStatusFailed {
			output.VerdictDistribution.Add("")
		}
		output.Runs = append(output.Runs, ExperimentRunOutput{
			ID:           run.ID,
			ProblemID:    run.ProblemID,
			AISolveRunID: run.AISolveRunID,
			SubmissionID: run.SubmissionID,
			AttemptNo:    run.AttemptNo,
			Verdict:      run.FinalVerdict,
			Status:       run.Status,
			ErrorMessage: run.ErrorMessage,
			CreatedAt:    run.CreatedAt,
		})
	}
	return output
}

func buildExperimentCostSummary(runs []model.ExperimentRun) ExperimentCostSummary {
	var summary ExperimentCostSummary
	for _, run := range runs {
		if run.AISolveRunID == nil || run.AISolveRun == nil {
			continue
		}

		summary.RunCount++
		summary.TotalTokenInput += run.AISolveRun.TokenInput
		summary.TotalTokenOutput += run.AISolveRun.TokenOutput
		summary.TotalLLMLatencyMS += run.AISolveRun.LLMLatencyMS
		summary.TotalLatencyMS += run.AISolveRun.TotalLatencyMS
	}

	summary.TotalTokens = summary.TotalTokenInput + summary.TotalTokenOutput
	if summary.RunCount == 0 {
		return summary
	}

	runCount := float64(summary.RunCount)
	summary.AverageTokenInput = float64(summary.TotalTokenInput) / runCount
	summary.AverageTokenOutput = float64(summary.TotalTokenOutput) / runCount
	summary.AverageTotalTokens = float64(summary.TotalTokens) / runCount
	summary.AverageLLMLatencyMS = float64(summary.TotalLLMLatencyMS) / runCount
	summary.AverageTotalLatencyMS = float64(summary.TotalLatencyMS) / runCount
	return summary
}

func defaultExperimentName(name string) string {
	if name = firstNonEmpty(name); name != "" {
		return name
	}
	return "batch-experiment-" + time.Now().UTC().Format("20060102-150405")
}

const modelVerdictAccepted = "AC"
