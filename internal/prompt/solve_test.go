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

func TestBuildVerdictSpecificRepairPromptsUseDistinctInstructions(t *testing.T) {
	problem := &model.Problem{
		Title:       "Repair",
		Description: "fix the solution",
		InputSpec:   "input",
		OutputSpec:  "output",
		Samples:     `[{"input":"1","output":"1"}]`,
	}

	waPrompt := BuildWARepairPrompt(problem, "int main() { return 0; }", "wrong answer on edge case")
	rePrompt := BuildRERepairPrompt(problem, "int main() { return 0; }", "segmentation fault")
	tlePrompt := BuildTLERepairPrompt(problem, "int main() { return 0; }", "time limit exceeded")

	for name, prompt := range map[string]string{
		"wa":  waPrompt,
		"re":  rePrompt,
		"tle": tlePrompt,
	} {
		if !strings.Contains(prompt, "PROMPT_TEMPLATE: repair_"+name) {
			t.Fatalf("expected %s repair prompt marker, got %q", name, prompt)
		}
	}

	if !strings.Contains(waPrompt, "diagnose the mistake") ||
		!strings.Contains(waPrompt, "at least 3 edge cases") ||
		!strings.Contains(waPrompt, "corrected algorithm") {
		t.Fatalf("expected WA repair prompt to require diagnosis, edge cases, and corrected algorithm, got %q", waPrompt)
	}
	if !strings.Contains(rePrompt, "root cause in implementation safety") ||
		!strings.Contains(rePrompt, "robust full code") {
		t.Fatalf("expected RE repair prompt to require safety diagnosis and robust code, got %q", rePrompt)
	}
	if !strings.Contains(tlePrompt, "old vs new complexity") ||
		!strings.Contains(tlePrompt, "more efficient rewrite") {
		t.Fatalf("expected TLE repair prompt to require complexity comparison and rewrite, got %q", tlePrompt)
	}
}
