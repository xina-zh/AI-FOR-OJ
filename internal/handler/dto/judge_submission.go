package dto

type JudgeSubmissionRequest struct {
	ProblemID  uint   `json:"problem_id" binding:"required"`
	SourceCode string `json:"source_code" binding:"required"`
	Language   string `json:"language" binding:"required"`
}

type JudgeSubmissionResponse struct {
	SubmissionID   uint   `json:"submission_id"`
	ProblemID      uint   `json:"problem_id"`
	Language       string `json:"language"`
	SourceType     string `json:"source_type"`
	Verdict        string `json:"verdict"`
	RuntimeMS      int    `json:"runtime_ms"`
	MemoryKB       int    `json:"memory_kb"`
	PassedCount    int    `json:"passed_count"`
	TotalCount     int    `json:"total_count"`
	MemoryExceeded bool   `json:"memory_exceeded"`
	ErrorMessage   string `json:"error_message,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
