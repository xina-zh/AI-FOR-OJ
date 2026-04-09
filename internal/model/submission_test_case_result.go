package model

type SubmissionTestCaseResult struct {
	CreatedModel

	SubmissionID uint       `gorm:"column:submission_id;not null;index:idx_submission_case_submission" json:"submission_id"`
	TestCaseID   uint       `gorm:"column:test_case_id;not null;index:idx_submission_case_test_case" json:"test_case_id"`
	CaseIndex    int        `gorm:"column:case_index;not null" json:"case_index"`
	Verdict      string     `gorm:"type:varchar(32);not null;index" json:"verdict"`
	RuntimeMS    int        `gorm:"column:runtime_ms;not null" json:"runtime_ms"`
	Stdout       string     `gorm:"column:stdout;type:longtext" json:"stdout"`
	Stderr       string     `gorm:"column:stderr;type:longtext" json:"stderr"`
	ExitCode     int        `gorm:"column:exit_code;not null;default:0" json:"exit_code"`
	TimedOut     bool       `gorm:"column:timed_out;not null;default:false" json:"timed_out"`
	Submission   Submission `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"submission,omitempty"`
	TestCase     TestCase   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"test_case,omitempty"`
}

func (SubmissionTestCaseResult) TableName() string {
	return "submission_test_case_results"
}
