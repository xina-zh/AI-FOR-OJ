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
	output, err := directCodegenStrategy{}.Execute(ctx, client, input)
	output.AgentName = AdaptiveRepairV1AgentName
	return output, err
}
