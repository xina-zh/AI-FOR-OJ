package prompt

import "ai-for-oj/internal/model"

func BuildRERepairPrompt(problem *model.Problem, previousCode, feedback string) string {
	return buildVerdictRepairPrompt(problem, "RE", previousCode, feedback, `
Focus on runtime-error repair:
- Add safety checks for indexes, empty data, division, recursion depth, and numeric bounds where relevant.
- Improve runtime robustness without changing the required input/output contract.
- Return a complete corrected program.
`)
}
