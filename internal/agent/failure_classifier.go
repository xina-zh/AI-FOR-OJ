package agent

import "strings"

const (
	FailureTypeWrongAnswer  = "wrong_answer"
	FailureTypeRuntimeError = "runtime_error"
	FailureTypeTimeLimit    = "time_limit"
	FailureTypeMemoryLimit  = "memory_limit"
	FailureTypeCompileError = "compile_error"
	FailureTypeUnknown      = "unknown"
)

type FailureClassification struct {
	Verdict     string
	FailureType string
	Repairable  bool
}

type FailureClassifier struct{}

func (FailureClassifier) Classify(verdict string) FailureClassification {
	normalized := strings.ToUpper(strings.TrimSpace(verdict))
	switch normalized {
	case "WA":
		return FailureClassification{Verdict: normalized, FailureType: FailureTypeWrongAnswer, Repairable: true}
	case "RE":
		return FailureClassification{Verdict: normalized, FailureType: FailureTypeRuntimeError, Repairable: true}
	case "TLE":
		return FailureClassification{Verdict: normalized, FailureType: FailureTypeTimeLimit, Repairable: true}
	case "MLE":
		return FailureClassification{Verdict: normalized, FailureType: FailureTypeMemoryLimit, Repairable: true}
	case "CE":
		return FailureClassification{Verdict: normalized, FailureType: FailureTypeCompileError, Repairable: true}
	default:
		return FailureClassification{Verdict: normalized, FailureType: FailureTypeUnknown, Repairable: false}
	}
}
