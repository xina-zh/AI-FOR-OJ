package prompt

import (
	"errors"
	"fmt"
	"strings"

	"ai-for-oj/internal/model"
)

const (
	DefaultSolvePromptName      = "default"
	CPP17MinimalSolvePromptName = "cpp17_minimal"
	StrictCPP17SolvePromptName  = "strict_cpp17"
)

var ErrUnknownSolvePrompt = errors.New("unknown solve prompt")

func ResolveSolvePromptName(name string) (string, error) {
	switch strings.TrimSpace(name) {
	case "":
		return DefaultSolvePromptName, nil
	case DefaultSolvePromptName:
		return DefaultSolvePromptName, nil
	case CPP17MinimalSolvePromptName:
		return CPP17MinimalSolvePromptName, nil
	case StrictCPP17SolvePromptName:
		return StrictCPP17SolvePromptName, nil
	default:
		return "", ErrUnknownSolvePrompt
	}
}

func DisplaySolvePromptName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return DefaultSolvePromptName
	}
	return trimmed
}

func BuildSolvePrompt(problem *model.Problem, promptName string) string {
	switch promptName {
	case CPP17MinimalSolvePromptName:
		return buildCPP17MinimalSolvePrompt(problem)
	case StrictCPP17SolvePromptName:
		return buildStrictCPP17SolvePrompt(problem)
	default:
		return buildDefaultSolvePrompt(problem)
	}
}

func BuildAnalysisPrompt(problem *model.Problem) string {
	return strings.TrimSpace(fmt.Sprintf(`
You are preparing to solve an online judge / competitive programming problem.

First produce a concise analysis instead of code.
Please include:
1. Problem understanding
2. Algorithm idea
3. Boundary cases / pitfalls

Do not write C++ code in this step.

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
`, problem.Title, problem.Description, problem.InputSpec, problem.OutputSpec, problem.Samples))
}

func BuildSolvePromptWithAnalysis(problem *model.Problem, promptName, analysis string) string {
	base := BuildSolvePrompt(problem, promptName)
	analysis = strings.TrimSpace(analysis)
	if analysis == "" {
		return base
	}

	return strings.TrimSpace(fmt.Sprintf(`
%s

Prior Analysis:
%s

Use the analysis above to produce the final answer now.
`, base, analysis))
}

func BuildRepairPrompt(problem *model.Problem, promptName, previousCode, feedback string) string {
	base := BuildSolvePrompt(problem, promptName)
	previousCode = strings.TrimSpace(previousCode)
	feedback = strings.TrimSpace(feedback)

	return strings.TrimSpace(fmt.Sprintf(`
%s

Your previous submission failed.
Please repair the existing C++17 solution based on the judge feedback below.
Do not ignore the failure reason.
Prefer minimal fixes when possible, but if necessary you may rewrite the solution completely.
Return a complete submit-ready C++17 program as exactly one markdown cpp code block.

Previous Code (cpp):
%s

Judge Feedback:
%s
`, base, previousCode, feedback))
}

func buildVerdictRepairPrompt(problem *model.Problem, promptName, templateName string, instructions []string, previousCode, feedback string) string {
	base := BuildSolvePrompt(problem, promptName)
	previousCode = strings.TrimSpace(previousCode)
	feedback = strings.TrimSpace(feedback)

	lines := make([]string, 0, len(instructions))
	for i, instruction := range instructions {
		lines = append(lines, fmt.Sprintf("%d. %s", i+1, instruction))
	}

	return strings.TrimSpace(fmt.Sprintf(`
%s

PROMPT_TEMPLATE: %s

Your previous submission failed.
Repair the solution using the verdict-specific guidance below.

Repair requirements:
%s

Previous Code (cpp):
%s

Judge Feedback:
%s
`, base, templateName, strings.Join(lines, "\n"), previousCode, feedback))
}

func buildDefaultSolvePrompt(problem *model.Problem) string {
	return strings.TrimSpace(fmt.Sprintf(`
PROMPT_TEMPLATE: default

You are solving an online judge problem.

Please write a correct solution in C++17.
Return the final answer as a markdown code block with language tag cpp.
Do not include explanation outside the code unless necessary.

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
`, problem.Title, problem.Description, problem.InputSpec, problem.OutputSpec, problem.Samples))
}

func buildCPP17MinimalSolvePrompt(problem *model.Problem) string {
	return strings.TrimSpace(fmt.Sprintf(`
PROMPT_TEMPLATE: cpp17_minimal

Solve this competitive programming problem.

Return one cpp code block only.
Use C++17.
Use stdin/stdout.

Problem Title:
%s

Problem Description:
%s

Input:
%s

Output:
%s

Samples:
%s
`, problem.Title, problem.Description, problem.InputSpec, problem.OutputSpec, problem.Samples))
}

func buildStrictCPP17SolvePrompt(problem *model.Problem) string {
	return strings.TrimSpace(fmt.Sprintf(`
PROMPT_TEMPLATE: strict_cpp17

You are solving an online judge / competitive programming problem.

Requirements:
1. Output a complete compilable C++17 program.
2. Return exactly one markdown code block with language tag cpp.
3. Use standard input and standard output only.
4. Do not add explanation, notes, or multiple code blocks.
5. Do not output prose before or after the code block.

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
`, problem.Title, problem.Description, problem.InputSpec, problem.OutputSpec, problem.Samples))
}
