package tooling

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"ai-for-oj/internal/model"
)

const (
	CallStatusOK      = "ok"
	CallStatusSkipped = "skipped"
	CallStatusFailed  = "failed"
)

var (
	ErrToolDisabled     = errors.New("tool disabled")
	ErrToolNotFound     = errors.New("tool not found")
	ErrToolCallLimitHit = errors.New("tool call limit reached")
)

type Tool interface {
	Name() string
	Run(ctx context.Context, input CallInput) (CallOutput, error)
}

type CallInput struct {
	Problem    *model.Problem
	ProblemID  uint
	SourceCode string
	Metadata   map[string]any
}

type CallOutput struct {
	ToolName string         `json:"tool_name"`
	Status   string         `json:"status"`
	Summary  string         `json:"summary,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type Runner struct {
	config       Config
	registry     *Registry
	callCount    int
	perToolCalls map[string]int
	results      []CallOutput
}

func (r *Runner) Call(ctx context.Context, name string, input CallInput) (CallOutput, error) {
	if r == nil {
		return CallOutput{}, ErrToolDisabled
	}
	name = strings.TrimSpace(name)
	if name == "" || !r.config.EnabledTool(name) {
		return CallOutput{}, ErrToolDisabled
	}
	if r.config.MaxCalls <= 0 || r.callCount >= r.config.MaxCalls {
		return CallOutput{}, ErrToolCallLimitHit
	}
	if limit := r.config.LimitFor(name); limit > 0 && r.perToolCalls[name] >= limit {
		return CallOutput{}, ErrToolCallLimitHit
	}
	tool, ok := r.registry.Lookup(name)
	if !ok {
		return CallOutput{}, fmt.Errorf("%w: %s", ErrToolNotFound, name)
	}

	result, err := tool.Run(ctx, input)
	if result.ToolName == "" {
		result.ToolName = name
	}
	r.callCount++
	r.perToolCalls[name]++
	r.results = append(r.results, result)
	return result, err
}

func (r *Runner) CallCount() int {
	if r == nil {
		return 0
	}
	return r.callCount
}

func (r *Runner) Results() []CallOutput {
	if r == nil {
		return nil
	}
	return append([]CallOutput(nil), r.results...)
}
