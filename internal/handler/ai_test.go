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
		AttemptNo        int    `json:"attempt_no"`
		Stage            string `json:"stage"`
		Verdict          string `json:"verdict"`
		FailureType      string `json:"failure_type"`
		RepairReason     string `json:"repair_reason"`
		StrategyPath     string `json:"strategy_path"`
		PromptPreview    string `json:"prompt_preview"`
		ExtractedCode    string `json:"extracted_code"`
		JudgePassedCount int    `json:"judge_passed_count"`
		JudgeTotalCount  int    `json:"judge_total_count"`
		TimedOut         bool   `json:"timed_out"`
		ErrorMessage     string `json:"error_message"`
		TokenInput       int64  `json:"token_input"`
		TokenOutput      int64  `json:"token_output"`
		LLMLatencyMS     int    `json:"llm_latency_ms"`
	} `json:"attempts"`
}

func TestAIHandlerGetRunIncludesAttemptDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	runRepo := &aiSolveRunHandlerRunRepository{
		run: &model.AISolveRun{
			BaseModel:    model.BaseModel{ID: 23},
			AttemptCount: 3,
			FailureType:  "time_limit",
			StrategyPath: "analysis,repair",
			Attempts: []model.AISolveAttempt{
				{
					BaseModel:        model.BaseModel{ID: 1},
					AttemptNo:        1,
					Stage:            "analysis",
					FailureType:      "wrong_answer",
					RepairReason:     "clarify the invariant",
					StrategyPath:     "analysis",
					PromptPreview:    "first prompt",
					ExtractedCode:    "code1",
					JudgeVerdict:     "WA",
					JudgePassedCount: 1,
					JudgeTotalCount:  3,
					TimedOut:         true,
					ErrorMessage:     "analysis failed",
					TokenInput:       10,
					TokenOutput:      20,
					LLMLatencyMS:     30,
				},
				{
					BaseModel:        model.BaseModel{ID: 2},
					AttemptNo:        2,
					Stage:            "repair",
					FailureType:      "time_limit",
					RepairReason:     "tighten edge cases",
					StrategyPath:     "repair",
					PromptPreview:    "second prompt",
					ExtractedCode:    "code2",
					JudgeVerdict:     "AC",
					JudgePassedCount: 2,
					JudgeTotalCount:  3,
					TimedOut:         true,
					ErrorMessage:     "timeout",
					TokenInput:       11,
					TokenOutput:      21,
					LLMLatencyMS:     31,
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
		t.Fatalf("expected attempts in response, got %+v", got.Attempts)
	}

	first := got.Attempts[0]
	if first.AttemptNo != 1 || first.Stage != "analysis" || first.Verdict != "WA" || first.FailureType != "wrong_answer" || first.RepairReason != "clarify the invariant" {
		t.Fatalf("unexpected first attempt payload: %+v", first)
	}
	if first.StrategyPath != "analysis" || first.PromptPreview != "first prompt" || first.ExtractedCode != "code1" || first.JudgePassedCount != 1 || first.JudgeTotalCount != 3 {
		t.Fatalf("unexpected first attempt metadata: %+v", first)
	}
	if !first.TimedOut || first.ErrorMessage != "analysis failed" || first.TokenInput != 10 || first.TokenOutput != 20 || first.LLMLatencyMS != 30 {
		t.Fatalf("unexpected first attempt timing/cost metadata: %+v", first)
	}

	second := got.Attempts[1]
	if second.AttemptNo != 2 || second.Stage != "repair" || second.Verdict != "AC" || second.FailureType != "time_limit" || second.RepairReason != "tighten edge cases" {
		t.Fatalf("unexpected second attempt payload: %+v", second)
	}
	if second.StrategyPath != "repair" || second.PromptPreview != "second prompt" || second.ExtractedCode != "code2" || second.JudgePassedCount != 2 || second.JudgeTotalCount != 3 {
		t.Fatalf("unexpected second attempt metadata: %+v", second)
	}
	if !second.TimedOut || second.ErrorMessage != "timeout" || second.TokenInput != 11 || second.TokenOutput != 21 || second.LLMLatencyMS != 31 {
		t.Fatalf("unexpected second attempt timing/cost metadata: %+v", second)
	}

	var raw map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal raw response: %v", err)
	}
	attempts, ok := raw["attempts"].([]any)
	if !ok || len(attempts) == 0 {
		t.Fatalf("expected raw attempts array, got %+v", raw["attempts"])
	}
	firstAttempt, ok := attempts[0].(map[string]any)
	if !ok {
		t.Fatalf("expected raw first attempt object, got %+v", attempts[0])
	}
	for _, key := range []string{"raw_response", "compile_stderr", "run_stdout", "run_stderr"} {
		if _, ok := firstAttempt[key]; ok {
			t.Fatalf("unexpected %q exposed in attempt payload: %+v", key, firstAttempt)
		}
	}
}
