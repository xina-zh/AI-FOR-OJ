package agent

import "strings"

type FailureType string

const (
	FailureTypeUnknown      FailureType = "unknown"
	FailureTypeWrongAnswer  FailureType = "wrong_answer"
	FailureTypeRuntimeError FailureType = "runtime_error"
	FailureTypeTimeLimit    FailureType = "time_limit"
)

// JudgeFailureObservation captures the judge output needed to classify a failure.
type JudgeFailureObservation struct {
	Verdict       string
	TimedOut      bool
	CompileStderr string
	RunStderr     string
	PassedCount   int
	TotalCount    int
	ExecStage     string
}

func ClassifyFailure(observation JudgeFailureObservation) FailureType {
	verdict := strings.ToUpper(strings.TrimSpace(observation.Verdict))

	switch {
	case observation.TimedOut, verdict == "TLE", verdict == "TIMEDOUT":
		return FailureTypeTimeLimit
	case verdict == "WA":
		return FailureTypeWrongAnswer
	case verdict == "RE":
		return FailureTypeRuntimeError
	default:
		return FailureTypeUnknown
	}
}
