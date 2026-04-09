package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"ai-for-oj/internal/handler/dto"
	"ai-for-oj/internal/repository"
	"ai-for-oj/internal/service"
)

type AIHandler struct {
	service *service.AISolveService
}

func NewAIHandler(service *service.AISolveService) *AIHandler {
	return &AIHandler{service: service}
}

func (h *AIHandler) Solve(c *gin.Context) {
	var req dto.AISolveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	output, err := h.service.Solve(c.Request.Context(), service.AISolveInput{
		ProblemID: req.ProblemID,
		Model:     req.Model,
	})
	if err != nil {
		runID := uint(0)
		if output != nil {
			runID = output.AISolveRunID
		}
		switch {
		case errors.Is(err, repository.ErrProblemNotFound):
			c.JSON(http.StatusNotFound, dto.AISolveErrorResponse{Error: err.Error(), AISolveRunID: runID})
		case errors.Is(err, service.ErrAISolveLLMFailed), errors.Is(err, service.ErrAISolveCodeNotExtracted):
			c.JSON(http.StatusBadGateway, dto.AISolveErrorResponse{Error: err.Error(), AISolveRunID: runID})
		default:
			c.JSON(http.StatusInternalServerError, dto.AISolveErrorResponse{Error: err.Error(), AISolveRunID: runID})
		}
		return
	}

	c.JSON(http.StatusCreated, dto.AISolveResponse{
		AISolveRunID:  output.AISolveRunID,
		ProblemID:     output.ProblemID,
		Model:         output.Model,
		PromptPreview: output.PromptPreview,
		RawResponse:   output.RawResponse,
		ExtractedCode: output.ExtractedCode,
		SubmissionID:  output.SubmissionID,
		Verdict:       output.Verdict,
		ErrorMessage:  output.ErrorMessage,
	})
}

func (h *AIHandler) GetRun(c *gin.Context) {
	runID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	run, err := h.service.GetRun(c.Request.Context(), runID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrAISolveRunNotFound):
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, dto.AISolveRunResponse{
		ID:            run.ID,
		ProblemID:     run.ProblemID,
		Model:         run.Model,
		PromptPreview: run.PromptPreview,
		RawResponse:   run.RawResponse,
		ExtractedCode: run.ExtractedCode,
		SubmissionID:  run.SubmissionID,
		Verdict:       run.Verdict,
		Status:        run.Status,
		ErrorMessage:  run.ErrorMessage,
		CreatedAt:     run.CreatedAt,
		UpdatedAt:     run.UpdatedAt,
	})
}
