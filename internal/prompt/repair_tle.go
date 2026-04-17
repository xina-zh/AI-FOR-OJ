package prompt

import "ai-for-oj/internal/model"

func BuildTLERepairPrompt(problem *model.Problem, promptName, previousCode, feedback string) string {
	return buildVerdictRepairPrompt(problem, promptName, "repair_tle", []string{
		"Compare the old vs new complexity clearly.",
		"Rewrite the algorithm in a more efficient form.",
		"Then provide the more efficient rewrite as one submit-ready C++17 program.",
	}, previousCode, feedback)
}
