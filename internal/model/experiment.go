package model

type Experiment struct {
	BaseModel

	ProblemID     *uint             `gorm:"column:problem_id;index:idx_experiment_problem" json:"problem_id,omitempty"`
	Name          string            `gorm:"column:name;type:varchar(255);not null;default:''" json:"name"`
	ModelName     string            `gorm:"column:model_name;type:varchar(128);not null;default:'';index" json:"model_name"`
	PromptName    string            `gorm:"column:prompt_name;type:varchar(64);not null;default:'';index" json:"prompt_name"`
	AgentName     string            `gorm:"column:agent_name;type:varchar(64);not null;default:'';index" json:"agent_name"`
	ToolingConfig string            `gorm:"column:tooling_config;type:varchar(2048);not null;default:'{}'" json:"tooling_config"`
	Status        string            `gorm:"type:varchar(32);not null;index" json:"status"`
	TotalCount    int               `gorm:"column:total_count;not null;default:0" json:"total_count"`
	SuccessCount  int               `gorm:"column:success_count;not null;default:0" json:"success_count"`
	ACCount       int               `gorm:"column:ac_count;not null;default:0" json:"ac_count"`
	FailedCount   int               `gorm:"column:failed_count;not null;default:0" json:"failed_count"`
	Problem       *Problem          `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"problem,omitempty"`
	Config        *ExperimentConfig `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"config,omitempty"`
	Runs          []ExperimentRun   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"runs,omitempty"`
}

func (Experiment) TableName() string {
	return "experiments"
}
