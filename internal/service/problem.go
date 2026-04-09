package service

import (
	"context"
	"fmt"

	"ai-for-oj/internal/model"
	"ai-for-oj/internal/repository"
)

type CreateProblemInput struct {
	Title         string
	Description   string
	InputSpec     string
	OutputSpec    string
	Samples       string
	TimeLimitMS   int
	MemoryLimitMB int
	Difficulty    string
	Tags          string
}

type ProblemOutput struct {
	ID            uint   `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	InputSpec     string `json:"input_spec"`
	OutputSpec    string `json:"output_spec"`
	Samples       string `json:"samples"`
	TimeLimitMS   int    `json:"time_limit_ms"`
	MemoryLimitMB int    `json:"memory_limit_mb"`
	Difficulty    string `json:"difficulty"`
	Tags          string `json:"tags"`
}

type CreateTestCaseInput struct {
	ProblemID      uint
	Input          string
	ExpectedOutput string
	IsSample       bool
}

type TestCaseOutput struct {
	ID             uint   `json:"id"`
	ProblemID      uint   `json:"problem_id"`
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
	IsSample       bool   `json:"is_sample"`
}

type ProblemService struct {
	problems  repository.ProblemRepository
	testCases repository.TestCaseRepository
}

func NewProblemService(
	problems repository.ProblemRepository,
	testCases repository.TestCaseRepository,
) *ProblemService {
	return &ProblemService{
		problems:  problems,
		testCases: testCases,
	}
}

func (s *ProblemService) Create(ctx context.Context, input CreateProblemInput) (*ProblemOutput, error) {
	problem := &model.Problem{
		Title:         input.Title,
		Description:   input.Description,
		InputSpec:     input.InputSpec,
		OutputSpec:    input.OutputSpec,
		Samples:       input.Samples,
		TimeLimitMS:   input.TimeLimitMS,
		MemoryLimitMB: input.MemoryLimitMB,
		Difficulty:    input.Difficulty,
		Tags:          input.Tags,
	}
	if err := s.problems.Create(ctx, problem); err != nil {
		return nil, fmt.Errorf("create problem: %w", err)
	}

	return toProblemOutput(problem), nil
}

func (s *ProblemService) List(ctx context.Context) ([]ProblemOutput, error) {
	problems, err := s.problems.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list problems: %w", err)
	}

	outputs := make([]ProblemOutput, 0, len(problems))
	for _, problem := range problems {
		problemCopy := problem
		outputs = append(outputs, *toProblemOutput(&problemCopy))
	}
	return outputs, nil
}

func (s *ProblemService) Get(ctx context.Context, problemID uint) (*ProblemOutput, error) {
	problem, err := s.problems.GetByID(ctx, problemID)
	if err != nil {
		return nil, err
	}
	return toProblemOutput(problem), nil
}

func (s *ProblemService) CreateTestCase(ctx context.Context, input CreateTestCaseInput) (*TestCaseOutput, error) {
	if _, err := s.problems.GetByID(ctx, input.ProblemID); err != nil {
		return nil, err
	}

	testCase := &model.TestCase{
		ProblemID:      input.ProblemID,
		Input:          input.Input,
		ExpectedOutput: input.ExpectedOutput,
		IsSample:       input.IsSample,
	}
	if err := s.testCases.Create(ctx, testCase); err != nil {
		return nil, fmt.Errorf("create test case: %w", err)
	}

	return toTestCaseOutput(testCase), nil
}

func (s *ProblemService) ListTestCases(ctx context.Context, problemID uint) ([]TestCaseOutput, error) {
	if _, err := s.problems.GetByID(ctx, problemID); err != nil {
		return nil, err
	}

	testCases, err := s.testCases.ListByProblemID(ctx, problemID)
	if err != nil {
		return nil, fmt.Errorf("list test cases: %w", err)
	}

	outputs := make([]TestCaseOutput, 0, len(testCases))
	for _, testCase := range testCases {
		testCaseCopy := testCase
		outputs = append(outputs, *toTestCaseOutput(&testCaseCopy))
	}
	return outputs, nil
}

func toProblemOutput(problem *model.Problem) *ProblemOutput {
	return &ProblemOutput{
		ID:            problem.ID,
		Title:         problem.Title,
		Description:   problem.Description,
		InputSpec:     problem.InputSpec,
		OutputSpec:    problem.OutputSpec,
		Samples:       problem.Samples,
		TimeLimitMS:   problem.TimeLimitMS,
		MemoryLimitMB: problem.MemoryLimitMB,
		Difficulty:    problem.Difficulty,
		Tags:          problem.Tags,
	}
}

func toTestCaseOutput(testCase *model.TestCase) *TestCaseOutput {
	return &TestCaseOutput{
		ID:             testCase.ID,
		ProblemID:      testCase.ProblemID,
		Input:          testCase.Input,
		ExpectedOutput: testCase.ExpectedOutput,
		IsSample:       testCase.IsSample,
	}
}
