package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"ai-for-oj/internal/handler/dto"
	"ai-for-oj/internal/repository"
	"ai-for-oj/internal/service"
)

type ExperimentHandler struct {
	service        *service.ExperimentService
	compareService *service.ExperimentCompareService
	repeatService  *service.ExperimentRepeatService
}

func NewExperimentHandler(
	service *service.ExperimentService,
	compareService *service.ExperimentCompareService,
	repeatService *service.ExperimentRepeatService,
) *ExperimentHandler {
	return &ExperimentHandler{
		service:        service,
		compareService: compareService,
		repeatService:  repeatService,
	}
}

func (h *ExperimentHandler) Run(c *gin.Context) {
	var req dto.RunExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	output, err := h.service.Run(c.Request.Context(), service.RunExperimentInput{
		Name:       req.Name,
		ProblemIDs: req.ProblemIDs,
		Model:      req.Model,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toExperimentResponse(output))
}

func (h *ExperimentHandler) Compare(c *gin.Context) {
	var req dto.CompareExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	output, err := h.compareService.Compare(c.Request.Context(), service.CompareExperimentInput{
		Name:           req.Name,
		ProblemIDs:     req.ProblemIDs,
		BaselineModel:  req.BaselineModel,
		CandidateModel: req.CandidateModel,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toExperimentCompareResponse(output))
}

func (h *ExperimentHandler) Repeat(c *gin.Context) {
	var req dto.RepeatExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	output, err := h.repeatService.Repeat(c.Request.Context(), service.RepeatExperimentInput{
		Name:        req.Name,
		ProblemIDs:  req.ProblemIDs,
		Model:       req.Model,
		RepeatCount: req.RepeatCount,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toExperimentRepeatResponse(output))
}

func (h *ExperimentHandler) Get(c *gin.Context) {
	experimentID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	output, err := h.service.Get(c.Request.Context(), experimentID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrExperimentNotFound):
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, toExperimentResponse(output))
}

func (h *ExperimentHandler) GetCompare(c *gin.Context) {
	compareID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	output, err := h.compareService.Get(c.Request.Context(), compareID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrExperimentCompareNotFound):
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, toExperimentCompareResponse(output))
}

func (h *ExperimentHandler) GetRepeat(c *gin.Context) {
	repeatID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	output, err := h.repeatService.Get(c.Request.Context(), repeatID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrExperimentRepeatNotFound):
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, toExperimentRepeatResponse(output))
}

func toExperimentResponse(output *service.ExperimentOutput) dto.ExperimentResponse {
	runs := make([]dto.ExperimentRunResponse, 0, len(output.Runs))
	for _, run := range output.Runs {
		runs = append(runs, dto.ExperimentRunResponse{
			ID:           run.ID,
			ProblemID:    run.ProblemID,
			AISolveRunID: run.AISolveRunID,
			SubmissionID: run.SubmissionID,
			AttemptNo:    run.AttemptNo,
			Verdict:      run.Verdict,
			Status:       run.Status,
			ErrorMessage: run.ErrorMessage,
			CreatedAt:    run.CreatedAt,
		})
	}

	return dto.ExperimentResponse{
		ID:                  output.ID,
		Name:                output.Name,
		Model:               output.Model,
		Status:              output.Status,
		TotalCount:          output.TotalCount,
		SuccessCount:        output.SuccessCount,
		ACCount:             output.ACCount,
		FailedCount:         output.FailedCount,
		VerdictDistribution: output.VerdictDistribution,
		CostSummary:         output.CostSummary,
		CreatedAt:           output.CreatedAt,
		UpdatedAt:           output.UpdatedAt,
		Runs:                runs,
	}
}

func toExperimentCompareResponse(output *service.ExperimentCompareOutput) dto.ExperimentCompareResponse {
	var baseline *dto.ExperimentResponse
	if output.BaselineSummary != nil {
		resp := toExperimentResponse(output.BaselineSummary)
		baseline = &resp
	}

	var candidate *dto.ExperimentResponse
	if output.CandidateSummary != nil {
		resp := toExperimentResponse(output.CandidateSummary)
		candidate = &resp
	}

	problems := make([]dto.ExperimentCompareProblemSummaryResponse, 0, len(output.ProblemSummaries))
	for _, problem := range output.ProblemSummaries {
		problems = append(problems, dto.ExperimentCompareProblemSummaryResponse{
			ProblemID:             problem.ProblemID,
			BaselineVerdict:       problem.BaselineVerdict,
			CandidateVerdict:      problem.CandidateVerdict,
			Changed:               problem.Changed,
			ChangeType:            problem.ChangeType,
			BaselineStatus:        problem.BaselineStatus,
			CandidateStatus:       problem.CandidateStatus,
			BaselineSubmissionID:  problem.BaselineSubmissionID,
			CandidateSubmissionID: problem.CandidateSubmissionID,
		})
	}

	highlighted := make([]dto.ExperimentCompareHighlightedProblemResponse, 0, len(output.HighlightedProblems))
	for _, problem := range output.HighlightedProblems {
		highlighted = append(highlighted, dto.ExperimentCompareHighlightedProblemResponse{
			ProblemID:             problem.ProblemID,
			BaselineVerdict:       problem.BaselineVerdict,
			CandidateVerdict:      problem.CandidateVerdict,
			Changed:               problem.Changed,
			ChangeType:            problem.ChangeType,
			BaselineSubmissionID:  problem.BaselineSubmissionID,
			CandidateSubmissionID: problem.CandidateSubmissionID,
		})
	}

	return dto.ExperimentCompareResponse{
		ID:                    output.ID,
		Name:                  output.Name,
		CompareDimension:      output.CompareDimension,
		BaselineValue:         output.BaselineValue,
		CandidateValue:        output.CandidateValue,
		ProblemIDs:            output.ProblemIDs,
		BaselineExperimentID:  output.BaselineExperimentID,
		CandidateExperimentID: output.CandidateExperimentID,
		BaselineSummary:       baseline,
		CandidateSummary:      candidate,
		BaselineDistribution:  output.BaselineDistribution,
		CandidateDistribution: output.CandidateDistribution,
		DeltaDistribution:     output.DeltaDistribution,
		CostComparison:        output.CostComparison,
		ImprovedCount:         output.ImprovedCount,
		RegressedCount:        output.RegressedCount,
		ChangedNonACCount:     output.ChangedNonACCount,
		ProblemSummaries:      problems,
		HighlightedProblems:   highlighted,
		DeltaACCount:          output.DeltaACCount,
		DeltaFailedCount:      output.DeltaFailedCount,
		Status:                output.Status,
		ErrorMessage:          output.ErrorMessage,
		CreatedAt:             output.CreatedAt,
		UpdatedAt:             output.UpdatedAt,
	}
}

func toExperimentRepeatResponse(output *service.ExperimentRepeatOutput) dto.ExperimentRepeatResponse {
	rounds := make([]dto.ExperimentRepeatRoundSummaryResponse, 0, len(output.RoundSummaries))
	for _, round := range output.RoundSummaries {
		rounds = append(rounds, dto.ExperimentRepeatRoundSummaryResponse{
			RoundNo:             round.RoundNo,
			ExperimentID:        round.ExperimentID,
			ACCount:             round.ACCount,
			FailedCount:         round.FailedCount,
			VerdictDistribution: round.VerdictDistribution,
			Status:              round.Status,
		})
	}

	problems := make([]dto.ExperimentRepeatProblemSummaryResponse, 0, len(output.ProblemSummaries))
	for _, problem := range output.ProblemSummaries {
		problems = append(problems, dto.ExperimentRepeatProblemSummaryResponse{
			ProblemID:           problem.ProblemID,
			TotalRounds:         problem.TotalRounds,
			ACCount:             problem.ACCount,
			FailedCount:         problem.FailedCount,
			ACRate:              problem.ACRate,
			VerdictDistribution: problem.VerdictDistribution,
			LatestVerdict:       problem.LatestVerdict,
		})
	}

	unstable := make([]dto.ExperimentRepeatUnstableProblemResponse, 0, len(output.MostUnstableProblems))
	for _, problem := range output.MostUnstableProblems {
		unstable = append(unstable, dto.ExperimentRepeatUnstableProblemResponse{
			ProblemID:           problem.ProblemID,
			TotalRounds:         problem.TotalRounds,
			ACCount:             problem.ACCount,
			FailedCount:         problem.FailedCount,
			ACRate:              problem.ACRate,
			VerdictDistribution: problem.VerdictDistribution,
			LatestVerdict:       problem.LatestVerdict,
			InstabilityScore:    problem.InstabilityScore,
			VerdictKindCount:    problem.VerdictKindCount,
		})
	}

	return dto.ExperimentRepeatResponse{
		ID:                         output.ID,
		Name:                       output.Name,
		Model:                      output.Model,
		ProblemIDs:                 output.ProblemIDs,
		RepeatCount:                output.RepeatCount,
		ExperimentIDs:              output.ExperimentIDs,
		TotalProblemCount:          output.TotalProblemCount,
		TotalRunCount:              output.TotalRunCount,
		OverallACCount:             output.OverallACCount,
		OverallFailedCount:         output.OverallFailedCount,
		AverageACCountPerRound:     output.AverageACCountPerRound,
		AverageFailedCountPerRound: output.AverageFailedCountPerRound,
		OverallACRate:              output.OverallACRate,
		BestRoundACCount:           output.BestRoundACCount,
		WorstRoundACCount:          output.WorstRoundACCount,
		CostSummary:                output.CostSummary,
		Status:                     output.Status,
		ErrorMessage:               output.ErrorMessage,
		RoundSummaries:             rounds,
		ProblemSummaries:           problems,
		MostUnstableProblems:       unstable,
		CreatedAt:                  output.CreatedAt,
		UpdatedAt:                  output.UpdatedAt,
	}
}
