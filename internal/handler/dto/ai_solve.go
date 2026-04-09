package dto

import "time"

type AISolveRequest struct {
	ProblemID uint   `json:"problem_id" binding:"required"`
	Model     string `json:"model"`
}

type AISolveResponse struct {
	AISolveRunID   uint   `json:"ai_solve_run_id"`
	ProblemID      uint   `json:"problem_id"`
	Model          string `json:"model,omitempty"`
	PromptPreview  string `json:"prompt_preview"`
	RawResponse    string `json:"raw_response,omitempty"`
	ExtractedCode  string `json:"extracted_code,omitempty"`
	SubmissionID   uint   `json:"submission_id"`
	Verdict        string `json:"verdict,omitempty"`
	ErrorMessage   string `json:"error_message,omitempty"`
	TokenInput     int64  `json:"token_input"`
	TokenOutput    int64  `json:"token_output"`
	LLMLatencyMS   int    `json:"llm_latency_ms"`
	TotalLatencyMS int    `json:"total_latency_ms"`
}

type AISolveErrorResponse struct {
	Error          string `json:"error"`
	AISolveRunID   uint   `json:"ai_solve_run_id,omitempty"`
	TokenInput     int64  `json:"token_input"`
	TokenOutput    int64  `json:"token_output"`
	LLMLatencyMS   int    `json:"llm_latency_ms"`
	TotalLatencyMS int    `json:"total_latency_ms"`
}

type AISolveRunResponse struct {
	ID             uint      `json:"id"`
	ProblemID      uint      `json:"problem_id"`
	Model          string    `json:"model,omitempty"`
	PromptPreview  string    `json:"prompt_preview,omitempty"`
	RawResponse    string    `json:"raw_response,omitempty"`
	ExtractedCode  string    `json:"extracted_code,omitempty"`
	SubmissionID   *uint     `json:"submission_id,omitempty"`
	Verdict        string    `json:"verdict,omitempty"`
	Status         string    `json:"status"`
	ErrorMessage   string    `json:"error_message,omitempty"`
	TokenInput     int64     `json:"token_input"`
	TokenOutput    int64     `json:"token_output"`
	LLMLatencyMS   int       `json:"llm_latency_ms"`
	TotalLatencyMS int       `json:"total_latency_ms"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
