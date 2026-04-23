package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
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
const maxLegacyRepairAttempts = 3

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
	ID               uint   `json:"id"`
	AttemptNo        int    `json:"attempt_no"`
	Stage            string `json:"stage"`
	Model            string `json:"model"`
	Verdict          string `json:"verdict"`
	FailureType      string `json:"failure_type"`
	RepairReason     string `json:"repair_reason"`
	StrategyPath     string `json:"strategy_path"`
	PromptPreview    string `json:"prompt_preview"`
	ExtractedCode    string `json:"extracted_code"`
	JudgePassedCount int    `json:"judge_passed_count"`
	JudgeTotalCount  int    `json:"judge_total_count"`
	TimedOut         bool   `json:"timed_out"`
	ErrorMessage     string `json:"error_message"`
	TokenInput       int64  `json:"token_input"`
	TokenOutput      int64  `json:"token_output"`
	LLMLatencyMS     int    `json:"llm_latency_ms"`
	TotalLatencyMS   int    `json:"total_latency_ms"`
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
	var attemptRepo repository.AISolveAttemptRepository
	if len(attempts) > 0 {
		attemptRepo = attempts[0]
	}
	return &AISolveService{
		problems:     problems,
		runs:         runs,
		attempts:     attemptRepo,
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
		return s.solveAdaptiveRepair(solveCtx, run, output, startedAt, problem, resolvedModel, resolvedPromptName, input.ProblemID)
	}

	return s.solveLegacyAgent(solveCtx, run, output, startedAt, problem, resolvedModel, resolvedPromptName, resolvedAgentName, input.ProblemID, toolingConfig)
}

func (s *AISolveService) GetRun(ctx context.Context, runID uint) (*model.AISolveRun, error) {
	run, err := s.runs.GetByID(ctx, runID)
	if err != nil {
		return nil, err
	}
	return run, nil
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
	run.SubmissionID = nil
	run.Verdict = ""
	run.TotalLatencyMS = elapsedMS(startedAt)
	if err := s.persistTerminalRun(run); err != nil {
		syncAISolveOutputFromRun(output, run)
		output.SubmissionID = 0
		output.Verdict = ""
		return output, fmt.Errorf("update ai solve run: %w", err)
	}
	syncAISolveOutputFromRun(output, run)
	output.SubmissionID = 0
	output.Verdict = ""
	return output, returnErr
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

func (s *AISolveService) submitAttempt(
	ctx context.Context,
	problemID uint,
	code string,
	run *model.AISolveRun,
	output *AISolveOutput,
	startedAt time.Time,
) (*JudgeSubmissionOutput, error) {
	if strings.TrimSpace(code) == "" {
		return nil, s.wrapFailure(ctx, run, output, startedAt, ErrAISolveCodeNotExtracted.Error(), ErrAISolveCodeNotExtracted)
	}

	run.ExtractedCode = code
	output.ExtractedCode = code

	judgeOutput, err := s.submissions.Submit(ctx, JudgeSubmissionInput{
		ProblemID:  problemID,
		SourceCode: code,
		Language:   model.LanguageCPP17,
		SourceType: model.SourceTypeAI,
	})
	if err != nil {
		return nil, s.wrapFailure(ctx, run, output, startedAt, err.Error(), err)
	}

	return judgeOutput, nil
}

func (s *AISolveService) solveLegacyAgent(
	ctx context.Context,
	run *model.AISolveRun,
	output *AISolveOutput,
	startedAt time.Time,
	problem *model.Problem,
	resolvedModel string,
	resolvedPromptName string,
	resolvedAgentName string,
	problemID uint,
	toolingConfig tooling.Config,
) (*AISolveOutput, error) {
	strategy, err := agent.ResolveSolveStrategy(resolvedAgentName)
	if err != nil {
		return nil, err
	}

	agentOutput, err := strategy.Execute(ctx, s.llmClient, agent.SolveInput{
		Problem:       problem,
		Model:         resolvedModel,
		PromptName:    resolvedPromptName,
		ToolingRunner: s.newToolingRunner(toolingConfig),
	})
	s.applyAttemptLLMOutput(run, output, resolvedModel, resolvedAgentName, agentOutput)
	if err != nil {
		return s.failRun(ctx, run, output, startedAt, fmt.Sprintf("%v: %v", ErrAISolveLLMFailed, err), fmt.Errorf("%w: %v", ErrAISolveLLMFailed, err))
	}

	code := extractCPPCode(agentOutput.RawResponse)
	lastJudgeOutput, err := s.submitAttempt(ctx, problemID, code, run, output, startedAt)
	if err != nil {
		return output, err
	}
	if lastJudgeOutput.Verdict == "AC" {
		return s.finishRun(run, output, startedAt, lastJudgeOutput)
	}
	if resolvedAgentName != agent.DirectCodegenRepairAgentName {
		return s.finishRun(run, output, startedAt, lastJudgeOutput)
	}

	for attempt := 2; attempt <= maxLegacyRepairAttempts; attempt++ {
		repairPrompt := prompt.BuildRepairPrompt(problem, resolvedPromptName, code, buildRepairFeedback(lastJudgeOutput))
		llmStartedAt := time.Now()
		llmResp, llmErr := s.llmClient.Generate(ctx, llm.GenerateRequest{
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
			return s.failRun(ctx, run, output, startedAt, fmt.Sprintf("%v: %v", ErrAISolveLLMFailed, llmErr), fmt.Errorf("%w: %v", ErrAISolveLLMFailed, llmErr))
		}

		code = extractCPPCode(llmResp.Content)
		lastJudgeOutput, err = s.submitAttempt(ctx, problemID, code, run, output, startedAt)
		if err != nil {
			return output, err
		}
		if lastJudgeOutput.Verdict == "AC" {
			return s.finishRun(run, output, startedAt, lastJudgeOutput)
		}
	}

	return s.finishRun(run, output, startedAt, lastJudgeOutput)
}

func (s *AISolveService) solveAdaptiveRepair(
	ctx context.Context,
	run *model.AISolveRun,
	output *AISolveOutput,
	startedAt time.Time,
	problem *model.Problem,
	resolvedModel string,
	resolvedPromptName string,
	problemID uint,
) (*AISolveOutput, error) {
	adapter := &adaptiveJudgeSubmitterAdapter{
		submitter: s.submissions,
		problemID: problemID,
	}

	result, err := agent.NewAdaptiveRepairCoordinator(agent.DefaultAdaptiveRepairMaxAttempts).Execute(ctx, s.llmClient, agent.SolveInput{
		Problem:        problem,
		Model:          resolvedModel,
		PromptName:     resolvedPromptName,
		JudgeSubmitter: adapter,
	})

	var judgeOutput *JudgeSubmissionOutput
	if err == nil {
		judgeOutput = adapter.lastOutput()
	}
	s.applyAdaptiveResult(run, output, result, judgeOutput, startedAt)
	if persistErr := s.persistAdaptiveAttempts(ctx, run.ID, result.StrategyPath, result.Attempts); persistErr != nil {
		return s.failRun(ctx, run, output, startedAt, persistErr.Error(), persistErr)
	}
	if err != nil {
		return s.failRun(ctx, run, output, startedAt, err.Error(), err)
	}

	return s.finishRun(run, output, startedAt, adapter.lastOutput())
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

func (s *AISolveService) applyAdaptiveResult(
	run *model.AISolveRun,
	output *AISolveOutput,
	result agent.AdaptiveRepairResult,
	judgeOutput *JudgeSubmissionOutput,
	startedAt time.Time,
) {
	run.Model = result.SolveOutput.Model
	run.AgentName = agent.DisplaySolveAgentName(result.SolveOutput.AgentName)
	run.PromptPreview = truncateForPreview(result.SolveOutput.PromptPreview, 800)
	run.RawResponse = result.SolveOutput.RawResponse
	run.ExtractedCode = extractCPPCode(result.SolveOutput.RawResponse)
	run.AttemptCount = result.AttemptCount
	run.FailureType = adaptiveRunFailureType(result, judgeOutput)
	run.StrategyPath = strings.Join(result.StrategyPath, ",")
	run.TokenInput = result.SolveOutput.TokenInput
	run.TokenOutput = result.SolveOutput.TokenOutput
	run.LLMLatencyMS = result.SolveOutput.LLMLatencyMS
	run.TotalLatencyMS = elapsedMS(startedAt)
	run.SubmissionID = nil
	run.Verdict = ""
	run.ErrorMessage = ""
	if judgeOutput != nil {
		run.SubmissionID = &judgeOutput.SubmissionID
		run.Verdict = judgeOutput.Verdict
		run.ErrorMessage = judgeOutput.ErrorMessage
	}

	syncAISolveOutputFromRun(output, run)
	output.RawResponse = truncateForPreview(result.SolveOutput.RawResponse, 4000)
	output.ExtractedCode = run.ExtractedCode
	output.SubmissionID = 0
	output.Verdict = ""
	output.ErrorMessage = ""
	if judgeOutput != nil {
		output.SubmissionID = judgeOutput.SubmissionID
		output.Verdict = judgeOutput.Verdict
		output.ErrorMessage = judgeOutput.ErrorMessage
	}
	output.Attempts = adaptiveAttemptsToOutput(result.StrategyPath, result.Attempts)
}

func adaptiveRunFailureType(result agent.AdaptiveRepairResult, judgeOutput *JudgeSubmissionOutput) string {
	if judgeOutput != nil && strings.EqualFold(strings.TrimSpace(judgeOutput.Verdict), "AC") {
		if failure := lastMeaningfulAdaptiveFailure(result.Attempts); failure != "" {
			return failure
		}
	}

	if failure := strings.TrimSpace(string(result.FinalFailure)); failure != "" {
		return failure
	}
	return string(agent.FailureTypeUnknown)
}

func lastMeaningfulAdaptiveFailure(attempts []agent.AdaptiveRepairAttempt) string {
	for i := len(attempts) - 1; i >= 0; i-- {
		failure := strings.TrimSpace(string(attempts[i].FailureType))
		if failure == "" || failure == string(agent.FailureTypeUnknown) {
			continue
		}
		return failure
	}
	return ""
}

func (s *AISolveService) persistAdaptiveAttempts(
	_ context.Context,
	runID uint,
	strategyPath []string,
	attempts []agent.AdaptiveRepairAttempt,
) error {
	if s.attempts == nil {
		return nil
	}

	updateCtx, cancel := terminalPersistenceContext()
	defer cancel()

	for _, attempt := range attempts {
		record := &model.AISolveAttempt{
			AISolveRunID:     runID,
			AttemptNo:        attempt.AttemptNo,
			Stage:            attempt.Stage,
			Model:            attempt.Model,
			FailureType:      string(attempt.FailureType),
			StrategyPath:     strategyPathForAttempt(strategyPath, attempt.AttemptNo),
			RepairReason:     repairReasonFromJudgeFeedback(attempt.JudgeFeedback),
			PromptPreview:    truncateForPreview(attempt.PromptPreview, 800),
			RawResponse:      attempt.RawResponse,
			ExtractedCode:    extractCPPCode(attempt.RawResponse),
			JudgeVerdict:     attempt.JudgeFeedback.Verdict,
			JudgePassedCount: attempt.JudgeFeedback.PassedCount,
			JudgeTotalCount:  attempt.JudgeFeedback.TotalCount,
			TimedOut:         attempt.JudgeFeedback.TimedOut,
			CompileStderr:    attempt.JudgeFeedback.CompileStderr,
			RunStderr:        attempt.JudgeFeedback.RunStderr,
			RunStdout:        attempt.JudgeFeedback.RunStdout,
			ErrorMessage:     attempt.JudgeFeedback.ErrorMessage,
			TokenInput:       attempt.TokenInput,
			TokenOutput:      attempt.TokenOutput,
			LLMLatencyMS:     attempt.LLMLatencyMS,
			TotalLatencyMS:   attempt.LLMLatencyMS,
		}
		if err := s.attempts.Create(updateCtx, record); err != nil {
			return err
		}
	}

	return nil
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

type adaptiveJudgeSubmitterAdapter struct {
	submitter JudgeSubmitter
	problemID uint
	lastJudge *JudgeSubmissionOutput
}

func (a *adaptiveJudgeSubmitterAdapter) Submit(ctx context.Context, sourceCode string) (*agent.JudgeFeedback, error) {
	if a == nil || a.submitter == nil {
		return nil, errors.New("judge submitter is required")
	}
	sourceCode = strings.TrimSpace(sourceCode)
	if sourceCode == "" {
		return nil, ErrAISolveCodeNotExtracted
	}

	output, err := a.submitter.Submit(ctx, JudgeSubmissionInput{
		ProblemID:  a.problemID,
		SourceCode: sourceCode,
		Language:   model.LanguageCPP17,
		SourceType: model.SourceTypeAI,
	})
	if err != nil {
		return nil, err
	}

	a.lastJudge = output
	return judgeFeedbackFromSubmission(output), nil
}

func (a *adaptiveJudgeSubmitterAdapter) lastOutput() *JudgeSubmissionOutput {
	if a == nil || a.lastJudge == nil {
		return nil
	}
	copy := *a.lastJudge
	return &copy
}

func judgeFeedbackFromSubmission(output *JudgeSubmissionOutput) *agent.JudgeFeedback {
	if output == nil {
		return nil
	}
	return &agent.JudgeFeedback{
		Verdict:       output.Verdict,
		TimedOut:      output.TimedOut,
		CompileStderr: output.CompileStderr,
		RunStdout:     output.RunStdout,
		RunStderr:     output.RunStderr,
		ErrorMessage:  output.ErrorMessage,
		PassedCount:   output.PassedCount,
		TotalCount:    output.TotalCount,
		ExecStage:     output.ExecStage,
	}
}

func strategyPathForAttempt(strategyPath []string, attemptNo int) string {
	if attemptNo <= 1 {
		return ""
	}

	limit := attemptNo - 1
	if limit > len(strategyPath) {
		limit = len(strategyPath)
	}
	if limit <= 0 {
		return ""
	}

	stages := make([]string, 0, limit)
	for i := 0; i < limit && i < len(strategyPath); i++ {
		stage := strings.TrimSpace(strategyPath[i])
		if stage == "" {
			continue
		}
		stages = append(stages, stage)
	}
	return strings.Join(stages, ",")
}

func repairReasonFromJudgeFeedback(feedback agent.JudgeFeedback) string {
	return firstNonEmpty(feedback.ErrorMessage, feedback.CompileStderr, feedback.RunStderr, feedback.RunStdout, feedback.Verdict)
}

func adaptiveAttemptsToOutput(strategyPath []string, attempts []agent.AdaptiveRepairAttempt) []AISolveAttemptOutput {
	if len(attempts) == 0 {
		return nil
	}
	outputs := make([]AISolveAttemptOutput, 0, len(attempts))
	for _, attempt := range attempts {
		outputs = append(outputs, AISolveAttemptOutput{
			AttemptNo:        attempt.AttemptNo,
			Stage:            attempt.Stage,
			Model:            attempt.Model,
			Verdict:          attempt.JudgeFeedback.Verdict,
			FailureType:      string(attempt.FailureType),
			RepairReason:     repairReasonFromJudgeFeedback(attempt.JudgeFeedback),
			StrategyPath:     strategyPathForAttempt(strategyPath, attempt.AttemptNo),
			PromptPreview:    truncateForPreview(attempt.PromptPreview, 800),
			ExtractedCode:    extractCPPCode(attempt.RawResponse),
			JudgePassedCount: attempt.JudgeFeedback.PassedCount,
			JudgeTotalCount:  attempt.JudgeFeedback.TotalCount,
			TimedOut:         attempt.JudgeFeedback.TimedOut,
			ErrorMessage:     attempt.JudgeFeedback.ErrorMessage,
			TokenInput:       attempt.TokenInput,
			TokenOutput:      attempt.TokenOutput,
			LLMLatencyMS:     attempt.LLMLatencyMS,
			TotalLatencyMS:   attempt.LLMLatencyMS,
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
			ID:               attempt.ID,
			AttemptNo:        attempt.AttemptNo,
			Stage:            attempt.Stage,
			Model:            attempt.Model,
			Verdict:          firstNonEmpty(attempt.JudgeVerdict, attempt.FailureType),
			FailureType:      attempt.FailureType,
			RepairReason:     attempt.RepairReason,
			StrategyPath:     attempt.StrategyPath,
			PromptPreview:    attempt.PromptPreview,
			ExtractedCode:    attempt.ExtractedCode,
			JudgePassedCount: attempt.JudgePassedCount,
			JudgeTotalCount:  attempt.JudgeTotalCount,
			TimedOut:         attempt.TimedOut,
			ErrorMessage:     attempt.ErrorMessage,
			TokenInput:       attempt.TokenInput,
			TokenOutput:      attempt.TokenOutput,
			LLMLatencyMS:     attempt.LLMLatencyMS,
			TotalLatencyMS:   attempt.TotalLatencyMS,
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

	updateCtx, cancel := terminalPersistenceContext()
	defer cancel()

	return s.runs.Update(updateCtx, run)
}

func terminalPersistenceContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
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
