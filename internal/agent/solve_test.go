package agent

import (
	"context"
	"errors"
	"slices"
	"testing"

	"ai-for-oj/internal/llm"
	"ai-for-oj/internal/model"
	"ai-for-oj/internal/tooling"
)

func TestExtractCPPCode(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "cpp fence",
			raw:  "text\n```cpp\nint main(){return 0;}\n```",
			want: "int main(){return 0;}",
		},
		{
			name: "generic fence",
			raw:  "```\n#include <bits/stdc++.h>\n```",
			want: "#include <bits/stdc++.h>",
		},
		{
			name: "no fence",
			raw:  "int main(){return 0;}",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractCPPCode(tt.raw); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

type fakeAgentLLMClient struct {
	responses []llm.GenerateResponse
	errors    []error
	requests  []llm.GenerateRequest
}

func (c *fakeAgentLLMClient) Generate(_ context.Context, req llm.GenerateRequest) (llm.GenerateResponse, error) {
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

type fakeAgentTool struct {
	name  string
	err   error
	calls int
}

func (t *fakeAgentTool) Name() string {
	return t.name
}

func (t *fakeAgentTool) Run(context.Context, tooling.CallInput) (tooling.CallOutput, error) {
	t.calls++
	return tooling.CallOutput{ToolName: t.name, Status: tooling.CallStatusFailed, Summary: "sample failed"}, t.err
}

func TestToolingCodegenAgentRegistration(t *testing.T) {
	resolved, err := ResolveSolveAgentName(ToolingCodegenV1AgentName)
	if err != nil {
		t.Fatalf("ResolveSolveAgentName returned error: %v", err)
	}
	if resolved != ToolingCodegenV1AgentName {
		t.Fatalf("expected %s, got %s", ToolingCodegenV1AgentName, resolved)
	}
	if !slices.Contains(ListSolveAgents(), ToolingCodegenV1AgentName) {
		t.Fatalf("expected ListSolveAgents to include %s", ToolingCodegenV1AgentName)
	}
}

func TestToolingCodegenAgentCallsSampleJudgeWhenEnabled(t *testing.T) {
	tool := &fakeAgentTool{name: tooling.SampleJudgeToolName}
	registry := tooling.NewRegistry()
	registry.Register(tool)
	runner := registry.NewRunner(tooling.Config{Enabled: []string{tooling.SampleJudgeToolName}, MaxCalls: 1})
	client := &fakeAgentLLMClient{
		responses: []llm.GenerateResponse{
			{Model: "mock", Content: "```cpp\nint main(){return 0;}\n```", InputTokens: 10, OutputTokens: 20},
		},
	}
	strategy, err := ResolveSolveStrategy(ToolingCodegenV1AgentName)
	if err != nil {
		t.Fatalf("ResolveSolveStrategy returned error: %v", err)
	}

	output, err := strategy.Execute(context.Background(), client, SolveInput{
		Problem:       &model.Problem{BaseModel: model.BaseModel{ID: 1}},
		Model:         "mock",
		ToolingRunner: runner,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if tool.calls != 1 || output.ToolCallCount != 1 {
		t.Fatalf("expected one tool call, tool=%d output=%+v", tool.calls, output)
	}
}

func TestToolingCodegenAgentContinuesWhenToolCallFails(t *testing.T) {
	tool := &fakeAgentTool{name: tooling.SampleJudgeToolName, err: errors.New("tool failed")}
	registry := tooling.NewRegistry()
	registry.Register(tool)
	runner := registry.NewRunner(tooling.Config{Enabled: []string{tooling.SampleJudgeToolName}, MaxCalls: 1})
	client := &fakeAgentLLMClient{
		responses: []llm.GenerateResponse{
			{Model: "mock", Content: "```cpp\nint main(){return 0;}\n```", InputTokens: 10, OutputTokens: 20},
		},
	}
	strategy, err := ResolveSolveStrategy(ToolingCodegenV1AgentName)
	if err != nil {
		t.Fatalf("ResolveSolveStrategy returned error: %v", err)
	}

	output, err := strategy.Execute(context.Background(), client, SolveInput{
		Problem:       &model.Problem{BaseModel: model.BaseModel{ID: 1}},
		Model:         "mock",
		ToolingRunner: runner,
	})
	if err != nil {
		t.Fatalf("Execute returned error despite tool failure: %v", err)
	}
	if output.RawResponse == "" || output.ToolCallCount != 1 {
		t.Fatalf("expected generated code and one attempted tool call, got %+v", output)
	}
}
