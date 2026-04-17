package agent

import (
	"context"
	"errors"
	"testing"

	"ai-for-oj/internal/llm"
	"ai-for-oj/internal/model"
)

type fakeSolveLLMClient struct {
	responses []llm.GenerateResponse
	errors    []error
	requests  []llm.GenerateRequest
}

func (c *fakeSolveLLMClient) Generate(_ context.Context, req llm.GenerateRequest) (llm.GenerateResponse, error) {
	c.requests = append(c.requests, req)
	index := len(c.requests) - 1
	if index < len(c.errors) && c.errors[index] != nil {
		return llm.GenerateResponse{}, c.errors[index]
	}
	if index < len(c.responses) {
		return c.responses[index], nil
	}
	return llm.GenerateResponse{}, nil
}

func TestResolveSolveAgentNameAdaptiveRepairV1(t *testing.T) {
	got, err := ResolveSolveAgentName("adaptive_repair_v1")
	if err != nil {
		t.Fatalf("ResolveSolveAgentName returned error: %v", err)
	}
	if got != "adaptive_repair_v1" {
		t.Fatalf("ResolveSolveAgentName = %q, want %q", got, "adaptive_repair_v1")
	}
}

func TestResolveSolveStrategyAdaptiveRepairV1(t *testing.T) {
	got, err := ResolveSolveStrategy("adaptive_repair_v1")
	if err != nil {
		t.Fatalf("ResolveSolveStrategy returned error: %v", err)
	}
	if got == nil {
		t.Fatal("ResolveSolveStrategy returned nil strategy")
	}
	if got.Name() != "adaptive_repair_v1" {
		t.Fatalf("ResolveSolveStrategy.Name() = %q, want %q", got.Name(), "adaptive_repair_v1")
	}
}

func TestSupportsSelfRepairDoesNotDriveAdaptiveRepair(t *testing.T) {
	if SupportsSelfRepair("adaptive_repair_v1") {
		t.Fatal("SupportsSelfRepair should not report adaptive_repair_v1 as self-repair capable")
	}
}

func TestAnalyzeThenCodegenFailureKeepsAnalysisModelPriority(t *testing.T) {
	client := &fakeSolveLLMClient{
		responses: []llm.GenerateResponse{
			{
				Model:   "analysis-model",
				Content: "analysis text",
			},
		},
		errors: []error{
			nil,
			errors.New("codegen failed"),
		},
	}

	output, err := analyzeThenCodegenStrategy{}.Execute(context.Background(), client, SolveInput{
		Problem: &model.Problem{
			Title:       "Echo",
			Description: "echo input",
			InputSpec:   "one line",
			OutputSpec:  "same line",
			Samples:     "[]",
		},
		Model:      "input-model",
		PromptName: "default",
	})
	if err == nil {
		t.Fatal("analyzeThenCodegenStrategy.Execute returned nil error, want failure")
	}

	if output.Model != "analysis-model" {
		t.Fatalf("output.Model = %q, want %q", output.Model, "analysis-model")
	}
}

func TestAnalyzeThenCodegenKeepsInputModelOnAnalysisError(t *testing.T) {
	client := &fakeSolveLLMClient{
		errors: []error{errors.New("analysis failed")},
	}

	output, err := analyzeThenCodegenStrategy{}.Execute(context.Background(), client, SolveInput{
		Problem: &model.Problem{
			Title:       "Echo",
			Description: "echo input",
			InputSpec:   "one line",
			OutputSpec:  "same line",
			Samples:     "[]",
		},
		Model:      "input-model",
		PromptName: "default",
	})
	if err == nil {
		t.Fatal("analyzeThenCodegenStrategy.Execute returned nil error, want failure")
	}

	if output.Model != "input-model" {
		t.Fatalf("output.Model = %q, want %q", output.Model, "input-model")
	}
}

func TestDirectCodegenKeepsInputModelOnError(t *testing.T) {
	client := &fakeSolveLLMClient{
		errors: []error{errors.New("generate failed")},
	}

	output, err := directCodegenStrategy{}.Execute(context.Background(), client, SolveInput{
		Problem: &model.Problem{
			Title:       "Echo",
			Description: "echo input",
			InputSpec:   "one line",
			OutputSpec:  "same line",
			Samples:     "[]",
		},
		Model:      "input-model",
		PromptName: "default",
	})
	if err == nil {
		t.Fatal("directCodegenStrategy.Execute returned nil error, want failure")
	}

	if output.Model != "input-model" {
		t.Fatalf("output.Model = %q, want %q", output.Model, "input-model")
	}
}

func TestDirectCodegenRepairKeepsInputModelOnError(t *testing.T) {
	client := &fakeSolveLLMClient{
		errors: []error{errors.New("generate failed")},
	}

	output, err := directCodegenRepairStrategy{}.Execute(context.Background(), client, SolveInput{
		Problem: &model.Problem{
			Title:       "Echo",
			Description: "echo input",
			InputSpec:   "one line",
			OutputSpec:  "same line",
			Samples:     "[]",
		},
		Model:      "input-model",
		PromptName: "default",
	})
	if err == nil {
		t.Fatal("directCodegenRepairStrategy.Execute returned nil error, want failure")
	}

	if output.Model != "input-model" {
		t.Fatalf("output.Model = %q, want %q", output.Model, "input-model")
	}
}
