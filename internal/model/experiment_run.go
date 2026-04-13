package model

type ExperimentRun struct {
	CreatedModel

	ExperimentID uint         `gorm:"column:experiment_id;not null;uniqueIndex:idx_run_attempt;index" json:"experiment_id"`
	ProblemID    uint         `gorm:"column:problem_id;not null;index" json:"problem_id"`
	AISolveRunID *uint        `gorm:"column:ai_solve_run_id;index" json:"ai_solve_run_id,omitempty"`
	SubmissionID *uint        `gorm:"column:submission_id;index" json:"submission_id,omitempty"`
	AttemptNo    int          `gorm:"column:attempt_no;not null;uniqueIndex:idx_run_attempt" json:"attempt_no"`
	FinalVerdict string       `gorm:"column:final_verdict;type:varchar(32);not null;default:'';index" json:"final_verdict"`
	Status       string       `gorm:"column:status;type:varchar(32);not null;index" json:"status"`
	ErrorMessage string       `gorm:"column:error_message;type:text" json:"error_message"`
	TokenInput   int64        `gorm:"column:token_input;not null" json:"token_input"`
	TokenOutput  int64        `gorm:"column:token_output;not null" json:"token_output"`
	LatencyMS    int          `gorm:"column:latency_ms;not null" json:"latency_ms"`
	ToolCalls    int          `gorm:"column:tool_calls;not null" json:"tool_calls"`
	AISolveRun   *AISolveRun  `gorm:"foreignKey:AISolveRunID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"ai_solve_run,omitempty"`
	Experiment   Experiment   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"experiment,omitempty"`
	Submission   *Submission  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"submission,omitempty"`
	TraceEvents  []TraceEvent `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"trace_events,omitempty"`
}

func (ExperimentRun) TableName() string {
	return "experiment_runs"
}
