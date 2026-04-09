package model

const (
	AISolveRunStatusRunning = "running"
	AISolveRunStatusSuccess = "success"
	AISolveRunStatusFailed  = "failed"
)

type AISolveRun struct {
	BaseModel

	ProblemID     uint   `gorm:"column:problem_id;not null;index" json:"problem_id"`
	Model         string `gorm:"column:model;type:varchar(128);not null;default:'';index" json:"model"`
	PromptPreview string `gorm:"column:prompt_preview;type:text" json:"prompt_preview"`
	RawResponse   string `gorm:"column:raw_response;type:longtext" json:"raw_response"`
	ExtractedCode string `gorm:"column:extracted_code;type:longtext" json:"extracted_code"`
	SubmissionID  *uint  `gorm:"column:submission_id;index" json:"submission_id,omitempty"`
	Verdict       string `gorm:"column:verdict;type:varchar(32);not null;default:'';index" json:"verdict"`
	Status        string `gorm:"column:status;type:varchar(32);not null;index" json:"status"`
	ErrorMessage  string `gorm:"column:error_message;type:text" json:"error_message"`
}

func (AISolveRun) TableName() string {
	return "ai_solve_runs"
}
