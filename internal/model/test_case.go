package model

type TestCase struct {
	CreatedModel

	ProblemID      uint    `gorm:"column:problem_id;not null;index:idx_test_case_problem;index:idx_test_case_problem_sample" json:"problem_id"`
	Input          string  `gorm:"type:longtext;not null" json:"input"`
	ExpectedOutput string  `gorm:"column:expected_output;type:longtext;not null" json:"expected_output"`
	IsSample       bool    `gorm:"column:is_sample;not null;index:idx_test_case_problem_sample" json:"is_sample"`
	Problem        Problem `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"problem,omitempty"`
}

func (TestCase) TableName() string {
	return "test_cases"
}
