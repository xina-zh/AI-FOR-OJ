package model

type AISolveAttempt struct {
	BaseModel

	AISolveRunID     uint   `gorm:"column:ai_solve_run_id;not null;index" json:"ai_solve_run_id"`
	AttemptNo        int    `gorm:"column:attempt_no;not null;index" json:"attempt_no"`
	Stage            string `gorm:"column:stage;type:varchar(32);not null;default:'';index" json:"stage"`
	FailureType      string `gorm:"column:failure_type;type:varchar(32);not null;default:'';index" json:"failure_type"`
	RepairReason     string `gorm:"column:repair_reason;type:text" json:"repair_reason"`
	StrategyPath     string `gorm:"column:strategy_path;type:varchar(255);not null;default:'';index" json:"strategy_path"`
	PromptPreview    string `gorm:"column:prompt_preview;type:text" json:"prompt_preview"`
	RawResponse      string `gorm:"column:raw_response;type:longtext" json:"raw_response"`
	ExtractedCode    string `gorm:"column:extracted_code;type:longtext" json:"extracted_code"`
	JudgeVerdict     string `gorm:"column:judge_verdict;type:varchar(32);not null;default:'';index" json:"judge_verdict"`
	JudgePassedCount int    `gorm:"column:judge_passed_count;not null;default:0" json:"judge_passed_count"`
	JudgeTotalCount  int    `gorm:"column:judge_total_count;not null;default:0" json:"judge_total_count"`
	TimedOut         bool   `gorm:"column:timed_out;not null;default:false" json:"timed_out"`
	CompileStderr    string `gorm:"column:compile_stderr;type:longtext" json:"compile_stderr"`
	RunStderr        string `gorm:"column:run_stderr;type:longtext" json:"run_stderr"`
	RunStdout        string `gorm:"column:run_stdout;type:longtext" json:"run_stdout"`
	ErrorMessage     string `gorm:"column:error_message;type:text" json:"error_message"`
	TokenInput       int64  `gorm:"column:token_input;not null;default:0" json:"token_input"`
	TokenOutput      int64  `gorm:"column:token_output;not null;default:0" json:"token_output"`
	LLMLatencyMS     int    `gorm:"column:llm_latency_ms;not null;default:0" json:"llm_latency_ms"`
}

func (AISolveAttempt) TableName() string {
	return "ai_solve_attempts"
}
