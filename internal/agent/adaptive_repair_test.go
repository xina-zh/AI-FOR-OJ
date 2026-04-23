package agent

import (
	"slices"
	"testing"
)

func TestAdaptiveRepairAgentRegistration(t *testing.T) {
	resolved, err := ResolveSolveAgentName("adaptive_repair_v1")
	if err != nil {
		t.Fatalf("ResolveSolveAgentName returned error: %v", err)
	}
	if resolved != AdaptiveRepairV1AgentName {
		t.Fatalf("expected %s, got %s", AdaptiveRepairV1AgentName, resolved)
	}
	if !slices.Contains(ListSolveAgents(), AdaptiveRepairV1AgentName) {
		t.Fatalf("expected ListSolveAgents to include %s", AdaptiveRepairV1AgentName)
	}
}

func TestFailureClassifierClassifiesRepairableVerdicts(t *testing.T) {
	classifier := FailureClassifier{}
	tests := []struct {
		name        string
		verdict     string
		failureType string
	}{
		{name: "wrong answer", verdict: "WA", failureType: FailureTypeWrongAnswer},
		{name: "runtime error", verdict: "RE", failureType: FailureTypeRuntimeError},
		{name: "time limit", verdict: "TLE", failureType: FailureTypeTimeLimit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classification := classifier.Classify(tt.verdict)
			if classification.FailureType != tt.failureType {
				t.Fatalf("expected failure type %s, got %+v", tt.failureType, classification)
			}
			if !classification.Repairable {
				t.Fatalf("expected verdict %s to be repairable", tt.verdict)
			}
		})
	}
}

func TestRepairPlannerSelectsVerdictSpecificStages(t *testing.T) {
	planner := RepairPlanner{MaxAttempts: 3}
	tests := []struct {
		name  string
		input FailureClassification
		stage string
	}{
		{
			name:  "wrong answer",
			input: FailureClassification{Verdict: "WA", FailureType: FailureTypeWrongAnswer, Repairable: true},
			stage: StageWAAnalysisRepair,
		},
		{
			name:  "runtime error",
			input: FailureClassification{Verdict: "RE", FailureType: FailureTypeRuntimeError, Repairable: true},
			stage: StageRESafetyRepair,
		},
		{
			name:  "time limit",
			input: FailureClassification{Verdict: "TLE", FailureType: FailureTypeTimeLimit, Repairable: true},
			stage: StageTLEComplexityRewrite,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan, ok := planner.NextRepair(1, tt.input)
			if !ok {
				t.Fatal("expected next repair")
			}
			if plan.Stage != tt.stage {
				t.Fatalf("expected stage %s, got %+v", tt.stage, plan)
			}
			if plan.FailureType != tt.input.FailureType {
				t.Fatalf("expected failure type %s, got %+v", tt.input.FailureType, plan)
			}
		})
	}
}

func TestRepairPlannerStopsWhenAttemptsExhausted(t *testing.T) {
	planner := RepairPlanner{MaxAttempts: 3}

	_, ok := planner.NextRepair(3, FailureClassification{
		Verdict:     "WA",
		FailureType: FailureTypeWrongAnswer,
		Repairable:  true,
	})
	if ok {
		t.Fatal("expected no next repair after attempts are exhausted")
	}
}
