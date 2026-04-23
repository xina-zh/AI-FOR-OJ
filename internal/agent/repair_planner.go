package agent

const (
	StageInitialCodegen       = "initial_codegen"
	StageWAAnalysisRepair     = "wa_analysis_repair"
	StageRESafetyRepair       = "re_safety_repair"
	StageTLEComplexityRewrite = "tle_complexity_rewrite"
	StageFallbackRewrite      = "fallback_rewrite"
)

type RepairPlan struct {
	Stage        string
	FailureType  string
	RepairReason string
}

type RepairPlanner struct {
	MaxAttempts int
}

func (p RepairPlanner) NextRepair(completedAttempts int, classification FailureClassification) (RepairPlan, bool) {
	maxAttempts := p.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	if completedAttempts >= maxAttempts || !classification.Repairable {
		return RepairPlan{}, false
	}

	stage := StageFallbackRewrite
	switch classification.FailureType {
	case FailureTypeWrongAnswer:
		stage = StageWAAnalysisRepair
	case FailureTypeRuntimeError:
		stage = StageRESafetyRepair
	case FailureTypeTimeLimit:
		stage = StageTLEComplexityRewrite
	}

	return RepairPlan{
		Stage:        stage,
		FailureType:  classification.FailureType,
		RepairReason: classification.Verdict,
	}, true
}
