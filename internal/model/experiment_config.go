package model

type ExperimentConfig struct {
	CreatedModel

	ExperimentID  uint       `gorm:"column:experiment_id;not null;uniqueIndex" json:"experiment_id"`
	ModelName     string     `gorm:"column:model_name;type:varchar(128);not null;index" json:"model_name"`
	PromptVersion string     `gorm:"column:prompt_version;type:varchar(64);not null;index" json:"prompt_version"`
	AgentName     string     `gorm:"column:agent_name;type:varchar(64);not null;index" json:"agent_name"`
	ToolingConfig string     `gorm:"column:tooling_config;type:longtext;not null" json:"tooling_config"`
	MaxRounds     int        `gorm:"column:max_rounds;not null" json:"max_rounds"`
	Temperature   float64    `gorm:"type:decimal(4,2);not null" json:"temperature"`
	Experiment    Experiment `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"experiment,omitempty"`
}

func (ExperimentConfig) TableName() string {
	return "experiment_configs"
}
