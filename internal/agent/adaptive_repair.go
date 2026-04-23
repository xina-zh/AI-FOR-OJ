package agent

import (
	"context"

	"ai-for-oj/internal/llm"
)

type adaptiveRepairStrategy struct{}

func (adaptiveRepairStrategy) Name() string {
	return AdaptiveRepairV1AgentName
}

func (adaptiveRepairStrategy) Execute(ctx context.Context, client llm.Client, input SolveInput) (SolveOutput, error) {
	if input.JudgeSubmitter == nil {
		output, err := directCodegenStrategy{}.Execute(ctx, client, input)
		output.AgentName = AdaptiveRepairV1AgentName
		return output, err
	}

	result, err := NewAdaptiveRepairCoordinator(DefaultAdaptiveRepairMaxAttempts).Execute(ctx, client, input)
	if err != nil {
		return result.SolveOutput, err
	}
	result.SolveOutput.AgentName = AdaptiveRepairV1AgentName
	return result.SolveOutput, nil
}
