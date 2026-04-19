package tooling

import (
	"context"
	"testing"
)

type fakeTool struct {
	name  string
	calls int
}

func (t *fakeTool) Name() string {
	return t.name
}

func (t *fakeTool) Execute(_ context.Context, input CallInput) (CallResult, error) {
	t.calls++
	return CallResult{
		ToolName: t.name,
		Status:   CallStatusOK,
		Summary:  "called with " + input.SourceCode,
	}, nil
}

func TestRunnerRejectsDisabledTool(t *testing.T) {
	registry := NewRegistry()
	registry.Register(&fakeTool{name: "sample_judge"})
	runner := registry.NewRunner(Config{})

	_, err := runner.Call(context.Background(), "sample_judge", CallInput{SourceCode: "code"})
	if err == nil {
		t.Fatal("expected disabled tool call to fail")
	}
}

func TestRunnerExecutesEnabledToolAndTracksCalls(t *testing.T) {
	tool := &fakeTool{name: "sample_judge"}
	registry := NewRegistry()
	registry.Register(tool)
	runner := registry.NewRunner(Config{
		EnabledTools: []string{"sample_judge"},
		MaxCalls:     2,
	})

	result, err := runner.Call(context.Background(), "sample_judge", CallInput{SourceCode: "code"})
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if result.Status != CallStatusOK {
		t.Fatalf("unexpected result: %+v", result)
	}
	if runner.CallCount() != 1 || tool.calls != 1 {
		t.Fatalf("expected one call, runner=%d tool=%d", runner.CallCount(), tool.calls)
	}
}

func TestRunnerEnforcesGlobalCallLimit(t *testing.T) {
	registry := NewRegistry()
	registry.Register(&fakeTool{name: "sample_judge"})
	runner := registry.NewRunner(Config{
		EnabledTools: []string{"sample_judge"},
		MaxCalls:     1,
	})

	if _, err := runner.Call(context.Background(), "sample_judge", CallInput{}); err != nil {
		t.Fatalf("first call returned error: %v", err)
	}
	if _, err := runner.Call(context.Background(), "sample_judge", CallInput{}); err == nil {
		t.Fatal("expected second call to fail")
	}
}
