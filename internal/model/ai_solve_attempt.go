package model

type AISolveAttempt struct {
	CreatedModel

	AISolveRunID   uint       `gorm:"column:ai_solve_run_id;not null;index;uniqueIndex:idx_ai_solve_attempt" json:"ai_solve_run_id"`
	AttemptNo      int        `gorm:"column:attempt_no;not null;uniqueIndex:idx_ai_solve_attempt" json:"attempt_no"`
	Stage          string     `gorm:"column:stage;type:varchar(64);not null" json:"stage"`
	Model          string     `gorm:"column:model;type:varchar(128);not null;default:''" json:"model"`
	PromptPreview  string     `gorm:"column:prompt_preview;type:longtext" json:"prompt_preview"`
	RawResponse    string     `gorm:"column:raw_response;type:longtext" json:"raw_response"`
	ExtractedCode  string     `gorm:"column:extracted_code;type:longtext" json:"extracted_code"`
	Verdict        string     `gorm:"column:verdict;type:varchar(32);not null;default:''" json:"verdict"`
	FailureType    string     `gorm:"column:failure_type;type:varchar(64);not null;default:''" json:"failure_type"`
	RepairReason   string     `gorm:"column:repair_reason;type:longtext" json:"repair_reason"`
	TokenInput     int64      `gorm:"column:token_input;not null;default:0" json:"token_input"`
	TokenOutput    int64      `gorm:"column:token_output;not null;default:0" json:"token_output"`
	LLMLatencyMS   int        `gorm:"column:llm_latency_ms;not null;default:0" json:"llm_latency_ms"`
	TotalLatencyMS int        `gorm:"column:total_latency_ms;not null;default:0" json:"total_latency_ms"`
	AISolveRun     AISolveRun `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

func (AISolveAttempt) TableName() string {
	return "ai_solve_attempts"
}
