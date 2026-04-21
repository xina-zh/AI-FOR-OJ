package model

const (
	ExperimentRepeatStatusRunning   = "running"
	ExperimentRepeatStatusCompleted = "completed"
	ExperimentRepeatStatusFailed    = "failed"
)

type ExperimentRepeat struct {
	BaseModel

	Name          string `gorm:"column:name;type:varchar(255);not null;default:''" json:"name"`
	ModelName     string `gorm:"column:model_name;type:varchar(128);not null;default:'';index" json:"model_name"`
	PromptName    string `gorm:"column:prompt_name;type:varchar(64);not null;default:'';index" json:"prompt_name"`
	AgentName     string `gorm:"column:agent_name;type:varchar(64);not null;default:'';index" json:"agent_name"`
	ToolingConfig string `gorm:"column:tooling_config;type:longtext;not null;default:'{}'" json:"tooling_config"`
	ProblemIDs    string `gorm:"column:problem_ids;type:longtext;not null" json:"problem_ids"`
	ExperimentIDs string `gorm:"column:experiment_ids;type:longtext;not null" json:"experiment_ids"`
	RepeatCount   int    `gorm:"column:repeat_count;not null;default:1" json:"repeat_count"`
	Status        string `gorm:"column:status;type:varchar(32);not null;index" json:"status"`
	ErrorMessage  string `gorm:"column:error_message;type:text" json:"error_message"`
}

func (ExperimentRepeat) TableName() string {
	return "experiment_repeats"
}
