package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"ai-for-oj/internal/handler/dto"
	"ai-for-oj/internal/repository"
	"ai-for-oj/internal/service"
)

type ProblemHandler struct {
	service *service.ProblemService
}

func NewProblemHandler(service *service.ProblemService) *ProblemHandler {
	return &ProblemHandler{service: service}
}

func (h *ProblemHandler) Create(c *gin.Context) {
	var req dto.CreateProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	output, err := h.service.Create(c.Request.Context(), service.CreateProblemInput{
		Title:         req.Title,
		Description:   req.Description,
		InputSpec:     req.InputSpec,
		OutputSpec:    req.OutputSpec,
		Samples:       req.Samples,
		TimeLimitMS:   req.TimeLimitMS,
		MemoryLimitMB: req.MemoryLimitMB,
		Difficulty:    req.Difficulty,
		Tags:          req.Tags,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.ProblemResponse{
		ID:            output.ID,
		Title:         output.Title,
		Description:   output.Description,
		InputSpec:     output.InputSpec,
		OutputSpec:    output.OutputSpec,
		Samples:       output.Samples,
		TimeLimitMS:   output.TimeLimitMS,
		MemoryLimitMB: output.MemoryLimitMB,
		Difficulty:    output.Difficulty,
		Tags:          output.Tags,
	})
}

func (h *ProblemHandler) List(c *gin.Context) {
	outputs, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	resp := make([]dto.ProblemResponse, 0, len(outputs))
	for _, output := range outputs {
		resp = append(resp, dto.ProblemResponse{
			ID:            output.ID,
			Title:         output.Title,
			Description:   output.Description,
			InputSpec:     output.InputSpec,
			OutputSpec:    output.OutputSpec,
			Samples:       output.Samples,
			TimeLimitMS:   output.TimeLimitMS,
			MemoryLimitMB: output.MemoryLimitMB,
			Difficulty:    output.Difficulty,
			Tags:          output.Tags,
		})
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ProblemHandler) Get(c *gin.Context) {
	problemID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	output, err := h.service.Get(c.Request.Context(), problemID)
	if err != nil {
		handleProblemError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ProblemResponse{
		ID:            output.ID,
		Title:         output.Title,
		Description:   output.Description,
		InputSpec:     output.InputSpec,
		OutputSpec:    output.OutputSpec,
		Samples:       output.Samples,
		TimeLimitMS:   output.TimeLimitMS,
		MemoryLimitMB: output.MemoryLimitMB,
		Difficulty:    output.Difficulty,
		Tags:          output.Tags,
	})
}

func (h *ProblemHandler) Delete(c *gin.Context) {
	problemID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), problemID); err != nil {
		handleProblemError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ProblemHandler) CreateTestCase(c *gin.Context) {
	problemID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.CreateTestCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	output, err := h.service.CreateTestCase(c.Request.Context(), service.CreateTestCaseInput{
		ProblemID:      problemID,
		Input:          req.Input,
		ExpectedOutput: req.ExpectedOutput,
		IsSample:       req.IsSample,
	})
	if err != nil {
		handleProblemError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.TestCaseResponse{
		ID:             output.ID,
		ProblemID:      output.ProblemID,
		Input:          output.Input,
		ExpectedOutput: output.ExpectedOutput,
		IsSample:       output.IsSample,
	})
}

func (h *ProblemHandler) ListTestCases(c *gin.Context) {
	problemID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	outputs, err := h.service.ListTestCases(c.Request.Context(), problemID)
	if err != nil {
		handleProblemError(c, err)
		return
	}

	resp := make([]dto.TestCaseResponse, 0, len(outputs))
	for _, output := range outputs {
		resp = append(resp, dto.TestCaseResponse{
			ID:             output.ID,
			ProblemID:      output.ProblemID,
			Input:          output.Input,
			ExpectedOutput: output.ExpectedOutput,
			IsSample:       output.IsSample,
		})
	}

	c.JSON(http.StatusOK, resp)
}

func handleProblemError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrProblemNotFound):
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
	}
}
