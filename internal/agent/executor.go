package agent

import (
	"context"
	"strings"
	"time"

	"ai-for-oj/internal/llm"
)

type llmExecution struct {
	PromptPreview string
	RawResponse   string
	TokenInput    int64
	TokenOutput   int64
	LLMLatencyMS  int
	Model         string
}

func executeLLMOnce(ctx context.Context, client llm.Client, modelName, promptText string) (llmExecution, error) {
	startedAt := time.Now()
	resp, err := client.Generate(ctx, llm.GenerateRequest{
		Prompt: promptText,
		Model:  modelName,
	})

	return llmExecution{
		PromptPreview: promptText,
		RawResponse:   resp.Content,
		TokenInput:    resp.InputTokens,
		TokenOutput:   resp.OutputTokens,
		LLMLatencyMS:  elapsedMS(startedAt),
		Model:         effectiveModel(resp.Model, modelName),
	}, err
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
