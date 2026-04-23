package prompt

import "ai-for-oj/internal/model"

func BuildTLERepairPrompt(problem *model.Problem, previousCode, feedback string) string {
	return buildVerdictRepairPrompt(problem, "TLE", previousCode, feedback, `
Focus on time-limit repair:
- Provide an old vs new complexity comparison.
- Prefer an algorithm rewrite when the current complexity is too high.
- Return a complete efficient C++17 program.
`)
}
