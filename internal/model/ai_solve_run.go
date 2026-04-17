package model

const (
	AISolveRunStatusRunning = "running"
	AISolveRunStatusSuccess = "success"
	AISolveRunStatusFailed  = "failed"
)

type AISolveRun struct {
	BaseModel

	ProblemID      uint             `gorm:"column:problem_id;not null;index" json:"problem_id"`
	Model          string           `gorm:"column:model;type:varchar(128);not null;default:'';index" json:"model"`
	PromptName     string           `gorm:"column:prompt_name;type:varchar(64);not null;default:'';index" json:"prompt_name"`
	AgentName      string           `gorm:"column:agent_name;type:varchar(64);not null;default:'';index" json:"agent_name"`
	AttemptCount   int              `gorm:"column:attempt_count;not null;default:0" json:"attempt_count"`
	FailureType    string           `gorm:"column:failure_type;type:varchar(32);not null;default:'';index" json:"failure_type"`
	StrategyPath   string           `gorm:"column:strategy_path;type:varchar(255);not null;default:'';index" json:"strategy_path"`
	PromptPreview  string           `gorm:"column:prompt_preview;type:text" json:"prompt_preview"`
	RawResponse    string           `gorm:"column:raw_response;type:longtext" json:"raw_response"`
	ExtractedCode  string           `gorm:"column:extracted_code;type:longtext" json:"extracted_code"`
	SubmissionID   *uint            `gorm:"column:submission_id;index" json:"submission_id,omitempty"`
	Verdict        string           `gorm:"column:verdict;type:varchar(32);not null;default:'';index" json:"verdict"`
	Status         string           `gorm:"column:status;type:varchar(32);not null;index" json:"status"`
	ErrorMessage   string           `gorm:"column:error_message;type:text" json:"error_message"`
	TokenInput     int64            `gorm:"column:token_input;not null;default:0" json:"token_input"`
	TokenOutput    int64            `gorm:"column:token_output;not null;default:0" json:"token_output"`
	LLMLatencyMS   int              `gorm:"column:llm_latency_ms;not null;default:0" json:"llm_latency_ms"`
	TotalLatencyMS int              `gorm:"column:total_latency_ms;not null;default:0" json:"total_latency_ms"`
	Attempts       []AISolveAttempt `gorm:"foreignKey:AISolveRunID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"attempts,omitempty"`
}

func (AISolveRun) TableName() string {
	return "ai_solve_runs"
}
