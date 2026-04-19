package tooling

import (
	"context"

	"ai-for-oj/internal/model"
)

const (
	CallStatusOK      = "ok"
	CallStatusSkipped = "skipped"
	CallStatusFailed  = "failed"
)

type Tool interface {
	Name() string
	Execute(ctx context.Context, input CallInput) (CallResult, error)
}

type CallInput struct {
	Problem    *model.Problem
	ProblemID  uint
	SourceCode string
}

type CallResult struct {
	ToolName string `json:"tool_name"`
	Status   string `json:"status"`
	Summary  string `json:"summary,omitempty"`
	Metadata string `json:"metadata,omitempty"`
}
