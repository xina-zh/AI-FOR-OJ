package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"ai-for-oj/internal/handler/dto"
	"ai-for-oj/internal/repository"
	"ai-for-oj/internal/service"
)

type SubmissionHandler struct {
	judgeService *service.JudgeSubmissionService
	queryService *service.SubmissionQueryService
}

func NewSubmissionHandler(
	judgeService *service.JudgeSubmissionService,
	queryService *service.SubmissionQueryService,
) *SubmissionHandler {
	return &SubmissionHandler{
		judgeService: judgeService,
		queryService: queryService,
	}
}

func (h *SubmissionHandler) Judge(c *gin.Context) {
	var req dto.JudgeSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	output, err := h.judgeService.Submit(c.Request.Context(), service.JudgeSubmissionInput{
		ProblemID:  req.ProblemID,
		SourceCode: req.SourceCode,
		Language:   req.Language,
	})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrProblemNotFound):
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		case errors.Is(err, service.ErrUnsupportedLanguage):
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, dto.JudgeSubmissionResponse{
		SubmissionID:   output.SubmissionID,
		ProblemID:      output.ProblemID,
		Language:       output.Language,
		SourceType:     output.SourceType,
		Verdict:        output.Verdict,
		RuntimeMS:      output.RuntimeMS,
		MemoryKB:       output.MemoryKB,
		PassedCount:    output.PassedCount,
		TotalCount:     output.TotalCount,
		MemoryExceeded: output.MemoryExceeded,
		ErrorMessage:   output.ErrorMessage,
	})
}

func (h *SubmissionHandler) List(c *gin.Context) {
	page, ok := parsePositiveIntQuery(c, "page", 1)
	if !ok {
		return
	}
	pageSize, ok := parsePositiveIntQuery(c, "page_size", 20)
	if !ok {
		return
	}
	problemID, ok := parseOptionalUintQuery(c, "problem_id")
	if !ok {
		return
	}

	output, err := h.queryService.List(c.Request.Context(), service.SubmissionListInput{
		Page:      page,
		PageSize:  pageSize,
		ProblemID: problemID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	resp := make([]dto.SubmissionSummaryResponse, 0, len(output.Items))
	for _, item := range output.Items {
		resp = append(resp, dto.SubmissionSummaryResponse{
			ID:           item.ID,
			ProblemID:    item.ProblemID,
			ProblemTitle: item.ProblemTitle,
			Language:     item.Language,
			SourceType:   item.SourceType,
			Verdict:      item.Verdict,
			RuntimeMS:    item.RuntimeMS,
			PassedCount:  item.PassedCount,
			TotalCount:   item.TotalCount,
			CreatedAt:    item.CreatedAt,
			UpdatedAt:    item.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, dto.SubmissionListResponse{
		Items:      resp,
		Page:       output.Page,
		PageSize:   output.PageSize,
		Total:      output.Total,
		TotalPages: output.TotalPages,
	})
}

func (h *SubmissionHandler) AggregateByProblem(c *gin.Context) {
	outputs, err := h.queryService.AggregateByProblem(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	resp := make([]dto.SubmissionProblemStatsResponse, 0, len(outputs))
	for _, output := range outputs {
		resp = append(resp, dto.SubmissionProblemStatsResponse{
			ProblemID:          output.ProblemID,
			ProblemTitle:       output.ProblemTitle,
			TotalSubmissions:   output.TotalSubmissions,
			ACCount:            output.ACCount,
			WACount:            output.WACount,
			CECount:            output.CECount,
			RECount:            output.RECount,
			TLECount:           output.TLECount,
			LatestSubmissionAt: output.LatestSubmissionAt,
		})
	}

	c.JSON(http.StatusOK, resp)
}

func (h *SubmissionHandler) Get(c *gin.Context) {
	submissionID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	output, err := h.queryService.Get(c.Request.Context(), submissionID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrSubmissionNotFound):
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		}
		return
	}

	var judgeResult *dto.SubmissionJudgeResultResponse
	if output.JudgeResult != nil {
		judgeResult = &dto.SubmissionJudgeResultResponse{
			ID:             output.JudgeResult.ID,
			Verdict:        output.JudgeResult.Verdict,
			RuntimeMS:      output.JudgeResult.RuntimeMS,
			MemoryKB:       output.JudgeResult.MemoryKB,
			PassedCount:    output.JudgeResult.PassedCount,
			TotalCount:     output.JudgeResult.TotalCount,
			CompileStderr:  output.JudgeResult.CompileStderr,
			RunStdout:      output.JudgeResult.RunStdout,
			RunStderr:      output.JudgeResult.RunStderr,
			ExitCode:       output.JudgeResult.ExitCode,
			TimedOut:       output.JudgeResult.TimedOut,
			MemoryExceeded: output.JudgeResult.MemoryExceeded,
			ExecStage:      output.JudgeResult.ExecStage,
			ErrorMessage:   output.JudgeResult.ErrorMessage,
			CreatedAt:      output.JudgeResult.CreatedAt,
			UpdatedAt:      output.JudgeResult.UpdatedAt,
		}
	}

	testCaseResults := make([]dto.SubmissionTestCaseResultResponse, 0, len(output.TestCaseResults))
	for _, item := range output.TestCaseResults {
		testCaseResults = append(testCaseResults, dto.SubmissionTestCaseResultResponse{
			TestCaseID:     item.TestCaseID,
			CaseIndex:      item.CaseIndex,
			Verdict:        item.Verdict,
			RuntimeMS:      item.RuntimeMS,
			Stdout:         item.Stdout,
			Stderr:         item.Stderr,
			ExitCode:       item.ExitCode,
			TimedOut:       item.TimedOut,
			MemoryExceeded: item.MemoryExceeded,
		})
	}

	c.JSON(http.StatusOK, dto.SubmissionDetailResponse{
		ID:              output.ID,
		ProblemID:       output.ProblemID,
		ProblemTitle:    output.ProblemTitle,
		Language:        output.Language,
		SourceType:      output.SourceType,
		SourceCode:      output.SourceCode,
		Verdict:         output.Verdict,
		RuntimeMS:       output.RuntimeMS,
		MemoryKB:        output.MemoryKB,
		PassedCount:     output.PassedCount,
		TotalCount:      output.TotalCount,
		CompileStderr:   output.CompileStderr,
		RunStdout:       output.RunStdout,
		RunStderr:       output.RunStderr,
		ExitCode:        output.ExitCode,
		TimedOut:        output.TimedOut,
		MemoryExceeded:  output.MemoryExceeded,
		ExecStage:       output.ExecStage,
		ErrorMessage:    output.ErrorMessage,
		CreatedAt:       output.CreatedAt,
		UpdatedAt:       output.UpdatedAt,
		JudgeResult:     judgeResult,
		TestCaseResults: testCaseResults,
	})
}
