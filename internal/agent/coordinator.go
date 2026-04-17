package agent

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"ai-for-oj/internal/llm"
	"ai-for-oj/internal/model"
	"ai-for-oj/internal/prompt"
)

const (
	adaptiveRepairInitialStage  = "initial_solve"
	defaultAdaptiveRepairBudget = 3
)

type JudgeFeedback struct {
	Verdict       string
	TimedOut      bool
	CompileStderr string
	RunStdout     string
	RunStderr     string
	ErrorMessage  string
	PassedCount   int
	TotalCount    int
	ExecStage     string
}

type JudgeSubmitter interface {
	Submit(ctx context.Context, sourceCode string) (*JudgeFeedback, error)
}

type AdaptiveRepairAttempt struct {
	AttemptNo     int
	Stage         string
	FailureType   FailureType
	PromptPreview string
	RawResponse   string
	Model         string
	TokenInput    int64
	TokenOutput   int64
	LLMLatencyMS  int
	JudgeFeedback JudgeFeedback
}

type AdaptiveRepairResult struct {
	SolveOutput   SolveOutput
	AttemptCount  int
	StrategyPath  []string
	Attempts      []AdaptiveRepairAttempt
	FinalFeedback *JudgeFeedback
	FinalFailure  FailureType
}

type AdaptiveRepairCoordinator struct {
	planner RepairPlanner
}

func NewAdaptiveRepairCoordinator(maxBudget int) AdaptiveRepairCoordinator {
	if maxBudget <= 0 {
		maxBudget = defaultAdaptiveRepairBudget
	}
	return AdaptiveRepairCoordinator{planner: NewRepairPlanner(maxBudget)}
}

func (c AdaptiveRepairCoordinator) Execute(ctx context.Context, client llm.Client, input SolveInput) (AdaptiveRepairResult, error) {
	if input.Problem == nil {
		return AdaptiveRepairResult{
			SolveOutput: SolveOutput{
				AgentName: AdaptiveRepairV1AgentName,
				Model:     input.Model,
			},
		}, fmt.Errorf("problem is required")
	}

	planner := c.planner
	if planner.maxBudget <= 0 {
		planner = NewRepairPlanner(defaultAdaptiveRepairBudget)
	}

	attempts := make([]AdaptiveRepairAttempt, 0, 4)
	strategyPath := make([]string, 0, 4)
	var totalInput int64
	var totalOutput int64
	var totalLatency int
	modelName := input.Model
	previousCode := ""
	currentStage := adaptiveRepairInitialStage
	var finalFeedback *JudgeFeedback
	var finalFailure FailureType = FailureTypeUnknown
	var lastPromptPreview string
	var lastRawResponse string

	for {
		var promptText string
		if len(attempts) == 0 {
			promptText = prompt.BuildSolvePrompt(input.Problem, input.PromptName)
		} else {
			promptText = buildAdaptiveRepairPrompt(input.Problem, input.PromptName, currentStage, previousCode, finalFeedback)
		}

		execution, err := executeLLMOnce(ctx, client, modelName, promptText)
		totalInput += execution.TokenInput
		totalOutput += execution.TokenOutput
		totalLatency += execution.LLMLatencyMS
		modelName = effectiveModel(execution.Model, modelName)
		lastPromptPreview = execution.PromptPreview
		lastRawResponse = execution.RawResponse
		previousCode = extractCPPCode(execution.RawResponse)

		attempt := AdaptiveRepairAttempt{
			AttemptNo:     len(attempts) + 1,
			Stage:         currentStage,
			PromptPreview: execution.PromptPreview,
			RawResponse:   execution.RawResponse,
			Model:         modelName,
			TokenInput:    execution.TokenInput,
			TokenOutput:   execution.TokenOutput,
			LLMLatencyMS:  execution.LLMLatencyMS,
		}
		attempts = append(attempts, attempt)

		if err != nil {
			return c.buildResult(modelName, lastPromptPreview, lastRawResponse, totalInput, totalOutput, totalLatency, attempts, strategyPath, finalFeedback, finalFailure), err
		}

		if input.JudgeSubmitter == nil {
			return c.buildResult(modelName, lastPromptPreview, lastRawResponse, totalInput, totalOutput, totalLatency, attempts, strategyPath, finalFeedback, finalFailure), nil
		}

		feedback, err := input.JudgeSubmitter.Submit(ctx, previousCode)
		if err != nil {
			return c.buildResult(modelName, lastPromptPreview, lastRawResponse, totalInput, totalOutput, totalLatency, attempts, strategyPath, finalFeedback, finalFailure), err
		}
		if feedback != nil {
			attempts[len(attempts)-1].JudgeFeedback = *feedback
			finalFeedback = feedback
			finalFailure = classifyJudgeFeedback(feedback)
			attempts[len(attempts)-1].FailureType = finalFailure
		}

		if feedback != nil && strings.EqualFold(strings.TrimSpace(feedback.Verdict), "AC") {
			return c.buildResult(modelName, lastPromptPreview, lastRawResponse, totalInput, totalOutput, totalLatency, attempts, strategyPath, finalFeedback, finalFailure), nil
		}

		decision := planner.Next(RepairPlanInput{
			AttemptCount:   len(attempts),
			LastFailure:    finalFailure,
			PreviousStages: strategyPath,
		})
		if decision.Stop {
			return c.buildResult(modelName, lastPromptPreview, lastRawResponse, totalInput, totalOutput, totalLatency, attempts, strategyPath, finalFeedback, finalFailure), nil
		}

		strategyPath = append(strategyPath, decision.Stage)
		currentStage = decision.Stage
	}
}

func (c AdaptiveRepairCoordinator) buildResult(
	modelName string,
	promptPreview string,
	rawResponse string,
	totalInput int64,
	totalOutput int64,
	totalLatency int,
	attempts []AdaptiveRepairAttempt,
	strategyPath []string,
	finalFeedback *JudgeFeedback,
	finalFailure FailureType,
) AdaptiveRepairResult {
	result := AdaptiveRepairResult{
		SolveOutput: SolveOutput{
			AgentName:     AdaptiveRepairV1AgentName,
			Model:         modelName,
			PromptPreview: promptPreview,
			RawResponse:   rawResponse,
			TokenInput:    totalInput,
			TokenOutput:   totalOutput,
			LLMLatencyMS:  totalLatency,
		},
		AttemptCount: len(attempts),
		StrategyPath: append([]string(nil), strategyPath...),
		Attempts:     append([]AdaptiveRepairAttempt(nil), attempts...),
		FinalFailure: finalFailure,
	}
	if finalFeedback != nil {
		copyFeedback := *finalFeedback
		result.FinalFeedback = &copyFeedback
	}
	return result
}

func classifyJudgeFeedback(feedback *JudgeFeedback) FailureType {
	if feedback == nil {
		return FailureTypeUnknown
	}
	return ClassifyFailure(JudgeFailureObservation{
		Verdict:       feedback.Verdict,
		TimedOut:      feedback.TimedOut,
		CompileStderr: feedback.CompileStderr,
		RunStderr:     feedback.RunStderr,
		PassedCount:   feedback.PassedCount,
		TotalCount:    feedback.TotalCount,
		ExecStage:     feedback.ExecStage,
	})
}

func buildAdaptiveRepairPrompt(problem *model.Problem, promptName, stage, previousCode string, feedback *JudgeFeedback) string {
	feedbackText := renderJudgeFeedback(feedback)
	switch stage {
	case RepairStageWAAnalysisRepair:
		return prompt.BuildWARepairPrompt(problem, promptName, previousCode, feedbackText)
	case RepairStageRESafetyRepair:
		return prompt.BuildRERepairPrompt(problem, promptName, previousCode, feedbackText)
	case RepairStageTLEComplexityRewrite:
		return prompt.BuildTLERepairPrompt(problem, promptName, previousCode, feedbackText)
	default:
		return prompt.BuildRepairPrompt(problem, promptName, previousCode, feedbackText)
	}
}

func renderJudgeFeedback(feedback *JudgeFeedback) string {
	if feedback == nil {
		return "No judge feedback is available."
	}

	lines := []string{
		fmt.Sprintf("Verdict: %s", firstNonEmptyString(strings.TrimSpace(feedback.Verdict), "unknown")),
		fmt.Sprintf("Passed Count: %d / %d", feedback.PassedCount, feedback.TotalCount),
	}
	if strings.TrimSpace(feedback.ExecStage) != "" {
		lines = append(lines, "Execution Stage: "+strings.TrimSpace(feedback.ExecStage))
	}
	if feedback.TimedOut {
		lines = append(lines, "Timed Out: true")
	}
	if strings.TrimSpace(feedback.ErrorMessage) != "" {
		lines = append(lines, "Error Message:\n"+strings.TrimSpace(feedback.ErrorMessage))
	}
	if strings.TrimSpace(feedback.CompileStderr) != "" {
		lines = append(lines, "Compile Stderr:\n"+strings.TrimSpace(feedback.CompileStderr))
	}
	if strings.TrimSpace(feedback.RunStderr) != "" {
		lines = append(lines, "Run Stderr:\n"+strings.TrimSpace(feedback.RunStderr))
	}
	if strings.TrimSpace(feedback.RunStdout) != "" {
		lines = append(lines, "Run Stdout:\n"+strings.TrimSpace(feedback.RunStdout))
	}
	return strings.Join(lines, "\n\n")
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

var (
	cppFencePattern     = regexp.MustCompile("(?is)```(?:cpp|c\\+\\+|cc|cxx)\\s*(.*?)```")
	genericFencePattern = regexp.MustCompile("(?is)```(?:[a-z0-9_+-]+)?\\s*(.*?)```")
)

func extractCPPCode(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	if matches := cppFencePattern.FindStringSubmatch(raw); len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}
	if matches := genericFencePattern.FindStringSubmatch(raw); len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}
	return strings.TrimSpace(strings.Trim(raw, "`"))
}
