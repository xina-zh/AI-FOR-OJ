package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"ai-for-oj/internal/config"
	"ai-for-oj/internal/handler"
	"ai-for-oj/internal/handler/dto"
	"ai-for-oj/internal/model"
	"ai-for-oj/internal/repository"
	"ai-for-oj/internal/service"
)

type fakeDB struct {
	err error
}

func (f fakeDB) PingContext(context.Context) error {
	return f.err
}

type fakeProblemRouterRepository struct {
	problem   *model.Problem
	deletedID uint
}

func (r *fakeProblemRouterRepository) Create(context.Context, *model.Problem) error {
	return nil
}

func (r *fakeProblemRouterRepository) List(context.Context) ([]model.Problem, error) {
	return nil, nil
}

func (r *fakeProblemRouterRepository) GetByID(_ context.Context, id uint) (*model.Problem, error) {
	if r.problem == nil || r.problem.ID != id {
		return nil, repository.ErrProblemNotFound
	}
	return r.problem, nil
}

func (r *fakeProblemRouterRepository) GetByIDWithTestCases(ctx context.Context, id uint) (*model.Problem, error) {
	return r.GetByID(ctx, id)
}

func (r *fakeProblemRouterRepository) Delete(_ context.Context, id uint) error {
	if r.problem == nil || r.problem.ID != id {
		return repository.ErrProblemNotFound
	}
	r.deletedID = id
	return nil
}

type fakeTestCaseRouterRepository struct{}

func (fakeTestCaseRouterRepository) Create(context.Context, *model.TestCase) error {
	return nil
}

func (fakeTestCaseRouterRepository) ListByProblemID(context.Context, uint) ([]model.TestCase, error) {
	return nil, nil
}

type fakeExperimentCompareRouterRepository struct {
	compares []model.ExperimentCompare
}

func (r *fakeExperimentCompareRouterRepository) Create(context.Context, *model.ExperimentCompare) error {
	return nil
}

func (r *fakeExperimentCompareRouterRepository) Update(context.Context, *model.ExperimentCompare) error {
	return nil
}

func (r *fakeExperimentCompareRouterRepository) List(context.Context, repository.ExperimentCompareListQuery) ([]model.ExperimentCompare, int64, error) {
	return r.compares, int64(len(r.compares)), nil
}

func (r *fakeExperimentCompareRouterRepository) GetByID(context.Context, uint) (*model.ExperimentCompare, error) {
	return nil, repository.ErrExperimentCompareNotFound
}

type fakeExperimentRouterRepository struct {
	experiments []model.Experiment
	runTrace    *model.ExperimentRun
}

func (r *fakeExperimentRouterRepository) Create(context.Context, *model.Experiment) error {
	return nil
}

func (r *fakeExperimentRouterRepository) Update(context.Context, *model.Experiment) error {
	return nil
}

func (r *fakeExperimentRouterRepository) CreateRun(context.Context, *model.ExperimentRun) error {
	return nil
}

func (r *fakeExperimentRouterRepository) List(context.Context, repository.ExperimentListQuery) ([]model.Experiment, int64, error) {
	return r.experiments, int64(len(r.experiments)), nil
}

func (r *fakeExperimentRouterRepository) GetByIDWithRuns(context.Context, uint) (*model.Experiment, error) {
	return nil, repository.ErrExperimentNotFound
}

func (r *fakeExperimentRouterRepository) GetRunTrace(_ context.Context, runID uint) (*model.ExperimentRun, error) {
	if r.runTrace == nil || r.runTrace.ID != runID {
		return nil, repository.ErrExperimentRunNotFound
	}
	return r.runTrace, nil
}

type fakeExperimentRepeatRouterRepository struct {
	repeats []model.ExperimentRepeat
}

func (r *fakeExperimentRepeatRouterRepository) Create(context.Context, *model.ExperimentRepeat) error {
	return nil
}

func (r *fakeExperimentRepeatRouterRepository) Update(context.Context, *model.ExperimentRepeat) error {
	return nil
}

func (r *fakeExperimentRepeatRouterRepository) List(context.Context, repository.ExperimentRepeatListQuery) ([]model.ExperimentRepeat, int64, error) {
	return r.repeats, int64(len(r.repeats)), nil
}

func (r *fakeExperimentRepeatRouterRepository) GetByID(context.Context, uint) (*model.ExperimentRepeat, error) {
	return nil, repository.ErrExperimentRepeatNotFound
}

type fakeExperimentRouterRunner struct{}

func (fakeExperimentRouterRunner) Run(context.Context, service.RunExperimentInput) (*service.ExperimentOutput, error) {
	return nil, nil
}

func (fakeExperimentRouterRunner) Get(context.Context, uint) (*service.ExperimentOutput, error) {
	return nil, repository.ErrExperimentNotFound
}

type fakeExperimentRouterAISolver struct{}

func (fakeExperimentRouterAISolver) Solve(context.Context, service.AISolveInput) (*service.AISolveOutput, error) {
	return nil, nil
}

func TestHealthRoute(t *testing.T) {
	cfg := config.Config{
		App: config.AppConfig{
			Name: "ai-for-oj",
			Env:  "test",
		},
	}

	healthService := service.NewHealthService("ai-for-oj", "test", fakeDB{})
	healthHandler := handler.NewHealthHandler(healthService)
	router := NewRouter(cfg, slog.Default(), healthHandler, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, resp.Code)
	}
}

func TestHealthRouteReturns503WhenDatabaseIsDown(t *testing.T) {
	cfg := config.Config{
		App: config.AppConfig{
			Name: "ai-for-oj",
			Env:  "test",
		},
	}

	healthService := service.NewHealthService("ai-for-oj", "test", fakeDB{err: errors.New("db down")})
	healthHandler := handler.NewHealthHandler(healthService)
	router := NewRouter(cfg, slog.Default(), healthHandler, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected %d, got %d", http.StatusServiceUnavailable, resp.Code)
	}
}

func TestProblemDeleteRoute(t *testing.T) {
	cfg := config.Config{App: config.AppConfig{Name: "ai-for-oj", Env: "test"}}
	healthService := service.NewHealthService("ai-for-oj", "test", fakeDB{})
	healthHandler := handler.NewHealthHandler(healthService)
	problemRepo := &fakeProblemRouterRepository{
		problem: &model.Problem{BaseModel: model.BaseModel{ID: 7}, Title: "delete me"},
	}
	problemService := service.NewProblemService(problemRepo, fakeTestCaseRouterRepository{})
	problemHandler := handler.NewProblemHandler(problemService)
	router := NewRouter(cfg, slog.Default(), healthHandler, problemHandler, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/problems/7", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("expected %d, got %d body=%s", http.StatusNoContent, resp.Code, resp.Body.String())
	}
	if problemRepo.deletedID != 7 {
		t.Fatalf("expected deleted id 7, got %d", problemRepo.deletedID)
	}
}

func TestProblemDeleteRouteReturns404(t *testing.T) {
	cfg := config.Config{App: config.AppConfig{Name: "ai-for-oj", Env: "test"}}
	healthService := service.NewHealthService("ai-for-oj", "test", fakeDB{})
	healthHandler := handler.NewHealthHandler(healthService)
	problemService := service.NewProblemService(&fakeProblemRouterRepository{}, fakeTestCaseRouterRepository{})
	problemHandler := handler.NewProblemHandler(problemService)
	router := NewRouter(cfg, slog.Default(), healthHandler, problemHandler, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/problems/404", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d body=%s", http.StatusNotFound, resp.Code, resp.Body.String())
	}
}

func TestExperimentOptionsRoute(t *testing.T) {
	cfg := config.Config{App: config.AppConfig{Name: "ai-for-oj", Env: "test"}}
	healthService := service.NewHealthService("ai-for-oj", "test", fakeDB{})
	healthHandler := handler.NewHealthHandler(healthService)
	metaHandler := handler.NewMetaHandler("test-model")
	router := NewRouter(cfg, slog.Default(), healthHandler, nil, nil, nil, metaHandler, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/meta/experiment-options", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d body=%s", http.StatusOK, resp.Code, resp.Body.String())
	}

	var body dto.ExperimentOptionsResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.DefaultModel != "test-model" {
		t.Fatalf("expected default model test-model, got %q", body.DefaultModel)
	}
	if len(body.Models) == 0 {
		t.Fatal("expected model options")
	}
	if len(body.Prompts) == 0 {
		t.Fatal("expected prompt options")
	}
	if len(body.Agents) == 0 {
		t.Fatal("expected agent options")
	}

	var raw map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decode raw response: %v", err)
	}
	toolingOptions, ok := raw["tooling_options"].([]any)
	if !ok || len(toolingOptions) == 0 {
		t.Fatalf("expected tooling options, got %#v", raw["tooling_options"])
	}
}

func TestExperimentListRoute(t *testing.T) {
	cfg := config.Config{App: config.AppConfig{Name: "ai-for-oj", Env: "test"}}
	healthService := service.NewHealthService("ai-for-oj", "test", fakeDB{})
	healthHandler := handler.NewHealthHandler(healthService)
	experimentService := service.NewExperimentService(
		&fakeExperimentRouterRepository{
			experiments: []model.Experiment{{
				BaseModel:  model.BaseModel{ID: 3},
				Name:       "history-run",
				ModelName:  "mock-model",
				Status:     service.ExperimentStatusCompleted,
				TotalCount: 2,
			}},
		},
		fakeExperimentRouterAISolver{},
		"mock-model",
	)
	experimentHandler := handler.NewExperimentHandler(experimentService, nil, nil)
	router := NewRouter(cfg, slog.Default(), healthHandler, nil, nil, nil, nil, experimentHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/experiments?page=1&page_size=20", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d body=%s", http.StatusOK, resp.Code, resp.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["page"].(float64) != 1 || body["page_size"].(float64) != 20 || body["total"].(float64) != 1 {
		t.Fatalf("unexpected pagination response: %+v", body)
	}
}

func TestCompareListRoute(t *testing.T) {
	cfg := config.Config{App: config.AppConfig{Name: "ai-for-oj", Env: "test"}}
	healthService := service.NewHealthService("ai-for-oj", "test", fakeDB{})
	healthHandler := handler.NewHealthHandler(healthService)
	compareService := service.NewExperimentCompareService(
		&fakeExperimentCompareRouterRepository{},
		fakeExperimentRouterRunner{},
		"mock-model",
	)
	experimentHandler := handler.NewExperimentHandler(nil, compareService, nil)
	router := NewRouter(cfg, slog.Default(), healthHandler, nil, nil, nil, nil, experimentHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/experiments/compare?page=1&page_size=4", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d body=%s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestRepeatListRoute(t *testing.T) {
	cfg := config.Config{App: config.AppConfig{Name: "ai-for-oj", Env: "test"}}
	healthService := service.NewHealthService("ai-for-oj", "test", fakeDB{})
	healthHandler := handler.NewHealthHandler(healthService)
	repeatService := service.NewExperimentRepeatService(
		&fakeExperimentRepeatRouterRepository{
			repeats: []model.ExperimentRepeat{{
				BaseModel:   model.BaseModel{ID: 4},
				Name:        "repeat-history",
				ModelName:   "mock-model",
				ProblemIDs:  "[1,2]",
				RepeatCount: 2,
				Status:      model.ExperimentRepeatStatusCompleted,
			}},
		},
		fakeExperimentRouterRunner{},
		"mock-model",
	)
	experimentHandler := handler.NewExperimentHandler(nil, nil, repeatService)
	router := NewRouter(cfg, slog.Default(), healthHandler, nil, nil, nil, nil, experimentHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/experiments/repeat?page=1&page_size=20", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d body=%s", http.StatusOK, resp.Code, resp.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["page"].(float64) != 1 || body["page_size"].(float64) != 20 || body["total"].(float64) != 1 {
		t.Fatalf("unexpected pagination response: %+v", body)
	}
}

func TestExperimentRunTraceRoute(t *testing.T) {
	cfg := config.Config{App: config.AppConfig{Name: "ai-for-oj", Env: "test"}}
	healthService := service.NewHealthService("ai-for-oj", "test", fakeDB{})
	healthHandler := handler.NewHealthHandler(healthService)
	aiSolveRunID := uint(11)
	submissionID := uint(21)
	experimentService := service.NewExperimentService(
		&fakeExperimentRouterRepository{
			runTrace: &model.ExperimentRun{
				CreatedModel: model.CreatedModel{ID: 7},
				ExperimentID: 5,
				ProblemID:    3,
				AISolveRunID: &aiSolveRunID,
				SubmissionID: &submissionID,
				AttemptNo:    1,
				FinalVerdict: "AC",
				Status:       service.ExperimentRunStatusSuccess,
				AISolveRun:   &model.AISolveRun{BaseModel: model.BaseModel{ID: aiSolveRunID}, Status: model.AISolveRunStatusSuccess, Verdict: "AC"},
				Submission:   &model.Submission{BaseModel: model.BaseModel{ID: submissionID}, ProblemID: 3},
				TraceEvents:  []model.TraceEvent{{CreatedModel: model.CreatedModel{ID: 31}, SequenceNo: 1, StepType: "llm_request", Content: "prompt", Metadata: `{"model":"mock"}`}},
			},
		},
		fakeExperimentRouterAISolver{},
		"mock-model",
	)
	experimentHandler := handler.NewExperimentHandler(experimentService, nil, nil)
	router := NewRouter(cfg, slog.Default(), healthHandler, nil, nil, nil, nil, experimentHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/experiment-runs/7/trace", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d body=%s", http.StatusOK, resp.Code, resp.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["experiment_run_id"].(float64) != 7 {
		t.Fatalf("unexpected run id response: %+v", body)
	}
	timeline, ok := body["timeline"].([]any)
	if !ok || len(timeline) != 1 {
		t.Fatalf("expected one timeline event, got %+v", body["timeline"])
	}
	if body["ai_solve_run"] == nil || body["submission"] == nil {
		t.Fatalf("expected linked solve and submission summaries, got %+v", body)
	}
}
