package sandbox

import (
	"context"
	"strings"
)

const (
	mockCompileErrorMarker        = "MOCK_CE"
	mockTimeLimitMarker           = "MOCK_TLE"
	mockMemoryLimitExceededMarker = "MOCK_MLE"
	mockRuntimeErrorMarker        = "MOCK_RE"
	mockWrongAnswerMarker         = "MOCK_WA"
)

type MockSandbox struct {
	sourceCodeByArtifact map[string]string
}

func NewMockSandbox() *MockSandbox {
	return &MockSandbox{
		sourceCodeByArtifact: make(map[string]string),
	}
}

func (m *MockSandbox) Compile(_ context.Context, req CompileRequest) (CompileResult, error) {
	if strings.Contains(req.SourceCode, mockCompileErrorMarker) {
		return CompileResult{
			Success:      false,
			Stderr:       "mock compile error triggered by source marker",
			ExitCode:     1,
			ErrorMessage: "mock compile error triggered by source marker",
		}, nil
	}

	artifactID := buildArtifactID(req.SourceCode)
	m.sourceCodeByArtifact[artifactID] = req.SourceCode

	return CompileResult{
		Success:    true,
		ArtifactID: artifactID,
	}, nil
}

func (m *MockSandbox) Run(_ context.Context, req RunRequest) (RunResult, error) {
	sourceCode := m.sourceCodeByArtifact[req.ArtifactID]

	switch {
	case strings.Contains(sourceCode, mockTimeLimitMarker):
		return RunResult{
			RuntimeMS: req.TimeLimitMS,
			MemoryKB:  0,
			TimedOut:  true,
			ExitCode:  -1,
		}, nil
	case strings.Contains(sourceCode, mockMemoryLimitExceededMarker):
		return RunResult{
			RuntimeMS:      1,
			MemoryKB:       req.MemoryLimitMB * 1024,
			ExitCode:       137,
			MemoryExceeded: true,
			ErrorMessage:   "memory limit exceeded",
		}, nil
	case strings.Contains(sourceCode, mockRuntimeErrorMarker):
		return RunResult{
			RuntimeMS:    1,
			MemoryKB:     1024,
			Stderr:       "mock runtime error triggered by source marker",
			ExitCode:     1,
			RuntimeError: true,
			ErrorMessage: "mock runtime error triggered by source marker",
		}, nil
	case strings.Contains(sourceCode, mockWrongAnswerMarker):
		return RunResult{
			Stdout:    "mock-wrong-answer",
			RuntimeMS: 1,
			MemoryKB:  1024,
			ExitCode:  0,
		}, nil
	default:
		return RunResult{
			Stdout:    req.Input,
			RuntimeMS: 1,
			MemoryKB:  1024,
			ExitCode:  0,
		}, nil
	}
}

func buildArtifactID(sourceCode string) string {
	return sourceCode
}
