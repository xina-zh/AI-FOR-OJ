package sandbox

import "context"

type Compiler interface {
	Compile(ctx context.Context, req CompileRequest) (CompileResult, error)
}

type Runner interface {
	Run(ctx context.Context, req RunRequest) (RunResult, error)
}

type Executor interface {
	Compiler
	Runner
}

type Cleaner interface {
	Cleanup(ctx context.Context, artifactID string) error
}

type CompileRequest struct {
	Language   string
	SourceCode string
}

type CompileResult struct {
	Success      bool
	ArtifactID   string
	Stderr       string
	ExitCode     int
	ErrorMessage string
}

type RunRequest struct {
	Language      string
	ArtifactID    string
	Input         string
	TimeLimitMS   int
	MemoryLimitMB int
}

type RunResult struct {
	Stdout       string
	Stderr       string
	ExitCode     int
	RuntimeMS    int
	MemoryKB     int
	TimedOut     bool
	RuntimeError bool
	ErrorMessage string
}
