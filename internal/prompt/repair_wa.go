package prompt

import "ai-for-oj/internal/model"

func BuildWARepairPrompt(problem *model.Problem, promptName, previousCode, feedback string) string {
	return buildVerdictRepairPrompt(problem, promptName, "repair_wa", []string{
		"diagnose the mistake in the previous solution.",
		"list at least 3 edge cases the repaired solution must handle.",
		"explain the corrected algorithm before writing code.",
		"then provide the full code as one submit-ready C++17 program.",
	}, previousCode, feedback)
}
