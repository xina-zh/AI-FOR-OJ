package agent

import "testing"

func TestResolveSolveAgentNameAdaptiveRepairV1(t *testing.T) {
	got, err := ResolveSolveAgentName("adaptive_repair_v1")
	if err != nil {
		t.Fatalf("ResolveSolveAgentName returned error: %v", err)
	}
	if got != "adaptive_repair_v1" {
		t.Fatalf("ResolveSolveAgentName = %q, want %q", got, "adaptive_repair_v1")
	}
}

func TestResolveSolveStrategyAdaptiveRepairV1(t *testing.T) {
	got, err := ResolveSolveStrategy("adaptive_repair_v1")
	if err != nil {
		t.Fatalf("ResolveSolveStrategy returned error: %v", err)
	}
	if got == nil {
		t.Fatal("ResolveSolveStrategy returned nil strategy")
	}
	if got.Name() != "adaptive_repair_v1" {
		t.Fatalf("ResolveSolveStrategy.Name() = %q, want %q", got.Name(), "adaptive_repair_v1")
	}
}

func TestSupportsSelfRepairDoesNotDriveAdaptiveRepair(t *testing.T) {
	if SupportsSelfRepair("adaptive_repair_v1") {
		t.Fatal("SupportsSelfRepair should not report adaptive_repair_v1 as self-repair capable")
	}
}
