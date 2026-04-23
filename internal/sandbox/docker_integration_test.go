package sandbox

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"ai-for-oj/internal/config"
	"ai-for-oj/internal/model"
)

func TestDockerSandboxDetectsOOMKilledIntegration(t *testing.T) {
	if os.Getenv("AI_FOR_OJ_DOCKER_INTEGRATION") != "1" {
		t.Skip("set AI_FOR_OJ_DOCKER_INTEGRATION=1 to run docker integration test")
	}

	s, err := NewDockerSandbox(config.SandboxConfig{
		WorkDir:          t.TempDir(),
		DockerImage:      "gcc:13",
		CompileTimeout:   10 * time.Second,
		RunTimeoutBuffer: 500 * time.Millisecond,
		CompileMemoryMB:  512,
	}, slog.Default())
	if err != nil {
		t.Fatalf("new docker sandbox: %v", err)
	}

	source := `#include <vector>
int main() {
  std::vector<char> data;
  while (true) data.resize(data.size() + 1024 * 1024, 1);
}`

	compileResult, err := s.Compile(context.Background(), CompileRequest{
		Language:   model.LanguageCPP17,
		SourceCode: source,
	})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if !compileResult.Success {
		t.Fatalf("compile failed: %+v", compileResult)
	}
	defer func() {
		_ = s.Cleanup(context.Background(), compileResult.ArtifactID)
	}()

	runResult, err := s.Run(context.Background(), RunRequest{
		Language:      model.LanguageCPP17,
		ArtifactID:    compileResult.ArtifactID,
		TimeLimitMS:   5000,
		MemoryLimitMB: 32,
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !runResult.MemoryExceeded {
		t.Fatalf("expected memory exceeded result, got %+v", runResult)
	}
}
