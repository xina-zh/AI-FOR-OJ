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

func TestParseDockerOOMKilled(t *testing.T) {
	for _, tc := range []struct {
		name string
		raw  string
		want bool
	}{
		{name: "true", raw: "true\n", want: true},
		{name: "false", raw: "false\n", want: false},
		{name: "empty", raw: "", want: false},
		{name: "garbage", raw: "not-json", want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := parseDockerBool(tc.raw); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestDockerSandboxRunMemoryKBForOOMFallsBackToLimit(t *testing.T) {
	result := runResultForOOM("stderr", 137, 64)
	if !result.MemoryExceeded {
		t.Fatalf("expected memory exceeded result, got %+v", result)
	}
	if result.MemoryKB != 64*1024 {
		t.Fatalf("expected memory fallback to limit, got %d", result.MemoryKB)
	}
	if result.RuntimeError {
		t.Fatalf("MLE should not also be marked runtime error: %+v", result)
	}
}
