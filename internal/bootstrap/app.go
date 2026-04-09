package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"ai-for-oj/internal/config"
	"ai-for-oj/internal/handler"
	"ai-for-oj/internal/judge"
	"ai-for-oj/internal/repository"
	"ai-for-oj/internal/runtime"
	"ai-for-oj/internal/sandbox"
	"ai-for-oj/internal/service"
)

type Container struct {
	Config config.Config
	Logger *slog.Logger
	DB     *gorm.DB
	SQLDB  *sql.DB
	Server *runtime.App
}

func Build(configPath string) (*Container, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	logger := NewLogger(cfg.Log)

	db, sqlDB, err := NewDatabase(cfg.Database, logger)
	if err != nil {
		return nil, fmt.Errorf("init database: %w", err)
	}

	if err := RunMigrations(db, cfg.Database, logger); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	healthService := service.NewHealthService(cfg.App.Name, cfg.App.Env, sqlDB)
	healthHandler := handler.NewHealthHandler(healthService)
	problemRepository := repository.NewProblemRepository(db)
	testCaseRepository := repository.NewTestCaseRepository(db)
	submissionRepository := repository.NewSubmissionRepository(db)
	problemService := service.NewProblemService(problemRepository, testCaseRepository)
	problemHandler := handler.NewProblemHandler(problemService)
	sandboxExecutor, err := sandbox.NewDockerSandbox(cfg.Sandbox, logger)
	if err != nil {
		return nil, fmt.Errorf("init docker sandbox: %w", err)
	}
	judgeEngine := judge.NewEngine(sandboxExecutor)
	judgeSubmissionService := service.NewJudgeSubmissionService(problemRepository, submissionRepository, judgeEngine)
	submissionQueryService := service.NewSubmissionQueryService(submissionRepository)
	submissionHandler := handler.NewSubmissionHandler(judgeSubmissionService, submissionQueryService)

	router := runtime.NewRouter(cfg, logger, healthHandler, problemHandler, submissionHandler)
	server := runtime.NewApp(cfg, logger, router, func(context.Context) error {
		return sqlDB.Close()
	})

	return &Container{
		Config: cfg,
		Logger: logger,
		DB:     db,
		SQLDB:  sqlDB,
		Server: server,
	}, nil
}
