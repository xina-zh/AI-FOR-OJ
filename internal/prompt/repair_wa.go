package prompt

import (
	"fmt"
	"strings"

	"ai-for-oj/internal/model"
)

func BuildWARepairPrompt(problem *model.Problem, previousCode, feedback string) string {
	return buildVerdictRepairPrompt(problem, "WA", previousCode, feedback, `
Focus on wrong-answer diagnosis:
- Identify the likely logic or algorithm mistake.
- Consider edge cases explicitly.
- Provide the algorithm correction before writing code.
`)
}

func buildVerdictRepairPrompt(problem *model.Problem, verdict, previousCode, feedback, guidance string) string {
	return strings.TrimSpace(fmt.Sprintf(`
You are repairing an online judge / competitive programming solution.

Use C++17 only.
Return exactly one markdown cpp code block.
Do not include prose outside the code block.

Problem Title:
%s

Problem Description:
%s

Input Specification:
%s

Output Specification:
%s

Samples:
%s

Previous Code:
%s

Judge Verdict: %s

Judge Feedback:
%s

Repair Guidance:
%s
`, problem.Title, problem.Description, problem.InputSpec, problem.OutputSpec, problem.Samples, strings.TrimSpace(previousCode), verdict, strings.TrimSpace(feedback), strings.TrimSpace(guidance)))
}
