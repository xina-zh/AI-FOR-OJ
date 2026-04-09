package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"ai-for-oj/internal/llm"
	"ai-for-oj/internal/model"
	"ai-for-oj/internal/repository"
)

var (
	ErrAISolveLLMFailed        = errors.New("llm solve failed")
	ErrAISolveCodeNotExtracted = errors.New("failed to extract cpp17 code from llm response")
)

type JudgeSubmitter interface {
	Submit(ctx context.Context, input JudgeSubmissionInput) (*JudgeSubmissionOutput, error)
}

type AISolveInput struct {
	ProblemID uint
	Model     string
}

type AISolveOutput struct {
	AISolveRunID   uint   `json:"ai_solve_run_id"`
	ProblemID      uint   `json:"problem_id"`
	Model          string `json:"model,omitempty"`
	PromptPreview  string `json:"prompt_preview"`
	RawResponse    string `json:"raw_response,omitempty"`
	ExtractedCode  string `json:"extracted_code,omitempty"`
	SubmissionID   uint   `json:"submission_id"`
	Verdict        string `json:"verdict,omitempty"`
	ErrorMessage   string `json:"error_message,omitempty"`
	TokenInput     int64  `json:"token_input"`
	TokenOutput    int64  `json:"token_output"`
	LLMLatencyMS   int    `json:"llm_latency_ms"`
	TotalLatencyMS int    `json:"total_latency_ms"`
}

type AISolveService struct {
	problems     repository.ProblemRepository
	runs         repository.AISolveRunRepository
	llmClient    llm.Client
	submissions  JudgeSubmitter
	defaultModel string
}

func NewAISolveService(
	problems repository.ProblemRepository,
	runs repository.AISolveRunRepository,
	llmClient llm.Client,
	submissions JudgeSubmitter,
	defaultModel string,
) *AISolveService {
	return &AISolveService{
		problems:     problems,
		runs:         runs,
		llmClient:    llmClient,
		submissions:  submissions,
		defaultModel: defaultModel,
	}
}

func (s *AISolveService) Solve(ctx context.Context, input AISolveInput) (*AISolveOutput, error) {
	startedAt := time.Now()
	resolvedModel := firstNonEmpty(input.Model, s.defaultModel)
	run := &model.AISolveRun{
		ProblemID: input.ProblemID,
		Model:     resolvedModel,
		Status:    model.AISolveRunStatusRunning,
	}
	if err := s.runs.Create(ctx, run); err != nil {
		return nil, fmt.Errorf("create ai solve run: %w", err)
	}

	output := &AISolveOutput{
		AISolveRunID: run.ID,
		ProblemID:    input.ProblemID,
		Model:        resolvedModel,
	}

	problem, err := s.problems.GetByID(ctx, input.ProblemID)
	if err != nil {
		return s.failRun(ctx, run, output, startedAt, err.Error(), err)
	}

	prompt := buildSolvePrompt(problem)
	run.PromptPreview = truncateForPreview(prompt, 800)
	output.PromptPreview = run.PromptPreview

	llmStartedAt := time.Now()
	llmResp, err := s.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		Model:  input.Model,
	})
	run.LLMLatencyMS = elapsedMS(llmStartedAt)
	if err != nil {
		return s.failRun(ctx, run, output, startedAt, fmt.Sprintf("%v: %v", ErrAISolveLLMFailed, err), fmt.Errorf("%w: %v", ErrAISolveLLMFailed, err))
	}

	run.Model = firstNonEmpty(llmResp.Model, resolvedModel)
	run.TokenInput = llmResp.InputTokens
	run.TokenOutput = llmResp.OutputTokens
	output.Model = run.Model
	run.RawResponse = llmResp.Content
	output.RawResponse = truncateForPreview(llmResp.Content, 4000)

	code := extractCPPCode(llmResp.Content)
	if strings.TrimSpace(code) == "" {
		return s.failRun(ctx, run, output, startedAt, ErrAISolveCodeNotExtracted.Error(), ErrAISolveCodeNotExtracted)
	}
	run.ExtractedCode = code
	output.ExtractedCode = code

	judgeOutput, err := s.submissions.Submit(ctx, JudgeSubmissionInput{
		ProblemID:  input.ProblemID,
		SourceCode: code,
		Language:   model.LanguageCPP17,
		SourceType: model.SourceTypeAI,
	})
	if err != nil {
		return s.failRun(ctx, run, output, startedAt, err.Error(), err)
	}

	run.Status = model.AISolveRunStatusSuccess
	run.ErrorMessage = judgeOutput.ErrorMessage
	run.Verdict = judgeOutput.Verdict
	run.SubmissionID = &judgeOutput.SubmissionID
	run.TotalLatencyMS = elapsedMS(startedAt)
	if err := s.runs.Update(ctx, run); err != nil {
		return nil, fmt.Errorf("update ai solve run: %w", err)
	}

	syncAISolveOutputFromRun(output, run)
	output.SubmissionID = judgeOutput.SubmissionID
	output.Verdict = judgeOutput.Verdict
	output.ErrorMessage = judgeOutput.ErrorMessage
	return output, nil
}

func (s *AISolveService) GetRun(ctx context.Context, runID uint) (*model.AISolveRun, error) {
	return s.runs.GetByID(ctx, runID)
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
	if err := s.runs.Update(ctx, run); err != nil {
		return nil, fmt.Errorf("update ai solve run: %w", err)
	}
	syncAISolveOutputFromRun(output, run)
	return output, returnErr
}

func buildSolvePrompt(problem *model.Problem) string {
	return strings.TrimSpace(fmt.Sprintf(`
You are solving an online judge problem.

Please write a correct solution in C++17.
Return the final answer as a markdown code block with language tag cpp.
Do not include explanation outside the code unless necessary.

Problem Title:
%s

Problem Description:
%s

Input Specification:
%s

Output Specification:
%s

Samples:
%s
`, problem.Title, problem.Description, problem.InputSpec, problem.OutputSpec, problem.Samples))
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
	if limit <= 0 || len(value) <= limit {
		return value
	}
	return value[:limit] + "...(truncated)"
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

func syncAISolveOutputFromRun(output *AISolveOutput, run *model.AISolveRun) {
	if output == nil || run == nil {
		return
	}
	output.Model = run.Model
	output.ErrorMessage = run.ErrorMessage
	output.TokenInput = run.TokenInput
	output.TokenOutput = run.TokenOutput
	output.LLMLatencyMS = run.LLMLatencyMS
	output.TotalLatencyMS = run.TotalLatencyMS
}
