package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"ai-for-oj/internal/agent"
	"ai-for-oj/internal/llm"
	"ai-for-oj/internal/model"
	"ai-for-oj/internal/prompt"
	"ai-for-oj/internal/repository"
	"ai-for-oj/internal/tooling"
)

var (
	ErrAISolveLLMFailed        = errors.New("llm solve failed")
	ErrAISolveCodeNotExtracted = errors.New("failed to extract cpp17 code from llm response")
)

const defaultAISolveExecutionTimeout = 5 * time.Minute
const maxAISolveAttempts = 3

type JudgeSubmitter interface {
	Submit(ctx context.Context, input JudgeSubmissionInput) (*JudgeSubmissionOutput, error)
}

type AISolveInput struct {
	ProblemID     uint
	Model         string
	PromptName    string
	AgentName     string
	ToolingConfig string
}

type AISolveOutput struct {
	AISolveRunID   uint                   `json:"ai_solve_run_id"`
	ProblemID      uint                   `json:"problem_id"`
	Model          string                 `json:"model,omitempty"`
	PromptName     string                 `json:"prompt_name"`
	AgentName      string                 `json:"agent_name"`
	PromptPreview  string                 `json:"prompt_preview"`
	RawResponse    string                 `json:"raw_response,omitempty"`
	ExtractedCode  string                 `json:"extracted_code,omitempty"`
	SubmissionID   uint                   `json:"submission_id"`
	Verdict        string                 `json:"verdict,omitempty"`
	ErrorMessage   string                 `json:"error_message,omitempty"`
	AttemptCount   int                    `json:"attempt_count"`
	FailureType    string                 `json:"failure_type,omitempty"`
	StrategyPath   string                 `json:"strategy_path,omitempty"`
	ToolingConfig  string                 `json:"tooling_config"`
	ToolCallCount  int                    `json:"tool_call_count"`
	TokenInput     int64                  `json:"token_input"`
	TokenOutput    int64                  `json:"token_output"`
	LLMLatencyMS   int                    `json:"llm_latency_ms"`
	TotalLatencyMS int                    `json:"total_latency_ms"`
	Attempts       []AISolveAttemptOutput `json:"attempts,omitempty"`
}

type AISolveAttemptOutput struct {
	ID             uint   `json:"id"`
	AttemptNo      int    `json:"attempt_no"`
	Stage          string `json:"stage"`
	Model          string `json:"model"`
	Verdict        string `json:"verdict"`
	FailureType    string `json:"failure_type"`
	RepairReason   string `json:"repair_reason"`
	TokenInput     int64  `json:"token_input"`
	TokenOutput    int64  `json:"token_output"`
	LLMLatencyMS   int    `json:"llm_latency_ms"`
	TotalLatencyMS int    `json:"total_latency_ms"`
}

type AISolveService struct {
	problems        repository.ProblemRepository
	runs            repository.AISolveRunRepository
	attempts        repository.AISolveAttemptRepository
	llmClient       llm.Client
	submissions     JudgeSubmitter
	defaultModel    string
	toolingRegistry *tooling.Registry
}

func NewAISolveService(
	problems repository.ProblemRepository,
	runs repository.AISolveRunRepository,
	llmClient llm.Client,
	submissions JudgeSubmitter,
	defaultModel string,
	attempts ...repository.AISolveAttemptRepository,
) *AISolveService {
	var attemptRepository repository.AISolveAttemptRepository
	if len(attempts) > 0 {
		attemptRepository = attempts[0]
	}
	return &AISolveService{
		problems:     problems,
		runs:         runs,
		attempts:     attemptRepository,
		llmClient:    llmClient,
		submissions:  submissions,
		defaultModel: defaultModel,
	}
}

func (s *AISolveService) SetToolingRegistry(registry *tooling.Registry) {
	s.toolingRegistry = registry
}

func (s *AISolveService) Solve(ctx context.Context, input AISolveInput) (*AISolveOutput, error) {
	solveCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), defaultAISolveExecutionTimeout)
	defer cancel()

	startedAt := time.Now()
	resolvedModel := firstNonEmpty(input.Model, s.defaultModel)
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
	toolingConfig, _, err := tooling.ResolveConfig(canonicalToolingConfig)
	if err != nil {
		return nil, err
	}
	run := &model.AISolveRun{
		ProblemID:     input.ProblemID,
		Model:         resolvedModel,
		PromptName:    resolvedPromptName,
		AgentName:     resolvedAgentName,
		ToolingConfig: canonicalToolingConfig,
		Status:        model.AISolveRunStatusRunning,
	}
	if err := s.runs.Create(solveCtx, run); err != nil {
		return nil, fmt.Errorf("create ai solve run: %w", err)
	}

	output := &AISolveOutput{
		AISolveRunID:  run.ID,
		ProblemID:     input.ProblemID,
		Model:         resolvedModel,
		PromptName:    resolvedPromptName,
		AgentName:     resolvedAgentName,
		ToolingConfig: canonicalToolingConfig,
	}

	problem, err := s.problems.GetByID(solveCtx, input.ProblemID)
	if err != nil {
		return s.failRun(solveCtx, run, output, startedAt, err.Error(), err)
	}

	if resolvedAgentName == agent.AdaptiveRepairV1AgentName {
		return s.solveAdaptiveRepair(solveCtx, input.ProblemID, problem, run, output, startedAt, resolvedModel, resolvedPromptName)
	}

	strategy, err := agent.ResolveSolveStrategy(resolvedAgentName)
	if err != nil {
		return nil, err
	}

	agentOutput, err := strategy.Execute(solveCtx, s.llmClient, agent.SolveInput{
		Problem:       problem,
		Model:         resolvedModel,
		PromptName:    resolvedPromptName,
		ToolingRunner: s.newToolingRunner(toolingConfig),
	})
	s.applyAttemptLLMOutput(run, output, resolvedModel, resolvedAgentName, agentOutput)
	if err != nil {
		return s.failRun(solveCtx, run, output, startedAt, fmt.Sprintf("%v: %v", ErrAISolveLLMFailed, err), fmt.Errorf("%w: %v", ErrAISolveLLMFailed, err))
	}

	code := agent.ExtractCPPCode(agentOutput.RawResponse)
	lastJudgeOutput, err := s.submitAttempt(solveCtx, input.ProblemID, code, run, output, startedAt)
	if err != nil {
		return output, err
	}
	if lastJudgeOutput.Verdict == "AC" {
		return s.finishRun(run, output, startedAt, lastJudgeOutput)
	}
	if !agent.SupportsSelfRepair(resolvedAgentName) {
		return s.finishRun(run, output, startedAt, lastJudgeOutput)
	}

	for attempt := 2; attempt <= maxAISolveAttempts; attempt++ {
		repairPrompt := prompt.BuildRepairPrompt(problem, resolvedPromptName, code, buildRepairFeedback(lastJudgeOutput))
		llmStartedAt := time.Now()
		llmResp, llmErr := s.llmClient.Generate(solveCtx, llm.GenerateRequest{
			Prompt: repairPrompt,
			Model:  resolvedModel,
		})
		repairOutput := agent.SolveOutput{
			AgentName:     resolvedAgentName,
			Model:         firstNonEmpty(llmResp.Model, resolvedModel),
			PromptPreview: repairPrompt,
			RawResponse:   llmResp.Content,
			TokenInput:    llmResp.InputTokens,
			TokenOutput:   llmResp.OutputTokens,
			LLMLatencyMS:  elapsedMS(llmStartedAt),
		}
		s.applyAttemptLLMOutput(run, output, resolvedModel, resolvedAgentName, repairOutput)
		if llmErr != nil {
			return s.failRun(solveCtx, run, output, startedAt, fmt.Sprintf("%v: %v", ErrAISolveLLMFailed, llmErr), fmt.Errorf("%w: %v", ErrAISolveLLMFailed, llmErr))
		}

		code = agent.ExtractCPPCode(llmResp.Content)
		lastJudgeOutput, err = s.submitAttempt(solveCtx, input.ProblemID, code, run, output, startedAt)
		if err != nil {
			return output, err
		}
		if lastJudgeOutput.Verdict == "AC" {
			return s.finishRun(run, output, startedAt, lastJudgeOutput)
		}
	}

	return s.finishRun(run, output, startedAt, lastJudgeOutput)
}

func (s *AISolveService) GetRun(ctx context.Context, runID uint) (*model.AISolveRun, error) {
	return s.runs.GetByID(ctx, runID)
}

func (s *AISolveService) solveAdaptiveRepair(
	ctx context.Context,
	problemID uint,
	problem *model.Problem,
	run *model.AISolveRun,
	output *AISolveOutput,
	startedAt time.Time,
	resolvedModel string,
	resolvedPromptName string,
) (*AISolveOutput, error) {
	coordinator := agent.AdaptiveRepairCoordinator{MaxAttempts: maxAISolveAttempts}
	coordinatorOutput, err := coordinator.Execute(
		ctx,
		s.llmClient,
		agent.SolveInput{
			Problem:    problem,
			Model:      resolvedModel,
			PromptName: resolvedPromptName,
		},
		adaptiveAgentJudgeSubmitter{
			submitter: s.submissions,
			input: JudgeSubmissionInput{
				ProblemID:  problemID,
				Language:   model.LanguageCPP17,
				SourceType: model.SourceTypeAI,
			},
		},
		aiSolveAttemptRecorder{repo: s.attempts, runID: run.ID},
	)
	s.applyAdaptiveCoordinatorOutput(run, output, coordinatorOutput)
	if err != nil {
		switch {
		case errors.Is(err, agent.ErrAdaptiveCodeNotExtracted):
			return s.failRun(ctx, run, output, startedAt, ErrAISolveCodeNotExtracted.Error(), ErrAISolveCodeNotExtracted)
		case len(coordinatorOutput.Attempts) == 0:
			return s.failRun(ctx, run, output, startedAt, fmt.Sprintf("%v: %v", ErrAISolveLLMFailed, err), fmt.Errorf("%w: %v", ErrAISolveLLMFailed, err))
		default:
			return s.failRun(ctx, run, output, startedAt, err.Error(), err)
		}
	}

	run.Status = model.AISolveRunStatusSuccess
	run.TotalLatencyMS = elapsedMS(startedAt)
	if run.TotalLatencyMS < coordinatorOutput.TotalLatencyMS {
		run.TotalLatencyMS = coordinatorOutput.TotalLatencyMS
	}
	output.TotalLatencyMS = run.TotalLatencyMS
	if err := s.persistTerminalRun(run); err != nil {
		syncAISolveOutputFromRun(output, run)
		return output, fmt.Errorf("update ai solve run: %w", err)
	}
	syncAISolveOutputFromRun(output, run)
	output.SubmissionID = coordinatorOutput.SubmissionID
	output.Verdict = coordinatorOutput.Verdict
	output.ErrorMessage = coordinatorOutput.ErrorMessage
	output.Attempts = coordinatorAttemptsToOutput(coordinatorOutput.Attempts)
	return output, nil
}

func (s *AISolveService) failRun(
	ctx context.Context,
	run *model.AISolveRun,
	output *AISolveOutput,
	startedAt time.Time,
	message string,
	returnErr error,
) (*AISolveOutput, error) {
	run.Status = model.AISolveRunStatusFailed
	run.ErrorMessage = message
	run.TotalLatencyMS = elapsedMS(startedAt)
	if err := s.persistTerminalRun(run); err != nil {
		syncAISolveOutputFromRun(output, run)
		return output, fmt.Errorf("update ai solve run: %w", err)
	}
	syncAISolveOutputFromRun(output, run)
	return output, returnErr
}

func truncateForPreview(value string, limit int) string {
	value = strings.TrimSpace(value)
	if limit <= 0 {
		return value
	}
	if utf8.RuneCountInString(value) <= limit {
		return value
	}
	return string([]rune(value)[:limit]) + "...(truncated)"
}

func (s *AISolveService) applyAttemptLLMOutput(
	run *model.AISolveRun,
	output *AISolveOutput,
	resolvedModel string,
	resolvedAgentName string,
	attempt agent.SolveOutput,
) {
	run.PromptPreview = truncateForPreview(attempt.PromptPreview, 800)
	run.Model = firstNonEmpty(attempt.Model, resolvedModel)
	run.AgentName = agent.DisplaySolveAgentName(firstNonEmpty(attempt.AgentName, resolvedAgentName))
	run.TokenInput += attempt.TokenInput
	run.TokenOutput += attempt.TokenOutput
	run.LLMLatencyMS += attempt.LLMLatencyMS
	run.RawResponse = attempt.RawResponse

	output.Model = run.Model
	output.AgentName = run.AgentName
	output.PromptPreview = run.PromptPreview
	output.RawResponse = truncateForPreview(attempt.RawResponse, 4000)
	output.TokenInput = run.TokenInput
	output.TokenOutput = run.TokenOutput
	output.LLMLatencyMS = run.LLMLatencyMS
	run.ToolCallCount += attempt.ToolCallCount
	output.ToolCallCount = run.ToolCallCount
}

func (s *AISolveService) newToolingRunner(cfg tooling.Config) *tooling.Runner {
	if s.toolingRegistry == nil {
		return nil
	}
	return s.toolingRegistry.NewRunner(cfg)
}

func (s *AISolveService) applyAdaptiveCoordinatorOutput(
	run *model.AISolveRun,
	output *AISolveOutput,
	coordinatorOutput agent.CoordinatorOutput,
) {
	run.Model = firstNonEmpty(coordinatorOutput.Model, run.Model)
	run.AgentName = agent.AdaptiveRepairV1AgentName
	run.PromptPreview = truncateForPreview(coordinatorOutput.PromptPreview, 800)
	run.RawResponse = coordinatorOutput.RawResponse
	run.ExtractedCode = coordinatorOutput.ExtractedCode
	run.Verdict = coordinatorOutput.Verdict
	run.ErrorMessage = coordinatorOutput.ErrorMessage
	run.TokenInput = coordinatorOutput.TokenInput
	run.TokenOutput = coordinatorOutput.TokenOutput
	run.LLMLatencyMS = coordinatorOutput.LLMLatencyMS
	run.TotalLatencyMS = coordinatorOutput.TotalLatencyMS
	run.AttemptCount = coordinatorOutput.AttemptCount
	run.FailureType = coordinatorOutput.FailureType
	run.StrategyPath = coordinatorOutput.StrategyPath
	if coordinatorOutput.SubmissionID != 0 {
		run.SubmissionID = &coordinatorOutput.SubmissionID
	}

	output.Model = run.Model
	output.AgentName = run.AgentName
	output.PromptPreview = run.PromptPreview
	output.RawResponse = truncateForPreview(coordinatorOutput.RawResponse, 4000)
	output.ExtractedCode = coordinatorOutput.ExtractedCode
	output.SubmissionID = coordinatorOutput.SubmissionID
	output.Verdict = coordinatorOutput.Verdict
	output.ErrorMessage = coordinatorOutput.ErrorMessage
	output.TokenInput = coordinatorOutput.TokenInput
	output.TokenOutput = coordinatorOutput.TokenOutput
	output.LLMLatencyMS = coordinatorOutput.LLMLatencyMS
	output.TotalLatencyMS = coordinatorOutput.TotalLatencyMS
	output.AttemptCount = coordinatorOutput.AttemptCount
	output.FailureType = coordinatorOutput.FailureType
	output.StrategyPath = coordinatorOutput.StrategyPath
	output.ToolingConfig = run.ToolingConfig
	output.ToolCallCount = run.ToolCallCount
	output.Attempts = coordinatorAttemptsToOutput(coordinatorOutput.Attempts)
}

func (s *AISolveService) submitAttempt(
	ctx context.Context,
	problemID uint,
	code string,
	run *model.AISolveRun,
	output *AISolveOutput,
	startedAt time.Time,
) (*JudgeSubmissionOutput, error) {
	run.ExtractedCode = code
	output.ExtractedCode = code

	judgeOutput, err := adaptiveJudgeSubmitterAdapter{
		submitter: s.submissions,
		input: JudgeSubmissionInput{
			ProblemID:  problemID,
			Language:   model.LanguageCPP17,
			SourceType: model.SourceTypeAI,
		},
	}.Submit(ctx, code)
	if err != nil {
		return nil, s.wrapFailure(ctx, run, output, startedAt, err.Error(), err)
	}

	return judgeOutput, nil
}

type adaptiveJudgeSubmitterAdapter struct {
	submitter JudgeSubmitter
	input     JudgeSubmissionInput
}

type adaptiveAgentJudgeSubmitter struct {
	submitter JudgeSubmitter
	input     JudgeSubmissionInput
}

func (a adaptiveAgentJudgeSubmitter) Submit(ctx context.Context, sourceCode string) (*agent.JudgeResult, error) {
	if strings.TrimSpace(sourceCode) == "" {
		return nil, ErrAISolveCodeNotExtracted
	}

	input := a.input
	input.SourceCode = sourceCode
	output, err := a.submitter.Submit(ctx, input)
	if err != nil {
		return nil, err
	}
	if output == nil {
		return &agent.JudgeResult{}, nil
	}
	return &agent.JudgeResult{
		SubmissionID: output.SubmissionID,
		Verdict:      output.Verdict,
		ErrorMessage: output.ErrorMessage,
		Feedback:     buildRepairFeedback(output),
	}, nil
}

type aiSolveAttemptRecorder struct {
	repo  repository.AISolveAttemptRepository
	runID uint
}

func (r aiSolveAttemptRecorder) RecordAttempt(ctx context.Context, attempt agent.AttemptRecord) error {
	if r.repo == nil {
		return nil
	}
	return r.repo.Create(ctx, &model.AISolveAttempt{
		AISolveRunID:   r.runID,
		AttemptNo:      attempt.AttemptNo,
		Stage:          attempt.Stage,
		Model:          attempt.Model,
		PromptPreview:  truncateForPreview(attempt.PromptPreview, 800),
		RawResponse:    attempt.RawResponse,
		ExtractedCode:  attempt.ExtractedCode,
		Verdict:        attempt.Verdict,
		FailureType:    attempt.FailureType,
		RepairReason:   attempt.RepairReason,
		TokenInput:     attempt.TokenInput,
		TokenOutput:    attempt.TokenOutput,
		LLMLatencyMS:   attempt.LLMLatencyMS,
		TotalLatencyMS: attempt.TotalLatencyMS,
	})
}

func (a adaptiveJudgeSubmitterAdapter) Submit(ctx context.Context, sourceCode string) (*JudgeSubmissionOutput, error) {
	if strings.TrimSpace(sourceCode) == "" {
		return nil, ErrAISolveCodeNotExtracted
	}

	input := a.input
	input.SourceCode = sourceCode
	return a.submitter.Submit(ctx, input)
}

func (s *AISolveService) finishRun(
	run *model.AISolveRun,
	output *AISolveOutput,
	startedAt time.Time,
	judgeOutput *JudgeSubmissionOutput,
) (*AISolveOutput, error) {
	run.Status = model.AISolveRunStatusSuccess
	run.ErrorMessage = judgeOutput.ErrorMessage
	run.Verdict = judgeOutput.Verdict
	run.SubmissionID = &judgeOutput.SubmissionID
	run.TotalLatencyMS = elapsedMS(startedAt)
	if err := s.persistTerminalRun(run); err != nil {
		syncAISolveOutputFromRun(output, run)
		output.SubmissionID = judgeOutput.SubmissionID
		output.Verdict = judgeOutput.Verdict
		output.ErrorMessage = judgeOutput.ErrorMessage
		return output, fmt.Errorf("update ai solve run: %w", err)
	}

	syncAISolveOutputFromRun(output, run)
	output.SubmissionID = judgeOutput.SubmissionID
	output.Verdict = judgeOutput.Verdict
	output.ErrorMessage = judgeOutput.ErrorMessage
	return output, nil
}

func (s *AISolveService) wrapFailure(
	ctx context.Context,
	run *model.AISolveRun,
	output *AISolveOutput,
	startedAt time.Time,
	message string,
	returnErr error,
) error {
	_, err := s.failRun(ctx, run, output, startedAt, message, returnErr)
	if err != nil {
		return err
	}
	return returnErr
}

func buildRepairFeedback(output *JudgeSubmissionOutput) string {
	if output == nil {
		return "No judge feedback is available."
	}

	parts := []string{
		fmt.Sprintf("Verdict: %s", firstNonEmpty(output.Verdict, "unknown")),
		fmt.Sprintf("Passed Count: %d / %d", output.PassedCount, output.TotalCount),
	}
	if strings.TrimSpace(output.ErrorMessage) != "" {
		parts = append(parts, "Error Message:\n"+truncateForPreview(output.ErrorMessage, 1200))
	}
	if strings.TrimSpace(output.ExecStage) != "" {
		parts = append(parts, "Execution Stage: "+output.ExecStage)
	}
	if output.TimedOut {
		parts = append(parts, "Timed Out: true")
	}
	if strings.TrimSpace(output.CompileStderr) != "" {
		parts = append(parts, "Compile Stderr:\n"+truncateForPreview(output.CompileStderr, 1200))
	}
	if strings.TrimSpace(output.RunStderr) != "" {
		parts = append(parts, "Run Stderr:\n"+truncateForPreview(output.RunStderr, 1200))
	}
	if strings.TrimSpace(output.RunStdout) != "" {
		parts = append(parts, "Run Stdout:\n"+truncateForPreview(output.RunStdout, 800))
	}
	failedCases := make([]string, 0, 3)
	for _, item := range output.TestCaseResults {
		if item.Verdict == "AC" {
			continue
		}
		failedCases = append(failedCases, fmt.Sprintf(
			"case #%d verdict=%s runtime_ms=%d timed_out=%t exit_code=%d stderr=%q stdout=%q",
			item.CaseIndex,
			item.Verdict,
			item.RuntimeMS,
			item.TimedOut,
			item.ExitCode,
			truncateForPreview(item.Stderr, 200),
			truncateForPreview(item.Stdout, 200),
		))
		if len(failedCases) == 3 {
			break
		}
	}
	if len(failedCases) > 0 {
		parts = append(parts, "Failed Test Cases:\n"+strings.Join(failedCases, "\n"))
	}

	return strings.Join(parts, "\n\n")
}

func coordinatorAttemptsToOutput(attempts []agent.AttemptRecord) []AISolveAttemptOutput {
	if len(attempts) == 0 {
		return nil
	}
	outputs := make([]AISolveAttemptOutput, 0, len(attempts))
	for _, attempt := range attempts {
		outputs = append(outputs, AISolveAttemptOutput{
			AttemptNo:      attempt.AttemptNo,
			Stage:          attempt.Stage,
			Model:          attempt.Model,
			Verdict:        attempt.Verdict,
			FailureType:    attempt.FailureType,
			RepairReason:   attempt.RepairReason,
			TokenInput:     attempt.TokenInput,
			TokenOutput:    attempt.TokenOutput,
			LLMLatencyMS:   attempt.LLMLatencyMS,
			TotalLatencyMS: attempt.TotalLatencyMS,
		})
	}
	return outputs
}

func modelAttemptsToOutput(attempts []model.AISolveAttempt) []AISolveAttemptOutput {
	if len(attempts) == 0 {
		return nil
	}
	outputs := make([]AISolveAttemptOutput, 0, len(attempts))
	for _, attempt := range attempts {
		outputs = append(outputs, AISolveAttemptOutput{
			ID:             attempt.ID,
			AttemptNo:      attempt.AttemptNo,
			Stage:          attempt.Stage,
			Model:          attempt.Model,
			Verdict:        attempt.Verdict,
			FailureType:    attempt.FailureType,
			RepairReason:   attempt.RepairReason,
			TokenInput:     attempt.TokenInput,
			TokenOutput:    attempt.TokenOutput,
			LLMLatencyMS:   attempt.LLMLatencyMS,
			TotalLatencyMS: attempt.TotalLatencyMS,
		})
	}
	return outputs
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func elapsedMS(start time.Time) int {
	if start.IsZero() {
		return 0
	}
	return int(time.Since(start).Milliseconds())
}

func (s *AISolveService) persistTerminalRun(run *model.AISolveRun) error {
	if run == nil {
		return nil
	}

	updateCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.runs.Update(updateCtx, run)
}

func syncAISolveOutputFromRun(output *AISolveOutput, run *model.AISolveRun) {
	if output == nil || run == nil {
		return
	}
	output.Model = run.Model
	output.PromptName = prompt.DisplaySolvePromptName(run.PromptName)
	output.AgentName = agent.DisplaySolveAgentName(run.AgentName)
	output.PromptPreview = run.PromptPreview
	output.ErrorMessage = run.ErrorMessage
	output.AttemptCount = run.AttemptCount
	output.FailureType = run.FailureType
	output.StrategyPath = run.StrategyPath
	output.ToolingConfig = run.ToolingConfig
	output.ToolCallCount = run.ToolCallCount
	output.Attempts = modelAttemptsToOutput(run.Attempts)
	output.TokenInput = run.TokenInput
	output.TokenOutput = run.TokenOutput
	output.LLMLatencyMS = run.LLMLatencyMS
	output.TotalLatencyMS = run.TotalLatencyMS
}
