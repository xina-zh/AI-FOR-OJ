package dto

type CreateProblemRequest struct {
	Title         string `json:"title" binding:"required"`
	Description   string `json:"description" binding:"required"`
	InputSpec     string `json:"input_spec" binding:"required"`
	OutputSpec    string `json:"output_spec" binding:"required"`
	Samples       string `json:"samples"`
	TimeLimitMS   int    `json:"time_limit_ms" binding:"required"`
	MemoryLimitMB int    `json:"memory_limit_mb" binding:"required"`
	Difficulty    string `json:"difficulty" binding:"required"`
	Tags          string `json:"tags"`
}

type ProblemResponse struct {
	ID            uint   `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	InputSpec     string `json:"input_spec"`
	OutputSpec    string `json:"output_spec"`
	Samples       string `json:"samples"`
	TimeLimitMS   int    `json:"time_limit_ms"`
	MemoryLimitMB int    `json:"memory_limit_mb"`
	Difficulty    string `json:"difficulty"`
	Tags          string `json:"tags"`
}

type CreateTestCaseRequest struct {
	Input          string `json:"input" binding:"required"`
	ExpectedOutput string `json:"expected_output" binding:"required"`
	IsSample       bool   `json:"is_sample"`
}

type TestCaseResponse struct {
	ID             uint   `json:"id"`
	ProblemID      uint   `json:"problem_id"`
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
	IsSample       bool   `json:"is_sample"`
}
