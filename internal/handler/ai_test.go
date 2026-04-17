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

	var got map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	attemptsValue, ok := got["attempts"].([]any)
	if !ok {
		t.Fatalf("expected attempts array in response, got %+v", got["attempts"])
	}
	if len(attemptsValue) != 2 {
		t.Fatalf("expected attempts in response, got %+v", attemptsValue)
	}

	attempt0, ok := attemptsValue[0].(map[string]any)
	if !ok {
		t.Fatalf("expected first attempt object, got %+v", attemptsValue[0])
	}

	requiredKeys := []string{
		"attempt_no",
		"stage",
		"verdict",
		"failure_type",
		"repair_reason",
		"strategy_path",
		"prompt_preview",
		"extracted_code",
		"judge_passed_count",
		"judge_total_count",
		"timed_out",
		"error_message",
		"token_input",
		"token_output",
		"llm_latency_ms",
	}
	for _, key := range requiredKeys {
		if _, ok := attempt0[key]; !ok {
			t.Fatalf("missing %q in attempt payload: %+v", key, attempt0)
		}
	}

	for _, key := range []string{"raw_response", "compile_stderr", "run_stdout", "run_stderr"} {
		if _, ok := attempt0[key]; ok {
			t.Fatalf("unexpected %q exposed in attempt payload: %+v", key, attempt0)
		}
	}

	expectString := func(payload map[string]any, key, want string) {
		t.Helper()
		gotValue, ok := payload[key].(string)
		if !ok {
			t.Fatalf("expected %q to be a string, got %#v", key, payload[key])
		}
		if gotValue != want {
			t.Fatalf("expected %q to be %q, got %q", key, want, gotValue)
		}
	}
	expectInt := func(payload map[string]any, key string, want int) {
		t.Helper()
		gotValue, ok := payload[key].(float64)
		if !ok {
			t.Fatalf("expected %q to be a number, got %#v", key, payload[key])
		}
		if int(gotValue) != want {
			t.Fatalf("expected %q to be %d, got %v", key, want, gotValue)
		}
	}
	expectBool := func(payload map[string]any, key string, want bool) {
		t.Helper()
		gotValue, ok := payload[key].(bool)
		if !ok {
			t.Fatalf("expected %q to be a bool, got %#v", key, payload[key])
		}
		if gotValue != want {
			t.Fatalf("expected %q to be %v, got %v", key, want, gotValue)
		}
	}

	expectInt(attempt0, "attempt_no", 1)
	expectString(attempt0, "stage", "analysis")
	expectString(attempt0, "verdict", "WA")
	expectString(attempt0, "failure_type", "wrong_answer")
	expectString(attempt0, "repair_reason", "clarify the invariant")
	expectString(attempt0, "strategy_path", "analysis")
	expectString(attempt0, "prompt_preview", "first prompt")
	expectString(attempt0, "extracted_code", "code1")
	expectInt(attempt0, "judge_passed_count", 1)
	expectInt(attempt0, "judge_total_count", 3)
	expectBool(attempt0, "timed_out", true)
	expectString(attempt0, "error_message", "analysis failed")
	expectInt(attempt0, "token_input", 10)
	expectInt(attempt0, "token_output", 20)
	expectInt(attempt0, "llm_latency_ms", 30)

	attempt1, ok := attemptsValue[1].(map[string]any)
	if !ok {
		t.Fatalf("expected second attempt object, got %+v", attemptsValue[1])
	}
	expectInt(attempt1, "attempt_no", 2)
	expectString(attempt1, "stage", "repair")
	expectString(attempt1, "verdict", "AC")
	expectString(attempt1, "failure_type", "time_limit")
	expectString(attempt1, "repair_reason", "tighten edge cases")
	expectString(attempt1, "strategy_path", "repair")
	expectString(attempt1, "prompt_preview", "second prompt")
	expectString(attempt1, "extracted_code", "code2")
	expectInt(attempt1, "judge_passed_count", 2)
	expectInt(attempt1, "judge_total_count", 3)
	expectBool(attempt1, "timed_out", true)
	expectString(attempt1, "error_message", "timeout")
	expectInt(attempt1, "token_input", 11)
	expectInt(attempt1, "token_output", 21)
	expectInt(attempt1, "llm_latency_ms", 31)

	if gotSummaryAttemptCount, ok := got["attempt_count"].(float64); !ok || int(gotSummaryAttemptCount) != 3 {
		t.Fatalf("expected attempt_count=3, got %+v", got["attempt_count"])
	}
	if gotFailureType, ok := got["failure_type"].(string); !ok || gotFailureType != "time_limit" {
		t.Fatalf("expected failure_type=time_limit, got %+v", got["failure_type"])
	}
	if gotStrategyPath, ok := got["strategy_path"].(string); !ok || gotStrategyPath != "analysis,repair" {
		t.Fatalf("expected strategy_path=analysis,repair, got %+v", got["strategy_path"])
	}
}
