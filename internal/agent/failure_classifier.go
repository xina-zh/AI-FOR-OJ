package agent

import "strings"

type FailureType string

const (
	FailureTypeUnknown      FailureType = "unknown"
	FailureTypeWrongAnswer  FailureType = "wrong_answer"
	FailureTypeRuntimeError FailureType = "runtime_error"
	FailureTypeTimeLimit    FailureType = "time_limit"
)

type FailureObservation struct {
	Verdict       string
	TimedOut      bool
	CompileStderr string
	RunStderr     string
	PassedCount   int
	TotalCount    int
	ExecStage     string
}

func ClassifyFailure(observation FailureObservation) FailureType {
	verdict := strings.ToUpper(strings.TrimSpace(observation.Verdict))
	stage := strings.ToLower(strings.TrimSpace(observation.ExecStage))

	switch {
	case observation.TimedOut, verdict == "TLE", verdict == "TIMEDOUT":
		return FailureTypeTimeLimit
	case verdict == "WA":
		return FailureTypeWrongAnswer
	case verdict == "RE":
		return FailureTypeRuntimeError
	case verdict == "CE":
		return FailureTypeUnknown
	case stage == "run" && observation.TotalCount > 0 && observation.PassedCount > 0 && observation.PassedCount < observation.TotalCount:
		return FailureTypeWrongAnswer
	case stage == "compile" && strings.TrimSpace(observation.CompileStderr) != "":
		return FailureTypeUnknown
	case stage == "run" && strings.TrimSpace(observation.RunStderr) != "":
		return FailureTypeUnknown
	default:
		return FailureTypeUnknown
	}
}
