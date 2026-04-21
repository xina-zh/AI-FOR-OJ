package agent

import (
	"context"
	"errors"
	"strings"
	"time"

	"ai-for-oj/internal/llm"
	"ai-for-oj/internal/prompt"
)

var ErrAdaptiveCodeNotExtracted = errors.New("failed to extract cpp17 code from adaptive repair response")

type JudgeResult struct {
	SubmissionID uint
	Verdict      string
	ErrorMessage string
	Feedback     string
}

type JudgeSubmitter interface {
	Submit(ctx context.Context, sourceCode string) (*JudgeResult, error)
}

type AttemptRecord struct {
	AttemptNo      int
	Stage          string
	Model          string
	PromptPreview  string
	RawResponse    string
	ExtractedCode  string
	Verdict        string
	FailureType    string
	RepairReason   string
	TokenInput     int64
	TokenOutput    int64
	LLMLatencyMS   int
	TotalLatencyMS int
}

type AttemptRecorder interface {
	RecordAttempt(ctx context.Context, attempt AttemptRecord) error
}

type CoordinatorOutput struct {
	AgentName      string
	Model          string
	PromptPreview  string
	RawResponse    string
	ExtractedCode  string
	SubmissionID   uint
	Verdict        string
	ErrorMessage   string
	TokenInput     int64
	TokenOutput    int64
	LLMLatencyMS   int
	TotalLatencyMS int
	AttemptCount   int
	FailureType    string
	StrategyPath   string
	Attempts       []AttemptRecord
}

type AdaptiveRepairCoordinator struct {
	MaxAttempts int
	Classifier  FailureClassifier
	Planner     RepairPlanner
}

func (c AdaptiveRepairCoordinator) Execute(
	ctx context.Context,
	client llm.Client,
	input SolveInput,
	submitter JudgeSubmitter,
	recorder AttemptRecorder,
) (CoordinatorOutput, error) {
	startedAt := time.Now()
	maxAttempts := c.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	planner := c.Planner
	if planner.MaxAttempts <= 0 {
		planner.MaxAttempts = maxAttempts
	}

	var output CoordinatorOutput
	output.AgentName = AdaptiveRepairV1AgentName
	output.Model = input.Model

	stage := StageInitialCodegen
	stages := make([]string, 0, maxAttempts)
	var previousCode string
	var feedback string
	var lastFailureType string

	for attemptNo := 1; attemptNo <= maxAttempts; attemptNo++ {
		promptText := prompt.BuildSolvePrompt(input.Problem, input.PromptName)
		if attemptNo > 1 {
			promptText = prompt.BuildRepairPromptForStage(input.Problem, stage, previousCode, feedback)
		}

		llmOutput, err := ExecutePrompt(ctx, client, input.Model, promptText)
		llmOutput.AgentName = AdaptiveRepairV1AgentName
		output.TokenInput += llmOutput.TokenInput
		output.TokenOutput += llmOutput.TokenOutput
		output.LLMLatencyMS += llmOutput.LLMLatencyMS
		output.Model = effectiveModel(llmOutput.Model, output.Model, input.Model)
		output.PromptPreview = llmOutput.PromptPreview
		output.RawResponse = llmOutput.RawResponse
		if err != nil {
			output.TotalLatencyMS = elapsedMS(startedAt)
			return output, err
		}

		code := ExtractCPPCode(llmOutput.RawResponse)
		output.ExtractedCode = code
		if strings.TrimSpace(code) == "" {
			output.TotalLatencyMS = elapsedMS(startedAt)
			return output, ErrAdaptiveCodeNotExtracted
		}

		judgeResult, err := submitter.Submit(ctx, code)
		if err != nil {
			output.TotalLatencyMS = elapsedMS(startedAt)
			return output, err
		}
		if judgeResult == nil {
			judgeResult = &JudgeResult{}
		}

		classification := c.Classifier.Classify(judgeResult.Verdict)
		failureType := ""
		if classification.Repairable {
			failureType = classification.FailureType
			lastFailureType = failureType
		}

		stages = append(stages, stage)
		attempt := AttemptRecord{
			AttemptNo:      attemptNo,
			Stage:          stage,
			Model:          output.Model,
			PromptPreview:  llmOutput.PromptPreview,
			RawResponse:    llmOutput.RawResponse,
			ExtractedCode:  code,
			Verdict:        judgeResult.Verdict,
			FailureType:    failureType,
			RepairReason:   judgeResult.Feedback,
			TokenInput:     llmOutput.TokenInput,
			TokenOutput:    llmOutput.TokenOutput,
			LLMLatencyMS:   llmOutput.LLMLatencyMS,
			TotalLatencyMS: elapsedMS(startedAt),
		}
		if recorder != nil {
			if err := recorder.RecordAttempt(ctx, attempt); err != nil {
				output.TotalLatencyMS = elapsedMS(startedAt)
				return output, err
			}
		}
		output.Attempts = append(output.Attempts, attempt)
		output.AttemptCount = len(output.Attempts)
		output.SubmissionID = judgeResult.SubmissionID
		output.Verdict = judgeResult.Verdict
		output.ErrorMessage = judgeResult.ErrorMessage
		output.FailureType = lastFailureType
		output.StrategyPath = strings.Join(stages, " -> ")
		output.TotalLatencyMS = elapsedMS(startedAt)

		if judgeResult.Verdict == "AC" {
			return output, nil
		}

		plan, ok := planner.NextRepair(attemptNo, classification)
		if !ok {
			return output, nil
		}
		stage = plan.Stage
		previousCode = code
		feedback = effectiveModel(judgeResult.Feedback, judgeResult.ErrorMessage, judgeResult.Verdict)
	}

	return output, nil
}
