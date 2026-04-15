package prompt

import (
	"strings"
	"testing"

	"ai-for-oj/internal/model"
)

func TestBuildSolvePromptUsesDifferentTemplates(t *testing.T) {
	problem := &model.Problem{
		Title:       "Echo",
		Description: "echo input",
		InputSpec:   "one line",
		OutputSpec:  "same line",
		Samples:     `[{"input":"hi","output":"hi"}]`,
	}

	defaultPrompt := BuildSolvePrompt(problem, DefaultSolvePromptName)
	minimalPrompt := BuildSolvePrompt(problem, CPP17MinimalSolvePromptName)
	strictPrompt := BuildSolvePrompt(problem, StrictCPP17SolvePromptName)

	if !strings.Contains(defaultPrompt, "PROMPT_TEMPLATE: default") {
		t.Fatalf("expected default prompt marker, got %q", defaultPrompt)
	}
	if !strings.Contains(minimalPrompt, "PROMPT_TEMPLATE: cpp17_minimal") {
		t.Fatalf("expected minimal prompt marker, got %q", minimalPrompt)
	}
	if !strings.Contains(strictPrompt, "PROMPT_TEMPLATE: strict_cpp17") {
		t.Fatalf("expected strict prompt marker, got %q", strictPrompt)
	}

	if defaultPrompt == minimalPrompt || defaultPrompt == strictPrompt || minimalPrompt == strictPrompt {
		t.Fatalf("expected all prompt templates to differ")
	}
}
