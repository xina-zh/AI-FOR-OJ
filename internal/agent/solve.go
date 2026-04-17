package agent

import (
	"context"
	"errors"
	"strings"
	"time"

	"ai-for-oj/internal/llm"
	"ai-for-oj/internal/model"
	"ai-for-oj/internal/prompt"
)

const (
	DirectCodegenAgentName       = "direct_codegen"
	DirectCodegenRepairAgentName = "direct_codegen_repair"
	AnalyzeThenCodegenAgentName  = "analyze_then_codegen"
	AdaptiveRepairV1AgentName    = "adaptive_repair_v1"
)

var ErrUnknownSolveAgent = errors.New("unknown solve agent")

type SolveInput struct {
	Problem    *model.Problem
	Model      string
	PromptName string
}

type SolveOutput struct {
	AgentName       string
	Model           string
	PromptPreview   string
	RawResponse     string
	TokenInput      int64
	TokenOutput     int64
	LLMLatencyMS    int
	AnalysisPreview string
}

type SolveStrategy interface {
	Name() string
	Execute(ctx context.Context, client llm.Client, input SolveInput) (SolveOutput, error)
}

func ResolveSolveAgentName(name string) (string, error) {
	switch strings.TrimSpace(name) {
	case "":
		return DirectCodegenAgentName, nil
	case DirectCodegenAgentName:
		return DirectCodegenAgentName, nil
	case DirectCodegenRepairAgentName:
		return DirectCodegenRepairAgentName, nil
	case AnalyzeThenCodegenAgentName:
		return AnalyzeThenCodegenAgentName, nil
	case AdaptiveRepairV1AgentName:
		return AdaptiveRepairV1AgentName, nil
	default:
		return "", ErrUnknownSolveAgent
	}
}

func DisplaySolveAgentName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return DirectCodegenAgentName
	}
	return trimmed
}

func ResolveSolveStrategy(name string) (SolveStrategy, error) {
	resolvedName, err := ResolveSolveAgentName(name)
	if err != nil {
		return nil, err
	}

	switch resolvedName {
	case DirectCodegenRepairAgentName:
		return directCodegenRepairStrategy{}, nil
	case AnalyzeThenCodegenAgentName:
		return analyzeThenCodegenStrategy{}, nil
	case AdaptiveRepairV1AgentName:
		return adaptiveRepairStrategy{}, nil
	default:
		return directCodegenStrategy{}, nil
	}
}

func SupportsSelfRepair(name string) bool {
	return strings.TrimSpace(name) == DirectCodegenRepairAgentName
}

type directCodegenStrategy struct{}

func (directCodegenStrategy) Name() string {
	return DirectCodegenAgentName
}

func (directCodegenStrategy) Execute(ctx context.Context, client llm.Client, input SolveInput) (SolveOutput, error) {
	finalPrompt := prompt.BuildSolvePrompt(input.Problem, input.PromptName)
	resp, latencyMS, err := generateOnce(ctx, client, input.Model, finalPrompt)
	if err != nil {
		return SolveOutput{
			AgentName:     DirectCodegenAgentName,
			Model:         input.Model,
			PromptPreview: finalPrompt,
			LLMLatencyMS:  latencyMS,
		}, err
	}

	return SolveOutput{
		AgentName:     DirectCodegenAgentName,
		Model:         effectiveModel(resp.Model, input.Model),
		PromptPreview: finalPrompt,
		RawResponse:   resp.Content,
		TokenInput:    resp.InputTokens,
		TokenOutput:   resp.OutputTokens,
		LLMLatencyMS:  latencyMS,
	}, nil
}

type analyzeThenCodegenStrategy struct{}

type directCodegenRepairStrategy struct{}

func (directCodegenRepairStrategy) Name() string {
	return DirectCodegenRepairAgentName
}

func (directCodegenRepairStrategy) Execute(ctx context.Context, client llm.Client, input SolveInput) (SolveOutput, error) {
	finalPrompt := prompt.BuildSolvePrompt(input.Problem, input.PromptName)
	resp, latencyMS, err := generateOnce(ctx, client, input.Model, finalPrompt)
	if err != nil {
		return SolveOutput{
			AgentName:     DirectCodegenRepairAgentName,
			Model:         input.Model,
			PromptPreview: finalPrompt,
			LLMLatencyMS:  latencyMS,
		}, err
	}

	return SolveOutput{
		AgentName:     DirectCodegenRepairAgentName,
		Model:         effectiveModel(resp.Model, input.Model),
		PromptPreview: finalPrompt,
		RawResponse:   resp.Content,
		TokenInput:    resp.InputTokens,
		TokenOutput:   resp.OutputTokens,
		LLMLatencyMS:  latencyMS,
	}, nil
}

func (analyzeThenCodegenStrategy) Name() string {
	return AnalyzeThenCodegenAgentName
}

func (analyzeThenCodegenStrategy) Execute(ctx context.Context, client llm.Client, input SolveInput) (SolveOutput, error) {
	analysisPrompt := prompt.BuildAnalysisPrompt(input.Problem)
	analysisResp, analysisLatencyMS, err := generateOnce(ctx, client, input.Model, analysisPrompt)
	if err != nil {
		return SolveOutput{
			AgentName:     AnalyzeThenCodegenAgentName,
			Model:         input.Model,
			PromptPreview: analysisPrompt,
			LLMLatencyMS:  analysisLatencyMS,
		}, err
	}

	finalPrompt := prompt.BuildSolvePromptWithAnalysis(input.Problem, input.PromptName, analysisResp.Content)
	codeResp, codeLatencyMS, err := generateOnce(ctx, client, input.Model, finalPrompt)
	if err != nil {
		return SolveOutput{
			AgentName:       AnalyzeThenCodegenAgentName,
			Model:           effectiveModel(analysisResp.Model, input.Model),
			PromptPreview:   finalPrompt,
			TokenInput:      analysisResp.InputTokens,
			TokenOutput:     analysisResp.OutputTokens,
			LLMLatencyMS:    analysisLatencyMS + codeLatencyMS,
			AnalysisPreview: analysisResp.Content,
		}, err
	}

	return SolveOutput{
		AgentName:       AnalyzeThenCodegenAgentName,
		Model:           effectiveModel(codeResp.Model, analysisResp.Model, input.Model),
		PromptPreview:   finalPrompt,
		RawResponse:     codeResp.Content,
		TokenInput:      analysisResp.InputTokens + codeResp.InputTokens,
		TokenOutput:     analysisResp.OutputTokens + codeResp.OutputTokens,
		LLMLatencyMS:    analysisLatencyMS + codeLatencyMS,
		AnalysisPreview: analysisResp.Content,
	}, nil
}

func generateOnce(ctx context.Context, client llm.Client, modelName, promptText string) (llm.GenerateResponse, int, error) {
	startedAt := time.Now()
	resp, err := client.Generate(ctx, llm.GenerateRequest{
		Prompt: promptText,
		Model:  modelName,
	})
	return resp, elapsedMS(startedAt), err
}

func elapsedMS(start time.Time) int {
	if start.IsZero() {
		return 0
	}
	return int(time.Since(start).Milliseconds())
}

func effectiveModel(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}
