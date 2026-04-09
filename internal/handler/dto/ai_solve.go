package dto

import "time"

type AISolveRequest struct {
	ProblemID uint   `json:"problem_id" binding:"required"`
	Model     string `json:"model"`
}

type AISolveResponse struct {
	AISolveRunID  uint   `json:"ai_solve_run_id"`
	ProblemID     uint   `json:"problem_id"`
	Model         string `json:"model,omitempty"`
	PromptPreview string `json:"prompt_preview"`
	RawResponse   string `json:"raw_response,omitempty"`
	ExtractedCode string `json:"extracted_code,omitempty"`
	SubmissionID  uint   `json:"submission_id"`
	Verdict       string `json:"verdict,omitempty"`
	ErrorMessage  string `json:"error_message,omitempty"`
}

type AISolveErrorResponse struct {
	Error        string `json:"error"`
	AISolveRunID uint   `json:"ai_solve_run_id,omitempty"`
}

type AISolveRunResponse struct {
	ID            uint      `json:"id"`
	ProblemID     uint      `json:"problem_id"`
	Model         string    `json:"model,omitempty"`
	PromptPreview string    `json:"prompt_preview,omitempty"`
	RawResponse   string    `json:"raw_response,omitempty"`
	ExtractedCode string    `json:"extracted_code,omitempty"`
	SubmissionID  *uint     `json:"submission_id,omitempty"`
	Verdict       string    `json:"verdict,omitempty"`
	Status        string    `json:"status"`
	ErrorMessage  string    `json:"error_message,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
