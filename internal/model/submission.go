package model

const (
	LanguageCPP17   = "cpp17"
	SourceTypeHuman = "human"
	SourceTypeAI    = "ai"
)

type Submission struct {
	BaseModel

	ProblemID       uint                       `gorm:"column:problem_id;not null;index:idx_submission_problem" json:"problem_id"`
	SourceCode      string                     `gorm:"column:source_code;type:longtext;not null" json:"source_code"`
	Language        string                     `gorm:"type:varchar(32);not null;index" json:"language"`
	SourceType      string                     `gorm:"column:source_type;type:varchar(16);not null;index" json:"source_type"`
	Problem         Problem                    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"problem,omitempty"`
	JudgeResult     *JudgeResult               `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"judge_result,omitempty"`
	TestCaseResults []SubmissionTestCaseResult `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"test_case_results,omitempty"`
}

func (Submission) TableName() string {
	return "submissions"
}
