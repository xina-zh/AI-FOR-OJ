package agent

import (
	"context"
	"errors"
	"strings"

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
	execution, err := executeLLMOnce(ctx, client, input.Model, finalPrompt)
	if err != nil {
		return SolveOutput{
			AgentName:     DirectCodegenAgentName,
			Model:         input.Model,
			PromptPreview: execution.PromptPreview,
			LLMLatencyMS:  execution.LLMLatencyMS,
		}, err
	}

	return SolveOutput{
		AgentName:     DirectCodegenAgentName,
		Model:         execution.Model,
		PromptPreview: execution.PromptPreview,
		RawResponse:   execution.RawResponse,
		TokenInput:    execution.TokenInput,
		TokenOutput:   execution.TokenOutput,
		LLMLatencyMS:  execution.LLMLatencyMS,
	}, nil
}

type analyzeThenCodegenStrategy struct{}

type directCodegenRepairStrategy struct{}

func (directCodegenRepairStrategy) Name() string {
	return DirectCodegenRepairAgentName
}

func (directCodegenRepairStrategy) Execute(ctx context.Context, client llm.Client, input SolveInput) (SolveOutput, error) {
	finalPrompt := prompt.BuildSolvePrompt(input.Problem, input.PromptName)
	execution, err := executeLLMOnce(ctx, client, input.Model, finalPrompt)
	if err != nil {
		return SolveOutput{
			AgentName:     DirectCodegenRepairAgentName,
			Model:         input.Model,
			PromptPreview: execution.PromptPreview,
			LLMLatencyMS:  execution.LLMLatencyMS,
		}, err
	}

	return SolveOutput{
		AgentName:     DirectCodegenRepairAgentName,
		Model:         execution.Model,
		PromptPreview: execution.PromptPreview,
		RawResponse:   execution.RawResponse,
		TokenInput:    execution.TokenInput,
		TokenOutput:   execution.TokenOutput,
		LLMLatencyMS:  execution.LLMLatencyMS,
	}, nil
}

func (analyzeThenCodegenStrategy) Name() string {
	return AnalyzeThenCodegenAgentName
}

func (analyzeThenCodegenStrategy) Execute(ctx context.Context, client llm.Client, input SolveInput) (SolveOutput, error) {
	analysisPrompt := prompt.BuildAnalysisPrompt(input.Problem)
	analysisExecution, err := executeLLMOnce(ctx, client, input.Model, analysisPrompt)
	if err != nil {
		return SolveOutput{
			AgentName:     AnalyzeThenCodegenAgentName,
			Model:         input.Model,
			PromptPreview: analysisExecution.PromptPreview,
			LLMLatencyMS:  analysisExecution.LLMLatencyMS,
		}, err
	}

	finalPrompt := prompt.BuildSolvePromptWithAnalysis(input.Problem, input.PromptName, analysisExecution.RawResponse)
	codeExecution, err := executeLLMOnce(ctx, client, input.Model, finalPrompt)
	if err != nil {
		return SolveOutput{
			AgentName:       AnalyzeThenCodegenAgentName,
			Model:           effectiveModel(analysisExecution.Model, input.Model),
			PromptPreview:   codeExecution.PromptPreview,
			TokenInput:      analysisExecution.TokenInput,
			TokenOutput:     analysisExecution.TokenOutput,
			LLMLatencyMS:    analysisExecution.LLMLatencyMS + codeExecution.LLMLatencyMS,
			AnalysisPreview: analysisExecution.RawResponse,
		}, err
	}

	return SolveOutput{
		AgentName:       AnalyzeThenCodegenAgentName,
		Model:           effectiveModel(codeExecution.Model, analysisExecution.Model, input.Model),
		PromptPreview:   codeExecution.PromptPreview,
		RawResponse:     codeExecution.RawResponse,
		TokenInput:      analysisExecution.TokenInput + codeExecution.TokenInput,
		TokenOutput:     analysisExecution.TokenOutput + codeExecution.TokenOutput,
		LLMLatencyMS:    analysisExecution.LLMLatencyMS + codeExecution.LLMLatencyMS,
		AnalysisPreview: analysisExecution.RawResponse,
	}, nil
}
