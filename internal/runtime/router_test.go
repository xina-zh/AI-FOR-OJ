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

type fakeExperimentRouterRunner struct{}

func (fakeExperimentRouterRunner) Run(context.Context, service.RunExperimentInput) (*service.ExperimentOutput, error) {
	return nil, nil
}

func (fakeExperimentRouterRunner) Get(context.Context, uint) (*service.ExperimentOutput, error) {
	return nil, repository.ErrExperimentNotFound
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
	if len(body.Prompts) == 0 {
		t.Fatal("expected prompt options")
	}
	if len(body.Agents) == 0 {
		t.Fatal("expected agent options")
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
