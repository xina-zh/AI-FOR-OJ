package prompt

import "ai-for-oj/internal/model"

func BuildRERepairPrompt(problem *model.Problem, previousCode, feedback string) string {
	return buildVerdictRepairPrompt(problem, "repair_re", []string{
		"Identify the root cause in implementation safety.",
		"Explain how the fix avoids crashes, invalid memory access, and other runtime hazards.",
		"Then provide the robust full code as one submit-ready C++17 program.",
	}, previousCode, feedback)
}
