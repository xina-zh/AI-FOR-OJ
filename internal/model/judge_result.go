package model

type JudgeResult struct {
	BaseModel

	SubmissionID   uint       `gorm:"column:submission_id;not null;uniqueIndex" json:"submission_id"`
	Verdict        string     `gorm:"type:varchar(32);not null;index" json:"verdict"`
	RuntimeMS      int        `gorm:"column:runtime_ms;not null" json:"runtime_ms"`
	MemoryKB       int        `gorm:"column:memory_kb;not null" json:"memory_kb"`
	PassedCount    int        `gorm:"column:passed_count;not null" json:"passed_count"`
	TotalCount     int        `gorm:"column:total_count;not null" json:"total_count"`
	CompileStderr  string     `gorm:"column:compile_stderr;type:longtext" json:"compile_stderr"`
	RunStdout      string     `gorm:"column:run_stdout;type:longtext" json:"run_stdout"`
	RunStderr      string     `gorm:"column:run_stderr;type:longtext" json:"run_stderr"`
	ExitCode       int        `gorm:"column:exit_code;not null;default:0" json:"exit_code"`
	TimedOut       bool       `gorm:"column:timed_out;not null;default:false" json:"timed_out"`
	MemoryExceeded bool       `gorm:"column:memory_exceeded;not null;default:false" json:"memory_exceeded"`
	ExecStage      string     `gorm:"column:exec_stage;type:varchar(16);not null;default:''" json:"exec_stage"`
	ErrorMessage   string     `gorm:"column:error_message;type:text" json:"error_message"`
	Submission     Submission `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"submission,omitempty"`
}

func (JudgeResult) TableName() string {
	return "judge_results"
}
