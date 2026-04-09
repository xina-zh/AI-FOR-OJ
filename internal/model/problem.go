package model

type Problem struct {
	BaseModel

	Title         string       `gorm:"type:varchar(255);not null;index" json:"title"`
	Description   string       `gorm:"type:longtext;not null" json:"description"`
	InputSpec     string       `gorm:"column:input_spec;type:longtext;not null" json:"input_spec"`
	OutputSpec    string       `gorm:"column:output_spec;type:longtext;not null" json:"output_spec"`
	Samples       string       `gorm:"type:longtext;not null" json:"samples"`
	TimeLimitMS   int          `gorm:"column:time_limit_ms;not null" json:"time_limit_ms"`
	MemoryLimitMB int          `gorm:"column:memory_limit_mb;not null" json:"memory_limit_mb"`
	Difficulty    string       `gorm:"type:varchar(32);not null;index" json:"difficulty"`
	Tags          string       `gorm:"type:longtext;not null" json:"tags"`
	TestCases     []TestCase   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"test_cases,omitempty"`
	Submissions   []Submission `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"submissions,omitempty"`
	Experiments   []Experiment `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"experiments,omitempty"`
}

func (Problem) TableName() string {
	return "problems"
}
