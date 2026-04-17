package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"ai-for-oj/internal/model"
	"ai-for-oj/internal/repository"
	"ai-for-oj/internal/service"
)

type aiSolveRunHandlerRunRepository struct {
	run *model.AISolveRun
}

func (r *aiSolveRunHandlerRunRepository) Create(context.Context, *model.AISolveRun) error {
	return nil
}

func (r *aiSolveRunHandlerRunRepository) Update(context.Context, *model.AISolveRun) error {
	return nil
}

func (r *aiSolveRunHandlerRunRepository) GetByID(_ context.Context, runID uint) (*model.AISolveRun, error) {
	if r.run == nil || r.run.ID != runID {
		return nil, repository.ErrAISolveRunNotFound
	}
	return r.run, nil
}

type aiSolveRunResponse struct {
	ID           uint   `json:"id"`
	AttemptCount int    `json:"attempt_count"`
	FailureType  string `json:"failure_type"`
	StrategyPath string `json:"strategy_path"`
	Attempts     []struct {
		Stage        string `json:"stage"`
		Verdict      string `json:"verdict"`
		FailureType  string `json:"failure_type"`
		RepairReason string `json:"repair_reason"`
	} `json:"attempts"`
}

func TestAIHandlerGetRunIncludesAttempts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	runRepo := &aiSolveRunHandlerRunRepository{
		run: &model.AISolveRun{
			BaseModel:    model.BaseModel{ID: 23},
			AttemptCount: 3,
			FailureType:  "time_limit",
			StrategyPath: "analysis,repair",
			Attempts: []model.AISolveAttempt{
				{
					BaseModel:    model.BaseModel{ID: 1},
					AttemptNo:    1,
					Stage:        "analysis",
					FailureType:  "wrong_answer",
					RepairReason: "clarify the invariant",
					JudgeVerdict: "WA",
				},
				{
					BaseModel:    model.BaseModel{ID: 2},
					AttemptNo:    2,
					Stage:        "repair",
					FailureType:  "time_limit",
					RepairReason: "tighten edge cases",
					JudgeVerdict: "AC",
				},
			},
		},
	}
	aiService := service.NewAISolveService(nil, runRepo, nil, nil, "mock-model")
	aiHandler := NewAIHandler(aiService)

	router := gin.New()
	router.GET("/api/v1/ai/solve-runs/:id", aiHandler.GetRun)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai/solve-runs/23", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
	}

	var got aiSolveRunResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if got.AttemptCount != 3 || got.FailureType != "time_limit" || got.StrategyPath != "analysis,repair" {
		t.Fatalf("expected summary attempt metadata in response, got %+v", got)
	}
	if len(got.Attempts) != 2 {
		t.Fatalf("expected attempts in response, got %+v", got)
	}
	if got.Attempts[0].Stage != "analysis" || got.Attempts[0].Verdict != "WA" || got.Attempts[0].FailureType != "wrong_answer" || got.Attempts[0].RepairReason != "clarify the invariant" {
		t.Fatalf("unexpected first attempt payload: %+v", got.Attempts[0])
	}
	if got.Attempts[1].Stage != "repair" || got.Attempts[1].Verdict != "AC" || got.Attempts[1].FailureType != "time_limit" || got.Attempts[1].RepairReason != "tighten edge cases" {
		t.Fatalf("unexpected second attempt payload: %+v", got.Attempts[1])
	}
}
