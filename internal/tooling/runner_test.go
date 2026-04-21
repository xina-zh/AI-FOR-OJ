package tooling

import (
	"context"
	"errors"
	"testing"
)

type fakeTool struct {
	name  string
	calls int
}

func (t fakeTool) Name() string {
	return t.name
}

func (t fakeTool) Run(context.Context, CallInput) (CallOutput, error) {
	return CallOutput{
		ToolName: t.name,
		Status:   CallStatusOK,
		Summary:  "ok",
	}, nil
}

type countingTool struct {
	name  string
	calls int
}

func (t *countingTool) Name() string {
	return t.name
}

func (t *countingTool) Run(context.Context, CallInput) (CallOutput, error) {
	t.calls++
	return CallOutput{
		ToolName: t.name,
		Status:   CallStatusOK,
		Summary:  "ok",
	}, nil
}

func TestRunnerRejectsDisabledTool(t *testing.T) {
	registry := NewRegistry()
	registry.Register(&countingTool{name: "sample_judge"})
	runner := registry.NewRunner(Config{})

	_, err := runner.Call(context.Background(), "sample_judge", CallInput{})
	if !errors.Is(err, ErrToolDisabled) {
		t.Fatalf("expected ErrToolDisabled, got %v", err)
	}
}

func TestRunnerEnforcesGlobalCallLimit(t *testing.T) {
	registry := NewRegistry()
	registry.Register(&countingTool{name: "sample_judge"})
	runner := registry.NewRunner(Config{
		Enabled:  []string{"sample_judge"},
		MaxCalls: 1,
	})

	if _, err := runner.Call(context.Background(), "sample_judge", CallInput{}); err != nil {
		t.Fatalf("first call returned error: %v", err)
	}
	if _, err := runner.Call(context.Background(), "sample_judge", CallInput{}); !errors.Is(err, ErrToolCallLimitHit) {
		t.Fatalf("expected ErrToolCallLimitHit, got %v", err)
	}
}

func TestRunnerEnforcesPerToolCallLimit(t *testing.T) {
	registry := NewRegistry()
	registry.Register(&countingTool{name: "sample_judge"})
	registry.Register(&countingTool{name: "trace_lookup"})
	runner := registry.NewRunner(Config{
		Enabled:         []string{"sample_judge", "trace_lookup"},
		MaxCalls:        3,
		PerToolMaxCalls: map[string]int{"sample_judge": 1},
	})

	if _, err := runner.Call(context.Background(), "sample_judge", CallInput{}); err != nil {
		t.Fatalf("first sample_judge call returned error: %v", err)
	}
	if _, err := runner.Call(context.Background(), "sample_judge", CallInput{}); !errors.Is(err, ErrToolCallLimitHit) {
		t.Fatalf("expected per-tool ErrToolCallLimitHit, got %v", err)
	}
	if _, err := runner.Call(context.Background(), "trace_lookup", CallInput{}); err != nil {
		t.Fatalf("expected different enabled tool to remain callable, got %v", err)
	}
	if runner.CallCount() != 2 {
		t.Fatalf("expected two successful calls, got %d", runner.CallCount())
	}
}
