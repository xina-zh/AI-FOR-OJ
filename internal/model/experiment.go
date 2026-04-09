package model

type Experiment struct {
	BaseModel

	ProblemID uint              `gorm:"column:problem_id;not null;index:idx_experiment_problem" json:"problem_id"`
	Status    string            `gorm:"type:varchar(32);not null;index" json:"status"`
	Problem   Problem           `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"problem,omitempty"`
	Config    *ExperimentConfig `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"config,omitempty"`
	Runs      []ExperimentRun   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"runs,omitempty"`
}

func (Experiment) TableName() string {
	return "experiments"
}
