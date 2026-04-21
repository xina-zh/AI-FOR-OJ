package sandbox

import (
	"testing"
	"time"
)

func TestDockerSandboxRunTimeoutUsesProblemLimitAndBuffer(t *testing.T) {
	s := &DockerSandbox{
		runTimeoutBuffer: 500 * time.Millisecond,
	}

	got := s.runTimeout(1000)
	want := 1500 * time.Millisecond
	if got != want {
		t.Fatalf("expected timeout %s, got %s", want, got)
	}
}

func TestDockerSandboxRunTimeoutFallsBackToDefault(t *testing.T) {
	s := &DockerSandbox{
		runTimeoutBuffer: 250 * time.Millisecond,
	}

	got := s.runTimeout(0)
	want := 1250 * time.Millisecond
	if got != want {
		t.Fatalf("expected timeout %s, got %s", want, got)
	}
}

func TestDockerSandboxMemoryLimitArgFallsBackToDefault(t *testing.T) {
	s := &DockerSandbox{}

	got := s.memoryLimitArg(0)
	if got != "256m" {
		t.Fatalf("expected memory limit 256m, got %s", got)
	}
}

func TestDockerSandboxRunResultForOOM(t *testing.T) {
	result := runResultForOOM("Killed", 137, 64)
	if !result.MemoryExceeded || !result.RuntimeError || result.MemoryKB < 64*1024 {
		t.Fatalf("expected memory exceeded result, got %+v", result)
	}
}
