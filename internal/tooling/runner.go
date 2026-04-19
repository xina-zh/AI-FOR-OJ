package tooling

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrToolDisabled     = errors.New("tool disabled")
	ErrToolNotFound     = errors.New("tool not found")
	ErrToolCallLimitHit = errors.New("tool call limit reached")
)

type Runner struct {
	config       Config
	tools        map[string]Tool
	callCount    int
	perToolCalls map[string]int
	results      []CallResult
}

func (r *Runner) Call(ctx context.Context, name string, input CallInput) (CallResult, error) {
	name = strings.TrimSpace(name)
	if name == "" || !r.config.Enabled(name) {
		return CallResult{}, ErrToolDisabled
	}
	if r.config.MaxCalls <= 0 || r.callCount >= r.config.MaxCalls {
		return CallResult{}, ErrToolCallLimitHit
	}
	if limit := r.config.LimitFor(name); limit > 0 && r.perToolCalls[name] >= limit {
		return CallResult{}, ErrToolCallLimitHit
	}
	tool, ok := r.tools[name]
	if !ok {
		return CallResult{}, fmt.Errorf("%w: %s", ErrToolNotFound, name)
	}

	result, err := tool.Execute(ctx, input)
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

func (r *Runner) Results() []CallResult {
	if r == nil {
		return nil
	}
	return append([]CallResult(nil), r.results...)
}
