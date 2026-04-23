package model

const (
	ExperimentCompareStatusRunning   = "running"
	ExperimentCompareStatusCompleted = "completed"
	ExperimentCompareStatusFailed    = "failed"
)

type ExperimentCompare struct {
	BaseModel

	Name                   string `gorm:"column:name;type:varchar(255);not null;default:''" json:"name"`
	CompareDimension       string `gorm:"column:compare_dimension;type:varchar(32);not null;index" json:"compare_dimension"`
	BaselineValue          string `gorm:"column:baseline_value;type:varchar(128);not null;default:''" json:"baseline_value"`
	CandidateValue         string `gorm:"column:candidate_value;type:varchar(128);not null;default:''" json:"candidate_value"`
	BaselinePromptName     string `gorm:"column:baseline_prompt_name;type:varchar(64);not null;default:''" json:"baseline_prompt_name"`
	CandidatePromptName    string `gorm:"column:candidate_prompt_name;type:varchar(64);not null;default:''" json:"candidate_prompt_name"`
	BaselineAgentName      string `gorm:"column:baseline_agent_name;type:varchar(64);not null;default:''" json:"baseline_agent_name"`
	CandidateAgentName     string `gorm:"column:candidate_agent_name;type:varchar(64);not null;default:''" json:"candidate_agent_name"`
	BaselineToolingConfig  string `gorm:"column:baseline_tooling_config;type:varchar(2048);not null;default:'{}'" json:"baseline_tooling_config"`
	CandidateToolingConfig string `gorm:"column:candidate_tooling_config;type:varchar(2048);not null;default:'{}'" json:"candidate_tooling_config"`
	ProblemIDs             string `gorm:"column:problem_ids;type:longtext;not null" json:"problem_ids"`
	BaselineExperimentID   *uint  `gorm:"column:baseline_experiment_id;index" json:"baseline_experiment_id,omitempty"`
	CandidateExperimentID  *uint  `gorm:"column:candidate_experiment_id;index" json:"candidate_experiment_id,omitempty"`
	DeltaACCount           int    `gorm:"column:delta_ac_count;not null;default:0" json:"delta_ac_count"`
	DeltaFailedCount       int    `gorm:"column:delta_failed_count;not null;default:0" json:"delta_failed_count"`
	Status                 string `gorm:"column:status;type:varchar(32);not null;index" json:"status"`
	ErrorMessage           string `gorm:"column:error_message;type:text" json:"error_message"`
}

func (ExperimentCompare) TableName() string {
	return "experiment_compares"
}
