package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"ai-for-oj/internal/agent"
	"ai-for-oj/internal/handler/dto"
	"ai-for-oj/internal/model"
	"ai-for-oj/internal/prompt"
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
		ProblemID:     req.ProblemID,
		Model:         req.Model,
		PromptName:    req.PromptName,
		AgentName:     req.AgentName,
		ToolingConfig: req.ToolingConfig,
	})
	if err != nil {
		errorResp := dto.AISolveErrorResponse{Error: err.Error()}
		runID := uint(0)
		if output != nil {
			runID = output.AISolveRunID
			errorResp.AISolveRunID = runID
			errorResp.PromptName = output.PromptName
			errorResp.AgentName = output.AgentName
			errorResp.TokenInput = output.TokenInput
			errorResp.TokenOutput = output.TokenOutput
			errorResp.LLMLatencyMS = output.LLMLatencyMS
			errorResp.TotalLatencyMS = output.TotalLatencyMS
			errorResp.AttemptCount = output.AttemptCount
			errorResp.FailureType = output.FailureType
			errorResp.StrategyPath = output.StrategyPath
			errorResp.ToolingConfig = output.ToolingConfig
			errorResp.ToolCallCount = output.ToolCallCount
		}
		switch {
		case errors.Is(err, agent.ErrUnknownSolveAgent):
			c.JSON(http.StatusBadRequest, errorResp)
		case errors.Is(err, prompt.ErrUnknownSolvePrompt):
			c.JSON(http.StatusBadRequest, errorResp)
		case errors.Is(err, repository.ErrProblemNotFound):
			errorResp.AISolveRunID = runID
			c.JSON(http.StatusNotFound, errorResp)
		case errors.Is(err, service.ErrAISolveLLMFailed), errors.Is(err, service.ErrAISolveCodeNotExtracted):
			errorResp.AISolveRunID = runID
			c.JSON(http.StatusBadGateway, errorResp)
		default:
			errorResp.AISolveRunID = runID
			c.JSON(http.StatusInternalServerError, errorResp)
		}
		return
	}

	c.JSON(http.StatusCreated, dto.AISolveResponse{
		AISolveRunID:   output.AISolveRunID,
		ProblemID:      output.ProblemID,
		Model:          output.Model,
		PromptName:     output.PromptName,
		AgentName:      output.AgentName,
		PromptPreview:  output.PromptPreview,
		RawResponse:    output.RawResponse,
		ExtractedCode:  output.ExtractedCode,
		SubmissionID:   output.SubmissionID,
		Verdict:        output.Verdict,
		ErrorMessage:   output.ErrorMessage,
		AttemptCount:   output.AttemptCount,
		FailureType:    output.FailureType,
		StrategyPath:   output.StrategyPath,
		ToolingConfig:  output.ToolingConfig,
		ToolCallCount:  output.ToolCallCount,
		TokenInput:     output.TokenInput,
		TokenOutput:    output.TokenOutput,
		LLMLatencyMS:   output.LLMLatencyMS,
		TotalLatencyMS: output.TotalLatencyMS,
		Attempts:       toAISolveAttemptResponses(output.Attempts),
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
		ID:             run.ID,
		ProblemID:      run.ProblemID,
		Model:          run.Model,
		PromptName:     prompt.DisplaySolvePromptName(run.PromptName),
		AgentName:      agent.DisplaySolveAgentName(run.AgentName),
		PromptPreview:  run.PromptPreview,
		RawResponse:    run.RawResponse,
		ExtractedCode:  run.ExtractedCode,
		SubmissionID:   run.SubmissionID,
		Verdict:        run.Verdict,
		Status:         run.Status,
		ErrorMessage:   run.ErrorMessage,
		AttemptCount:   run.AttemptCount,
		FailureType:    run.FailureType,
		StrategyPath:   run.StrategyPath,
		ToolingConfig:  run.ToolingConfig,
		ToolCallCount:  run.ToolCallCount,
		TokenInput:     run.TokenInput,
		TokenOutput:    run.TokenOutput,
		LLMLatencyMS:   run.LLMLatencyMS,
		TotalLatencyMS: run.TotalLatencyMS,
		CreatedAt:      run.CreatedAt,
		UpdatedAt:      run.UpdatedAt,
		Attempts:       toModelAISolveAttemptResponses(run.Attempts),
	})
}

func toAISolveAttemptResponses(attempts []service.AISolveAttemptOutput) []dto.AISolveAttemptResponse {
	if len(attempts) == 0 {
		return nil
	}
	responses := make([]dto.AISolveAttemptResponse, 0, len(attempts))
	for _, attempt := range attempts {
		responses = append(responses, dto.AISolveAttemptResponse{
			ID:               attempt.ID,
			AttemptNo:        attempt.AttemptNo,
			Stage:            attempt.Stage,
			Model:            attempt.Model,
			Verdict:          attempt.Verdict,
			FailureType:      attempt.FailureType,
			RepairReason:     attempt.RepairReason,
			StrategyPath:     attempt.StrategyPath,
			PromptPreview:    attempt.PromptPreview,
			ExtractedCode:    attempt.ExtractedCode,
			JudgePassedCount: attempt.JudgePassedCount,
			JudgeTotalCount:  attempt.JudgeTotalCount,
			TimedOut:         attempt.TimedOut,
			ErrorMessage:     attempt.ErrorMessage,
			TokenInput:       attempt.TokenInput,
			TokenOutput:      attempt.TokenOutput,
			LLMLatencyMS:     attempt.LLMLatencyMS,
			TotalLatencyMS:   attempt.TotalLatencyMS,
		})
	}
	return responses
}

func toModelAISolveAttemptResponses(attempts []model.AISolveAttempt) []dto.AISolveAttemptResponse {
	if len(attempts) == 0 {
		return nil
	}
	responses := make([]dto.AISolveAttemptResponse, 0, len(attempts))
	for _, attempt := range attempts {
		responses = append(responses, dto.AISolveAttemptResponse{
			ID:               attempt.ID,
			AttemptNo:        attempt.AttemptNo,
			Stage:            attempt.Stage,
			Model:            attempt.Model,
			Verdict:          attempt.JudgeVerdict,
			FailureType:      attempt.FailureType,
			RepairReason:     attempt.RepairReason,
			StrategyPath:     attempt.StrategyPath,
			PromptPreview:    attempt.PromptPreview,
			ExtractedCode:    attempt.ExtractedCode,
			JudgePassedCount: attempt.JudgePassedCount,
			JudgeTotalCount:  attempt.JudgeTotalCount,
			TimedOut:         attempt.TimedOut,
			ErrorMessage:     attempt.ErrorMessage,
			TokenInput:       attempt.TokenInput,
			TokenOutput:      attempt.TokenOutput,
			LLMLatencyMS:     attempt.LLMLatencyMS,
			TotalLatencyMS:   attempt.TotalLatencyMS,
		})
	}
	return responses
}
