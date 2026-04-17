package agent

const (
	RepairStageWAAnalysisRepair     = "wa_analysis_repair"
	RepairStageRESafetyRepair       = "re_safety_repair"
	RepairStageTLEComplexityRewrite = "tle_complexity_rewrite"
	RepairStageFallbackRewrite      = "fallback_rewrite"
)

type RepairPlanInput struct {
	AttemptCount   int
	LastFailure    FailureType
	PreviousStages []string
	MaxBudget      int
}

type RepairPlanDecision struct {
	Stage string
	Stop  bool
}

type RepairPlanner struct {
	maxBudget int
}

func NewRepairPlanner(maxBudget int) RepairPlanner {
	return RepairPlanner{maxBudget: maxBudget}
}

func (p RepairPlanner) Next(input RepairPlanInput) RepairPlanDecision {
	budget := input.MaxBudget
	if budget <= 0 {
		budget = p.maxBudget
	}
	if budget <= 0 || input.AttemptCount >= budget {
		return RepairPlanDecision{Stop: true}
	}

	stage := stageForFailure(input.LastFailure)
	if stage == "" {
		stage = RepairStageFallbackRewrite
	}

	if !containsStage(input.PreviousStages, stage) {
		return RepairPlanDecision{Stage: stage}
	}

	if stage != RepairStageFallbackRewrite && !containsStage(input.PreviousStages, RepairStageFallbackRewrite) {
		return RepairPlanDecision{Stage: RepairStageFallbackRewrite}
	}

	return RepairPlanDecision{Stop: true}
}

func stageForFailure(failure FailureType) string {
	switch failure {
	case FailureTypeWrongAnswer:
		return RepairStageWAAnalysisRepair
	case FailureTypeRuntimeError:
		return RepairStageRESafetyRepair
	case FailureTypeTimeLimit:
		return RepairStageTLEComplexityRewrite
	default:
		return ""
	}
}

func containsStage(stages []string, target string) bool {
	for _, stage := range stages {
		if stage == target {
			return true
		}
	}
	return false
}
