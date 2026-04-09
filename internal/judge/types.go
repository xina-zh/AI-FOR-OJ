package judge

import (
	"context"

	"ai-for-oj/internal/model"
)

type Engine interface {
	Judge(ctx context.Context, req Request) (Result, error)
}

type Request struct {
	Problem    *model.Problem
	TestCases  []model.TestCase
	Language   string
	SourceCode string
}

type Result struct {
	Verdict         string
	RuntimeMS       int
	MemoryKB        int
	PassedCount     int
	TotalCount      int
	CompileStderr   string
	RunStdout       string
	RunStderr       string
	ExitCode        int
	TimedOut        bool
	ExecStage       string
	ErrorMessage    string
	TestCaseResults []TestCaseResult
}

type TestCaseResult struct {
	TestCaseID uint
	CaseIndex  int
	Verdict    string
	RuntimeMS  int
	Stdout     string
	Stderr     string
	ExitCode   int
	TimedOut   bool
}
