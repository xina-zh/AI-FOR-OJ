package agent

import (
	"context"
	"regexp"
	"strings"

	"ai-for-oj/internal/llm"
)

var (
	cppFencePattern     = regexp.MustCompile("(?is)```(?:cpp|c\\+\\+|cc|cxx)\\s*(.*?)```")
	genericFencePattern = regexp.MustCompile("(?is)```(?:[a-z0-9_+-]+)?\\s*(.*?)```")
)

func ExecutePrompt(ctx context.Context, client llm.Client, modelName, promptText string) (SolveOutput, error) {
	resp, latencyMS, err := generateOnce(ctx, client, modelName, promptText)
	output := SolveOutput{
		Model:         effectiveModel(resp.Model, modelName),
		PromptPreview: promptText,
		RawResponse:   resp.Content,
		TokenInput:    resp.InputTokens,
		TokenOutput:   resp.OutputTokens,
		LLMLatencyMS:  latencyMS,
	}
	return output, err
}

func ExtractCPPCode(raw string) string {
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
	return ""
}
