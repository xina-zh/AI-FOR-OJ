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

func TestBuildVerdictSpecificRepairPrompts(t *testing.T) {
	problem := &model.Problem{
		Title:       "Echo",
		Description: "echo input",
		InputSpec:   "one line",
		OutputSpec:  "same line",
		Samples:     `[{"input":"hi","output":"hi"}]`,
	}
	previousCode := "int main(){return 0;}"
	feedback := "case #1 expected hi got bye"

	tests := []struct {
		name     string
		prompt   string
		verdict  string
		required []string
	}{
		{
			name:    "wrong answer",
			prompt:  BuildWARepairPrompt(problem, previousCode, feedback),
			verdict: "WA",
			required: []string{
				"edge cases",
				"algorithm correction",
			},
		},
		{
			name:    "runtime error",
			prompt:  BuildRERepairPrompt(problem, previousCode, feedback),
			verdict: "RE",
			required: []string{
				"safety checks",
				"runtime robustness",
			},
		},
		{
			name:    "time limit",
			prompt:  BuildTLERepairPrompt(problem, previousCode, feedback),
			verdict: "TLE",
			required: []string{
				"complexity comparison",
				"algorithm rewrite",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			required := append([]string{
				problem.Description,
				previousCode,
				feedback,
				"Judge Verdict: " + tt.verdict,
				"C++17",
				"Return exactly one markdown cpp code block",
			}, tt.required...)
			for _, needle := range required {
				if !strings.Contains(tt.prompt, needle) {
					t.Fatalf("expected prompt to include %q, got %q", needle, tt.prompt)
				}
			}
		})
	}
}
